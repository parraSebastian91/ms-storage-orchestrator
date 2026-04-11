package storageUseCase

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	AplicationModel "github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/application/model"
	"github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/application/useCase/storageApplication/command"
	domainModels "github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/domain/models"
	inbound "github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/domain/ports/inbound"
	outbound "github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/domain/ports/outbound"
)

var _ inbound.IStorageUseCase = (*StorageUsecase)(nil)

type StorageUsecase struct {
	storageService   outbound.IStorageService
	messagePublisher outbound.IMessagePublisher
	mediaRepository  outbound.IMediaRepository
	logger           outbound.ILoggerService
}

func NewStorageApplication(
	storageService outbound.IStorageService,
	messagePublisher outbound.IMessagePublisher,
	mediaRepository outbound.IMediaRepository,
	logger outbound.ILoggerService,
) *StorageUsecase {
	return &StorageUsecase{
		storageService:   storageService,
		messagePublisher: messagePublisher,
		mediaRepository:  mediaRepository,
		logger:           logger,
	}
}

func (sa *StorageUsecase) UploadFile(ctx context.Context, uploadFormDTO command.UploadFileCommand) error {

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

	// sa.messagePublisher.EnqueueWork(ctx, "file_uploaded", []byte(uploadFormDTO.FileMetadata.FileName))

	return nil
}

func (sa *StorageUsecase) ExecuteProcessFile(ctx context.Context, objectKey string) error {

	mediaModel, err := sa.mediaRepository.GetMediaMetadata(ctx, objectKey)
	if err != nil {
		sa.logger.Error("Failed to get media metadata", map[string]interface{}{
			"objectKey": objectKey,
			"error":     err.Error(),
		})
		return err
	}

	switch mediaModel.CategoryProcess {
	case domainModels.CATEGORY_PROCESS_USER_AVATAR:
		err = sa.messagePublisher.PublishTypeImage(ctx, mediaModel)
	case domainModels.CATEGORY_PROCESS_USER_BANNER:
		err = sa.messagePublisher.PublishTypeImage(ctx, mediaModel)
	case domainModels.CATEGORY_PROCESS_DOCUMENT:
		err = sa.messagePublisher.PublishTypeDocument(ctx, mediaModel)
	default:
		err = sa.messagePublisher.PublishTypeArchive(ctx, mediaModel)
	}

	if err != nil {
		sa.logger.Error("Failed to publish message for file processing", map[string]interface{}{
			"objectKey": objectKey,
			"error":     err.Error(),
		})
		return err
	}

	return nil
}

func (sa *StorageUsecase) ExecuteGetPresignedPutURL(ctx context.Context, command command.GetPresignedPutURLCommand) (string, error) {
	uuidUser := strings.TrimSpace(command.UUID)
	objectType := strings.TrimSpace(command.ObjectType)
	re := regexp.MustCompile(`\\s+`)
	fileName := strings.TrimSpace(re.ReplaceAllString(command.FileName, "_"))
	contentType := strings.TrimSpace(command.ContentType)

	if uuidUser == "" || objectType == "" || fileName == "" || contentType == "" {
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
	var category string
	var mediaType string
	var uuidFile = uuid.New()
	switch objectType {
	case domainModels.CATEGORY_PROCESS_USER_AVATAR:
		objectKey = fmt.Sprintf(`profile-pictures/%s/%s/temp/%s-%s.%s`, uuidUser, domainModels.CATEGORY_PROCESS_USER_AVATAR, fileName, uuidFile, extension)
		category = domainModels.CATEGORY_PROCESS_USER_AVATAR
		mediaType = domainModels.MEDIA_TYPE_IMAGE
	case domainModels.CATEGORY_PROCESS_USER_BANNER:
		objectKey = fmt.Sprintf(`profile-pictures/%s/%s/temp/%s-%s.%s`, uuidUser, domainModels.CATEGORY_PROCESS_USER_BANNER, fileName, uuidFile, extension)
		category = domainModels.CATEGORY_PROCESS_USER_BANNER
		mediaType = domainModels.MEDIA_TYPE_IMAGE
	case domainModels.CATEGORY_PROCESS_DOCUMENT:
		objectKey = fmt.Sprintf(`documents/%s/%s/temp/%s-%s.%s`, uuidUser, domainModels.CATEGORY_PROCESS_DOCUMENT, fileName, uuidFile, extension)
		category = domainModels.CATEGORY_PROCESS_DOCUMENT
		mediaType = domainModels.MEDIA_TYPE_DOCUMENT
	default:
		objectKey = fmt.Sprintf(`others/%s/%s-%s`, uuidUser, uuidFile, objectType)
		category = "others"
		mediaType = domainModels.MEDIA_TYPE_ARCHIVE
	}

	err := sa.mediaRepository.CreateMediaMetadata(ctx, AplicationModel.StorageModel{
		OwnerUUID:       uuidUser,
		MediaType:       mediaType,
		CategoryProcess: category,
		NameFile:        fileName,
		FormatFile:      contentType,
		StorageKey:      objectKey,
	})
	if err != nil {
		sa.logger.Error("Failed to persist media metadata", map[string]interface{}{
			"error":      err.Error(),
			"ownerId":    uuidUser,
			"mediaType":  mediaType,
			"category":   category,
			"storageKey": objectKey,
		})
		return "", err
	}

	return sa.storageService.GetPresignedURL(ctx, objectKey, domainModels.STORAGE_OPERATION_PUT)
}
