package storageapplication

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/application"
	"github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/application/storageApplication/command"
	domainModels "github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/domain/models"
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

func (sa *StorageApplication) GetPresignedPutURL(ctx context.Context, uuid string, objectType string, fileName string, contentType string) (string, error) {
	uuid = strings.TrimSpace(uuid)
	objectType = strings.TrimSpace(objectType)
	fileName = strings.TrimSpace(fileName)
	contentType = strings.TrimSpace(contentType)

	if uuid == "" || objectType == "" || fileName == "" || contentType == "" {
		return "", fmt.Errorf("uuid, objectType, fileName and contentType are required")
	}

	extension := strings.TrimPrefix(contentType, ".")
	if strings.Contains(extension, "/") {
		parts := strings.Split(extension, "/")
		extension = parts[len(parts)-1]
	}
	if extension == "" {
		return "", fmt.Errorf("invalid contentType")
	}

	var objectKey string
	switch objectType {
	case domainModels.USER_AVATAR:
		objectKey = fmt.Sprintf(`profile-pictures/%s/%s/temp/%s-%d.%s`, uuid, domainModels.USER_AVATAR, fileName, time.Now().Unix(), extension)

	case domainModels.USER_BANNER:
		objectKey = fmt.Sprintf(`profile-pictures/%s/%s/temp/%s-%d.%s`, uuid, domainModels.USER_BANNER, fileName, time.Now().Unix(), extension)

	case domainModels.DOCUMENT:
		objectKey = fmt.Sprintf(`documents/%s/%s/temp/%s-%d.%s`, uuid, domainModels.DOCUMENT, fileName, time.Now().Unix(), extension)

	default:
		objectKey = fmt.Sprintf(`others/%s/%d-%s`, uuid, time.Now().Unix(), objectType)
	}
	return sa.storageService.GetPresignedURL(ctx, objectKey, domainModels.STORAGE_OPERATION_PUT)
}
