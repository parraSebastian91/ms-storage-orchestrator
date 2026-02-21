package database

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/parraSebastian91/ms-storage-orchestrator.git/src/infrastructure/config"
	"github.com/parraSebastian91/ms-storage-orchestrator.git/src/infrastructure/observability"
)

// ...existing code...

type PostgresClient struct {
	Pool *pgxpool.Pool
}

func NewPostgresClient(cfg config.PostgresConfig, logger *observability.CustomLogger) (*PostgresClient, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	// Opcional: tuning básico
	maxConns, err := strconv.Atoi(cfg.MaxConns)
	if err != nil {
		return nil, fmt.Errorf("invalid maxConns value: %v", err)
	}
	poolConfig.MinConns = 2
	poolConfig.MaxConns = int32(maxConns)
	poolConfig.MaxConnIdleTime = 5 * time.Minute
	poolConfig.MaxConnLifetime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		logger.Error("Error creating Postgres pool", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	// Health check rápido
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		logger.Error("Error pinging Postgres pool", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	logger.Info("Postgres pool initialized successfully", nil)
	return &PostgresClient{Pool: pool}, nil
}

func (c *PostgresClient) Close() {
	if c != nil && c.Pool != nil {
		c.Pool.Close()
	}
}
