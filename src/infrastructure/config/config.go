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

type RoutingKeysConfig struct {
	MediaImageResize       string
	MediaVideoTranscode    string
	MediaDocumentUpload    string
	DteProcessNotification string
}

type ExchangeConfig struct {
	ExchangeTask         string
	ExchangeNotification string
}

type RabbitMQConfig struct {
	URL         string
	Exchange    ExchangeConfig
	RoutingKeys RoutingKeysConfig
}

type StorageBucketConfig struct {
	PublicOriginal   string
	PublicProcessed  string
	PrivateOriginal  string
	PrivateProcessed string
}

type StorageConfig struct {
	Endpoint        string
	PublicEndpoint  string // URL pública visible por el browser (para presigned URLs)
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	Buckets         StorageBucketConfig
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
			URL: getEnvOrDefault("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
			Exchange: ExchangeConfig{
				ExchangeTask:         getEnvOrDefault("RABBITMQ_EXCHANGE_TASK", "storage_tasks_exchange"),
				ExchangeNotification: getEnvOrDefault("RABBITMQ_EXCHANGE_NOTIFICATION", "storage_notifications_exchange"),
			},
			RoutingKeys: RoutingKeysConfig{
				MediaImageResize:       getEnvOrDefault("RABBITMQ_KEY_MEDIA_IMAGE_RESIZE", "media.image.resize"),
				MediaVideoTranscode:    getEnvOrDefault("RABBITMQ_KEY_MEDIA_VIDEO_TRANSCODE", "media.video.transcode"),
				MediaDocumentUpload:    getEnvOrDefault("RABBITMQ_KEY_MEDIA_DOCUMENT_UPLOAD", "media.document.upload"),
				DteProcessNotification: getEnvOrDefault("RABBITMQ_KEY_DTE_PROCESS_NOTIFICATION", "dte.process.notification"),
			},
		},
		Storage: StorageConfig{
			Endpoint:        getEnvOrDefault("STORAGE_ENDPOINT", "localhost:9000"),
			PublicEndpoint:  getEnvOrDefault("STORAGE_PUBLIC_ENDPOINT", ""),
			AccessKeyID:     getEnvOrDefault("STORAGE_ACCESS_KEY", "minioadmin"),
			SecretAccessKey: getEnvOrDefault("STORAGE_SECRET_KEY", "minioadmin"),
			UseSSL:          getEnvOrDefault("STORAGE_USE_SSL", "false") == "true",
			Buckets: StorageBucketConfig{
				PublicOriginal:   getEnvOrDefault("STORAGE_BUCKET_PUBLIC_ORIGINAL", "seis-app-public-original"),
				PublicProcessed:  getEnvOrDefault("STORAGE_BUCKET_PUBLIC_PROCESSED", "seis-app-public-processed"),
				PrivateOriginal:  getEnvOrDefault("STORAGE_BUCKET_PRIVATE_ORIGINAL", "seis-app-private-original"),
				PrivateProcessed: getEnvOrDefault("STORAGE_BUCKET_PRIVATE_PROCESSED", "seis-app-private-processed"),
			},
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
