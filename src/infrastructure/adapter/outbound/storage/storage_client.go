package storage

import (
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/parraSebastian91/ms-storage-orchestrator.git/src/infrastructure/config"
	"github.com/parraSebastian91/ms-storage-orchestrator.git/src/infrastructure/observability"
)

type StorageClient struct {
	minioClient   *minio.Client
	logger        *observability.CustomLogger
	bucketNameRaw string
}

func NewStorageClient(cfg config.StorageConfig, logger *observability.CustomLogger) (*StorageClient, error) {
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
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
	return &StorageClient{
		minioClient:   minioClient,
		logger:        logger,
		bucketNameRaw: cfg.BucketNameRaw,
	}, nil
}

func (c *StorageClient) Close() {
	c.logger.Info("Closing StorageClient resources", nil)
	// No hay recursos específicos que cerrar para el cliente de MinIO, pero si hubiera conexiones o recursos adicionales, se cerrarían aquí.
}
