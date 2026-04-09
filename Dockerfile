# Development stage (hot-reload with go run, source mounted via volume)
FROM golang:1.25-alpine AS development
WORKDIR /app
RUN apk add --no-cache git ca-certificates tzdata
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
EXPOSE 3100
CMD ["go", "run", "./cmd/api"]

# Build stage
FROM golang:1.25-alpine AS builder

# Install git and ca-certificates (needed for fetching dependencies)
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/bin/api ./cmd/api

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S appuser && \
    adduser -u 1001 -S appuser -G appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/api .

# Copy environment files (opcional)
COPY --from=builder /app/.env.dev .env

# Change ownership to non-root user
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 3100

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3100/health || exit 1

# Run the application
CMD ["./api"]
