package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	domainModels "github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/domain/models"
	ports "github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/domain/ports/outbound"
)

var _ ports.IStorageService = (*StorageMinIOServiceImpl)(nil)

type StorageMinIOServiceImpl struct {
	storageClient *StorageClient
	logger        ports.ILoggerService
}

func NewStorageMinIOServiceImpl(storageClient *StorageClient, logger ports.ILoggerService) *StorageMinIOServiceImpl {
	return &StorageMinIOServiceImpl{
		storageClient: storageClient,
		logger:        logger,
	}
}

func (c *StorageMinIOServiceImpl) UploadFile(ctx context.Context, fileName string, fileContent io.Reader, size int64) error {

	if size > 0 {
		c.logger.Info("Uploading file with known size", map[string]interface{}{
			"fileName": fileName,
			"size":     size,
		})
		_, err := c.storageClient.minioClient.PutObject(
			ctx,
			c.storageClient.buckets.PrivateOriginal,
			fileName,
			fileContent,
			size,
			minio.PutObjectOptions{},
		)
		if err != nil {
			c.logger.Error("Failed to upload file with known size", map[string]interface{}{
				"fileName": fileName,
				"size":     size,
				"error":    err.Error(),
			})
			return err
		}
		c.logger.Info("File uploaded successfully with known size", map[string]interface{}{
			"fileName": fileName,
			"size":     size,
		})
		return nil
	}

	data, err := io.ReadAll(fileContent)
	if err != nil {
		c.logger.Error("Failed to read file content", map[string]interface{}{
			"fileName": fileName,
			"error":    err.Error(),
		})
		return err
	}

	_, err = c.storageClient.minioClient.PutObject(
		ctx,
		c.storageClient.buckets.PrivateOriginal,
		fileName,
		io.NopCloser(bytes.NewReader(data)), // reader desde bytes
		int64(len(data)),                    // tamaño conocido
		minio.PutObjectOptions{},
	)

	if err != nil {
		c.logger.Error("Failed to upload file", map[string]interface{}{
			"fileName": fileName,
			"error":    err.Error(),
		})
		return err
	}

	c.logger.Info("File uploaded successfully", map[string]interface{}{
		"fileName": fileName,
		"size":     len(data),
	})
	return nil
}

func (c *StorageMinIOServiceImpl) DownloadFile(ctx context.Context, fileName string) (io.Reader, error) {
	// Implementa la lógica para descargar el archivo desde el almacenamiento local o desde un servicio de almacenamiento en la nube.
	// Por ejemplo, podrías abrir un archivo local:
	// inFile, err := os.Open("/path/to/storage/" + fileName)
	// if err != nil {
	//     return nil, err
	// }
	// return inFile, nil

	fmt.Printf("Downloading file: %s\n", fileName)
	return nil, nil
}

func (c *StorageMinIOServiceImpl) DeleteFile(ctx context.Context, fileName string) error {
	// Implementa la lógica para eliminar el archivo del almacenamiento local o de un servicio de almacenamiento en la nube.
	// Por ejemplo, podrías eliminar un archivo local:
	// return os.Remove("/path/to/storage/" + fileName)

	fmt.Printf("Deleting file: %s\n", fileName)
	return nil
}

func (c *StorageMinIOServiceImpl) ListFiles(ctx context.Context) ([]string, error) {
	// Implementa la lógica para listar los archivos disponibles en el almacenamiento local o en un servicio de almacenamiento en la nube.
	// Por ejemplo, podrías leer los nombres de los archivos en un directorio local:
	// files, err := ioutil.ReadDir("/path/to/storage/")
	// if err != nil {
	//     return nil, err
	// }
	// var fileNames []string
	// for _, file := range files {
	//     if !file.IsDir() {
	//         fileNames = append(fileNames, file.Name())
	//     }
	// }
	// return fileNames, nil

	fmt.Println("Listing files")
	return []string{}, nil
}

func (c *StorageMinIOServiceImpl) GetPresignedURL(ctx context.Context, objectKey string, operation string) (string, error) {
	c.logger.Info("Generating presigned URL", map[string]interface{}{
		"objectKey": objectKey,
		"operation": operation,
	})
	if c == nil || c.storageClient == nil || c.storageClient.presignClient == nil {
		return "", fmt.Errorf("storage client is not initialized")
	}

	if objectKey == "" {
		return "", fmt.Errorf("object key is required")
	}

	switch operation {
	case domainModels.STORAGE_OPERATION_PUT:
		presigned, err := c.storageClient.presignClient.PresignedPutObject(ctx, c.storageClient.buckets.PrivateOriginal, objectKey, time.Duration(5)*time.Minute)
		if err != nil {
			c.logger.Error("Failed to generate presigned PUT URL", map[string]interface{}{
				"bucket":    c.storageClient.buckets.PrivateOriginal,
				"objectKey": objectKey,
				"error":     err.Error(),
			})
			return "", err
		}
		return presigned.String(), nil
	case domainModels.STORAGE_OPERATION_GET:
		presigned, err := c.storageClient.presignClient.PresignedGetObject(ctx, c.storageClient.buckets.PrivateOriginal, objectKey, time.Duration(5)*time.Minute, c.storageClient.presignClient.EndpointURL().Query())
		if err != nil {
			c.logger.Error("Failed to generate presigned GET URL", map[string]interface{}{
				"bucket":    c.storageClient.buckets.PrivateOriginal,
				"objectKey": objectKey,
				"error":     err.Error(),
			})
			return "", err
		}
		return presigned.String(), nil
	default:
		return "", fmt.Errorf("unsupported operation: %s", operation)
	}
}
