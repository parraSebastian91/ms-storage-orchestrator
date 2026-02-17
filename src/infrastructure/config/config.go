package config

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/joho/godotenv"
)

type Config struct {
	Postgres PostgresConfig
	Server   ServerConfig
}

type ServerConfig struct {
	Port        string
	Env         string
	ServiceName string
	MinLogLevel string
	FiberConfig fiber.Config
	CorsConfig  cors.Config
}

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func InitConfig() *Config {
	// Cargar archivo .env (ignorar error en producción)
	if err := godotenv.Load(".env.dev"); err != nil {
		log.Println("Warning: .env.dev file not found, using environment variables")
	}

	return &Config{
		Postgres: PostgresConfig{
			Host:     getEnvOrDefault("POSTGRE_HOST", "localhost"),
			Port:     getEnvOrDefault("POSTGRE_PORT", "5432"),
			User:     getEnvOrDefault("POSTGRE_USER", "admin"),
			Password: getEnvOrDefault("POSTGRE_PASSWORD", "secret"),
			DBName:   getEnvOrDefault("POSTGRE_NAME", "storage_db"),
		},
		Server: ServerConfig{
			Port:        getEnvOrDefault("PORT", "8080"),
			Env:         getEnvOrDefault("ENV", "development"),
			ServiceName: getEnvOrDefault("SERVICE_NAME", "ms-storage-orchestrator"),
			MinLogLevel: getEnvOrDefault("MIN_LOG_LEVEL", "info"),
			FiberConfig: getFiberConfig(getEnvOrDefault("ENV", "development")),
			CorsConfig:  getCorsConfig(getEnvOrDefault("ENV", "development")),
		},
	}
}

func getFiberConfig(env string) fiber.Config {
	switch env {
	case "production":
		return fiber.Config{
			ServerHeader: "Fiber",
			AppName:      getEnvOrDefault("SERVICE_NAME", "ms-storage-orchestrator"),
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		}
	default:
		return fiber.Config{
			ServerHeader: "Fiber",
			AppName:      getEnvOrDefault("SERVICE_NAME", "ms-storage-orchestrator"),
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		}
	}
}

func getCorsConfig(env string) cors.Config {
	switch env {
	case "production":
		return cors.Config{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{"GET,POST,PUT,DELETE,OPTIONS"},
		}
	default:
		return cors.Config{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{"GET,POST,PUT,DELETE,OPTIONS"},
		}
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
