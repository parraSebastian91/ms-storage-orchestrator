package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/parraSebastian91/ms-storage-orchestrator.git/src/infrastructure/adapter/outbound/database"
	"github.com/parraSebastian91/ms-storage-orchestrator.git/src/infrastructure/config"
	"github.com/parraSebastian91/ms-storage-orchestrator.git/src/infrastructure/observability"
)

type AppResources struct {
	logger         *observability.CustomLogger
	postgresClient *database.PostgresClient
}

func main() {

	// ======== INICIALIZACION CONFIGURACION ========

	cfg := config.InitConfig()

	// ======== INICIALIZACION APLICACION ========

	app := InitFiberApp(cfg.Server.FiberConfig, cfg.Server.CorsConfig)

	// ======== INICIALIZACION OBESERVADORES ========

	logger := observability.NewCustomLogger(cfg.Server.ServiceName, cfg.Server.MinLogLevel, cfg.Server.Env == "production")

	// ======== INICIALIZACION RECURSOS ========

	resources, err := InitResources(cfg, logger)
	if err != nil {
		logger.Fatal("Error initializing resources: " + err.Error())
	}

	startServer(app, cfg.Server.Port, resources)
}

func InitFiberApp(config fiber.Config, corsConfig cors.Config) *fiber.App {
	app := fiber.New(config)
	app.Use(cors.New(corsConfig))
	return app
}

func InitResources(cfg *config.Config, logger *observability.CustomLogger) (*AppResources, error) {

	postgresClient, err := database.NewPostgresClient(cfg.Postgres, logger)
	if err != nil {
		return nil, err
	}

	return &AppResources{
		logger:         logger,
		postgresClient: postgresClient,
	}, nil
}

func startServer(app *fiber.App, port string, resources *AppResources) {

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		resources.logger.Info("Starting server", map[string]interface{}{
			"port": port,
		})

		if err := app.Listen(":" + port); err != nil {
			resources.logger.Fatal("Error starting server", map[string]interface{}{
				"error": err.Error(),
			})
			quit <- syscall.SIGTERM
		}
	}()

	sig := <-quit
	resources.logger.Info("Shutting down server", map[string]interface{}{
		"signal": sig.String(),
	})
	gracefulShutdown(app, resources)

}

func gracefulShutdown(app *fiber.App, resources *AppResources) {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resources.logger.Info("Initiating graceful shutdown")

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		resources.logger.Error("Forced shutdown due to error", map[string]interface{}{
			"error": err.Error(),
		})
	}

	resources.logger.Info("Closing resources")

	closeResources := func(name string, closeFunc func() error) {
		resources.logger.Info("Closing resource", map[string]interface{}{
			"resource": name,
		})
		if err := closeFunc(); err != nil {
			resources.logger.Error("Error closing resource", map[string]interface{}{
				"resource": name,
				"error":    err.Error(),
			})
		} else {
			resources.logger.Info("Resource closed successfully", map[string]interface{}{
				"resource": name,
			})
		}
	}

	closeResources("PostgresClient", func() error {
		resources.postgresClient.Close()
		return nil
	})

	resources.logger.Info("Shutdown complete")

}
