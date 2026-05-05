package storageUseCase

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

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

var keyUnsafeCharsRegex = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)
var fileNameUnsafeCharsRegex = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)
var repeatedUnderscoreRegex = regexp.MustCompile(`_+`)

func sanitizeObjectKeySegment(value string) string {
	v := strings.TrimSpace(value)
	v = strings.ReplaceAll(v, " ", "_")
	v = keyUnsafeCharsRegex.ReplaceAllString(v, "_")
	v = repeatedUnderscoreRegex.ReplaceAllString(v, "_")
	v = strings.Trim(v, "._-")
	if v == "" {
		return "file"
	}
	return v
}

func sanitizeFileNameForStorageAndDB(value string) string {
	v := strings.ToValidUTF8(value, "")
	v = strings.TrimSpace(v)
	v = strings.ReplaceAll(v, " ", "_")
	v = fileNameUnsafeCharsRegex.ReplaceAllString(v, "_")
	v = repeatedUnderscoreRegex.ReplaceAllString(v, "_")
	v = strings.Trim(v, "_-")
	if v == "" {
		return "file"
	}
	if len(v) > 120 {
		v = v[:120]
	}
	return v
}

func sanitizeObjectKeyExtension(value string) string {
	v := strings.TrimSpace(strings.TrimPrefix(value, "."))
	if strings.Contains(v, "/") {
		parts := strings.Split(v, "/")
		v = parts[len(parts)-1]
	}
	v = sanitizeObjectKeySegment(v)
	v = strings.ToLower(v)
	if v == "file" {
		return "bin"
	}
	return v
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

	sa.logger.Info("ExecuteProcessFile started", map[string]interface{}{
		"objectKey": objectKey,
	})
	mediaModel, err := sa.mediaRepository.GetMediaMetadata(ctx, objectKey)
	if err != nil {
		sa.logger.Error("Failed to get media metadata", map[string]interface{}{
			"objectKey": objectKey,
			"error":     err.Error(),
		})
		return err
	}
	sa.logger.Info("Media metadata retrieved successfully", map[string]interface{}{
		"correlationId": mediaModel.CorrelationId,
	})

	switch mediaModel.CategoryProcess {
	case domainModels.CATEGORY_PROCESS_USER_AVATAR:
		err = sa.messagePublisher.PublishTypeImage(ctx, mediaModel)
	case domainModels.CATEGORY_PROCESS_USER_BANNER:
		err = sa.messagePublisher.PublishTypeImage(ctx, mediaModel)
	case domainModels.CATEGORY_PROCESS_DOCUMENT_DTO:
		err = sa.messagePublisher.PublishTypeDocument(ctx, mediaModel)
	default:
		err = sa.messagePublisher.PublishTypeArchive(ctx, mediaModel)
	}

	if err != nil {
		sa.logger.Error("Failed to publish message for file processing", map[string]interface{}{
			"correlationId": mediaModel.CorrelationId,
			"error":         err.Error(),
		})
		return err
	}

	return nil
}

func (sa *StorageUsecase) ExecuteNotifyProcessObject(ctx context.Context, notifyModel AplicationModel.NotifyModel) error {
	start := time.Now()
	sa.logger.Info("ExecuteNotifyProcessObject started", map[string]interface{}{
		"CategoryProcess": notifyModel.Category,
		"correlationId":   notifyModel.CorrelationId,
	})

	switch notifyModel.Category {
	case domainModels.CATEGORY_PROCESS_DOCUMENT_DTO:
		err := sa.messagePublisher.PublishDteProcessNotification(ctx, notifyModel)
		if err != nil {
			sa.logger.Error("Failed to publish DTE process notification", map[string]interface{}{
				"correlationId": notifyModel.CorrelationId,
				"error":         err.Error(),
			})
			return err
		}
	default:
		sa.logger.Warn("No specific notification handler for category, skipping", map[string]interface{}{
			"CategoryProcess": notifyModel.Category,
			"correlationId":   notifyModel.CorrelationId,
		})
	}

	sa.logger.Info("ExecuteNotifyProcessObject finished", map[string]interface{}{
		"CategoryProcess": notifyModel.Category,
		"correlationId":   notifyModel.CorrelationId,
		"durationMs":      time.Since(start).Milliseconds(),
	})
	return nil
}

