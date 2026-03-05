package storageapplication

import (
	"context"
	"fmt"
	"strings"

	"github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/application"
	"github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/application/storageApplication/command"
	ports "github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/domain/ports/outbound"
)

var _ application.IStorageApplicationUseCase = (*StorageApplication)(nil)

type StorageApplication struct {
	storageService   ports.IStorageService
	messagePublisher ports.IWorkPublisher
	logger           ports.ILoggerService
}

func NewStorageApplication(
	storageService ports.IStorageService,
	messagePublisher ports.IWorkPublisher,
	logger ports.ILoggerService,
) *StorageApplication {
	return &StorageApplication{
		storageService:   storageService,
		messagePublisher: messagePublisher,
		logger:           logger,
	}
}

func (sa *StorageApplication) UploadFile(ctx context.Context, uploadFormDTO command.UploadFileCommand) error {

	// path dentro del bucket para almacer archivo raw
	// bucketname/multimedia/img/avatar/uuid_filename.ext
	// bucketname/multimedia/img/banner/uuid_filename.ext
	// bucketname/multimedia/video/publicacion/uuid_filename.ext
	// bucketname/multimedia/video/historia/uuid_filename.ext
	// bucketname/documento/tipo_doc/publicacion/uuid_filename.ext

	kingObject := "multimedia"
	typeObject := strings.Split(uploadFormDTO.FileMetadata.ContentType, "/")[0]
	subtypeObject := "avatar"
	fileExtension := strings.Split(uploadFormDTO.FileMetadata.ContentType, "/")[1]

	pathFile := fmt.Sprintf("%s/%s/%s/%s_%s.%s",
		kingObject,
		typeObject,
		subtypeObject,
		uploadFormDTO.FileMetadata.UUID,
		uploadFormDTO.FileMetadata.FileName,
		fileExtension)

	err := sa.storageService.UploadFile(ctx, pathFile, uploadFormDTO.FileData, int64(uploadFormDTO.FileMetadata.FileSize))
	if err != nil {
		sa.logger.Error("Failed to upload file", map[string]interface{}{
			"fileName": uploadFormDTO.FileMetadata.FileName,
			"error":    err.Error(),
		})
		return err
	}

	sa.messagePublisher.EnqueueWork(ctx, "file_uploaded", []byte(uploadFormDTO.FileMetadata.FileName))

	return nil
}

func (sa *StorageApplication) DownloadFile(ctx context.Context, fileName string) ([]byte, error) {
	// Lógica para descargar el archivo desde el servicio de almacenamiento
	return nil, nil
}

func (sa *StorageApplication) DeleteFile(ctx context.Context, fileName string) error {
	// Lógica para eliminar el archivo del servicio de almacenamiento
	return nil
}

func (sa *StorageApplication) ListFiles(ctx context.Context) ([]string, error) {
	// Lógica para listar los archivos disponibles en el servicio de almacenamiento
	return nil, nil
}
