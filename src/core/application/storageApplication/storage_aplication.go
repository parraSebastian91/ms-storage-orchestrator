package storageapplication

import (
	"context"
	"io"

	"github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/application"
	ports "github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/domain/ports/outbound"
)

var _ application.IStorageApplicationUseCase = (*StorageApplication)(nil)

type StorageApplication struct {
	storageService   ports.IStorageService
	messagePublisher ports.IMessagePublisher
	logger           ports.ILoggerService
}

func NewStorageApplication(
	storageService ports.IStorageService,
	messagePublisher ports.IMessagePublisher,
	logger ports.ILoggerService,
) *StorageApplication {
	return &StorageApplication{
		storageService:   storageService,
		messagePublisher: messagePublisher,
		logger:           logger,
	}
}

func (a *StorageApplication) UploadFile(ctx context.Context, fileData io.Reader, fileName string, size int64) error {
	err := a.storageService.UploadFile(ctx, fileName, fileData, size)
	if err != nil {
		a.logger.Error("Failed to upload file", map[string]interface{}{
			"fileName": fileName,
			"error":    err.Error(),
		})
		return err
	}

	return nil
}

func (a *StorageApplication) DownloadFile(ctx context.Context, fileName string) ([]byte, error) {
	// Lógica para descargar el archivo desde el servicio de almacenamiento
	return nil, nil
}

func (a *StorageApplication) DeleteFile(ctx context.Context, fileName string) error {
	// Lógica para eliminar el archivo del servicio de almacenamiento
	return nil
}

func (a *StorageApplication) ListFiles(ctx context.Context) ([]string, error) {
	// Lógica para listar los archivos disponibles en el servicio de almacenamiento
	return nil, nil
}
