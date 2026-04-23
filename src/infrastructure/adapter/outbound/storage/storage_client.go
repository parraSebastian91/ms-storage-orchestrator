package storage

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/parraSebastian91/ms-storage-orchestrator.git/src/infrastructure/config"
	"github.com/parraSebastian91/ms-storage-orchestrator.git/src/infrastructure/observability"
)

type StorageClient struct {
	minioClient    *minio.Client
	presignClient  *minio.Client
	logger         *observability.CustomLogger
	bucketNameRaw  string
	publicEndpoint string // host:port o scheme://host:port visible por el browser
}

const defaultMinioRegion = "us-east-1"

func parseEndpointAndSecure(raw string, defaultSecure bool) (string, bool, error) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return "", defaultSecure, fmt.Errorf("empty endpoint")
	}

	if !strings.Contains(v, "://") {
		return v, defaultSecure, nil
	}

	u, err := url.Parse(v)
	if err != nil {
		return "", defaultSecure, err
	}
	if u.Host == "" {
		return "", defaultSecure, fmt.Errorf("invalid endpoint: missing host")
	}

	secure := defaultSecure
	switch strings.ToLower(u.Scheme) {
	case "https":
		secure = true
	case "http":
		secure = false
	}

	return u.Host, secure, nil
}

func NewStorageClient(cfg config.StorageConfig, logger *observability.CustomLogger) (*StorageClient, error) {
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
		Region: defaultMinioRegion,
	})
	if err != nil {
		logger.Error("Error creating MinIO client", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}
	logger.Info("Storage client created successfully", map[string]interface{}{
		"endpoint": cfg.Endpoint,
	})

	presignClient := minioClient
	if strings.TrimSpace(cfg.PublicEndpoint) != "" {
		publicHost, publicSecure, parseErr := parseEndpointAndSecure(cfg.PublicEndpoint, cfg.UseSSL)
		if parseErr != nil {
			logger.Warn("Invalid STORAGE_PUBLIC_ENDPOINT, using internal endpoint for presign", map[string]interface{}{
				"publicEndpoint": cfg.PublicEndpoint,
				"error":          parseErr.Error(),
			})
		} else {
			publicClient, publicErr := minio.New(publicHost, &minio.Options{
				Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
				Secure: publicSecure,
				Region: defaultMinioRegion,
			})
			if publicErr != nil {
				logger.Warn("Failed to create MinIO presign client with public endpoint, using internal endpoint", map[string]interface{}{
					"publicEndpoint": cfg.PublicEndpoint,
					"error":          publicErr.Error(),
				})
			} else {
				presignClient = publicClient
				logger.Info("MinIO presign client configured with public endpoint", map[string]interface{}{
					"publicEndpoint": cfg.PublicEndpoint,
				})
			}
		}
	}

	return &StorageClient{
		minioClient:    minioClient,
		presignClient:  presignClient,
		logger:         logger,
		bucketNameRaw:  cfg.BucketNameRaw,
		publicEndpoint: cfg.PublicEndpoint,
	}, nil
}

func (c *StorageClient) Close() {
	c.logger.Info("Closing StorageClient resources", nil)
	// No hay recursos específicos que cerrar para el cliente de MinIO, pero si hubiera conexiones o recursos adicionales, se cerrarían aquí.
}