func (sa *StorageUsecase) ExecuteGetPresignedPutURL(ctx context.Context, command command.GetPresignedPutURLCommand) (string, error) {
	start := time.Now()
	uuidUser := strings.TrimSpace(command.UUID)
	objectType := strings.TrimSpace(command.ObjectType)
	fileName := sanitizeFileNameForStorageAndDB(command.FileName)
	contentType := strings.TrimSpace(command.ContentType)
	correlationId := strings.TrimSpace(command.CorrelationId)

	sa.logger.Info("ExecuteGetPresignedPutURL started", map[string]interface{}{
		"ownerId":       uuidUser,
		"objectType":    objectType,
		"contentType":   contentType,
		"correlationId": correlationId,
	})

	if uuidUser == "" || objectType == "" || contentType == "" {
		sa.logger.Warn("ExecuteGetPresignedPutURL validation failed", map[string]interface{}{
			"ownerId":       uuidUser,
			"objectType":    objectType,
			"contentType":   contentType,
			"correlationId": correlationId,
		})
		return "", fmt.Errorf("uuid, objectType, fileName and contentType are required")
	}

	extension := sanitizeObjectKeyExtension(contentType)
	if extension == "" {
		return "", fmt.Errorf("invalid contentType")
	}

	safeObjectType := sanitizeObjectKeySegment(objectType)

	var objectKey string
	var category string
	var mediaType string
	var uuidFile = uuid.New()
	switch objectType {
	case domainModels.CATEGORY_PROCESS_USER_AVATAR:
		objectKey = fmt.Sprintf(`private/profile-pictures/%s/%s/temp/%s-%s.%s`, uuidUser, domainModels.CATEGORY_PROCESS_USER_AVATAR, fileName, uuidFile, extension)
		category = domainModels.CATEGORY_PROCESS_USER_AVATAR
		mediaType = domainModels.MEDIA_TYPE_IMAGE
	case domainModels.CATEGORY_PROCESS_USER_BANNER:
		objectKey = fmt.Sprintf(`private/profile-pictures/%s/%s/temp/%s-%s.%s`, uuidUser, domainModels.CATEGORY_PROCESS_USER_BANNER, fileName, uuidFile, extension)
		category = domainModels.CATEGORY_PROCESS_USER_BANNER
		mediaType = domainModels.MEDIA_TYPE_IMAGE
	case domainModels.CATEGORY_PROCESS_DOCUMENT_DTO:
		objectKey = fmt.Sprintf(`private/documents/%s/%s/%s-%s.%s`, uuidUser, domainModels.CATEGORY_PROCESS_DOCUMENT_DTO, fileName, uuidFile, extension)
		category = domainModels.CATEGORY_PROCESS_DOCUMENT_DTO
		mediaType = domainModels.MEDIA_TYPE_DOCUMENT
	default:
		objectKey = fmt.Sprintf(`private/others/%s/%s-%s`, uuidUser, uuidFile, safeObjectType)
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
		CorrelationId:   correlationId,
	})
	if err != nil {
		sa.logger.Error("Failed to persist media metadata", map[string]interface{}{
			"error":         err.Error(),
			"ownerId":       uuidUser,
			"mediaType":     mediaType,
			"category":      category,
			"storageKey":    objectKey,
			"correlationId": correlationId,
			"durationMs":    time.Since(start).Milliseconds(),
		})
		return "", err
	}

	sa.logger.Info("Media metadata persisted", map[string]interface{}{
		"ownerId":       uuidUser,
		"mediaType":     mediaType,
		"category":      category,
		"storageKey":    objectKey,
		"correlationId": correlationId,
	})

	presignedURL, err := sa.storageService.GetPresignedURL(ctx, objectKey, domainModels.STORAGE_OPERATION_PUT)
	if err != nil {
		sa.logger.Error("Failed to generate presigned URL", map[string]interface{}{
			"error":         err.Error(),
			"storageKey":    objectKey,
			"correlationId": correlationId,
			"durationMs":    time.Since(start).Milliseconds(),
		})
		return "", err
	}

	sa.logger.Info("ExecuteGetPresignedPutURL finished", map[string]interface{}{
		"ownerId":       uuidUser,
		"storageKey":    objectKey,
		"correlationId": correlationId,
		"durationMs":    time.Since(start).Milliseconds(),
	})

	return presignedURL, nil
}
