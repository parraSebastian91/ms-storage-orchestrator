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
	RabbitMQ RabbitMQConfig
	Storage  StorageConfig
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
	MaxConns string
}

type RabbitMQConfig struct {
	URL      string
	Exchange string
	Queue    string
}

type StorageConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	BucketNameRaw   string
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
			MaxConns: getEnvOrDefault("POSTGRE_MAX_CONNS", "10"),
		},
		Server: ServerConfig{
			Port:        getEnvOrDefault("PORT", "8080"),
			Env:         getEnvOrDefault("ENV", "development"),
			ServiceName: getEnvOrDefault("SERVICE_NAME", "ms-storage-orchestrator"),
			MinLogLevel: getEnvOrDefault("MIN_LOG_LEVEL", "info"),
			FiberConfig: getFiberConfig(getEnvOrDefault("ENV", "development")),
			CorsConfig:  getCorsConfig(getEnvOrDefault("ENV", "development")),
		},
		RabbitMQ: RabbitMQConfig{
			URL:      getEnvOrDefault("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
			Exchange: getEnvOrDefault("RABBITMQ_EXCHANGE", "storage_task_exchange"),
			Queue:    getEnvOrDefault("RABBITMQ_QUEUE", "storage_tasks"),
		},
		Storage: StorageConfig{
			Endpoint:        getEnvOrDefault("STORAGE_ENDPOINT", "localhost:9000"),
			AccessKeyID:     getEnvOrDefault("STORAGE_ACCESS_KEY", "minioadmin"),
			SecretAccessKey: getEnvOrDefault("STORAGE_SECRET_KEY", "minioadmin"),
			UseSSL:          getEnvOrDefault("STORAGE_USE_SSL", "false") == "true",
			BucketNameRaw:   getEnvOrDefault("STORAGE_BUCKET_NAME_RAW", "storage-bucket-raw"),
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
