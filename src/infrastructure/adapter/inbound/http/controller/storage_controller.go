package controller

import (
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	AplicationModel "github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/application/model"
	"github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/application/useCase/storageApplication/command"
	inbound "github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/domain/ports/inbound"
	outbound "github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/domain/ports/outbound"
	inbound_dto "github.com/parraSebastian91/ms-storage-orchestrator.git/src/infrastructure/adapter/inbound/http/dto"
)

type StorageController struct {
	storageApplication inbound.IStorageUseCase
	logger             outbound.ILoggerService
}

// NewStorageController crea una nueva instancia de StorageController
func NewStorageController(storageApplication inbound.IStorageUseCase, logger outbound.ILoggerService) *StorageController {
	return &StorageController{
		storageApplication: storageApplication,
		logger:             logger,
	}
}
func (uc *StorageController) UploadFile(ctx fiber.Ctx) error {
	var UploadFormDTO inbound_dto.UploadFormDTO

	if errUuid := ctx.Bind().Query(&UploadFormDTO); errUuid != nil {
		uc.logger.Error("Error al bindear el formulario", map[string]interface{}{
			"error": errUuid.Error(),
		})
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Faltan Datos en la Solicitud",
		})
	}

	if err := ctx.Bind().Form(&UploadFormDTO); err != nil {
		uc.logger.Error("Formulario inválido", map[string]interface{}{
			"error": err.Error(),
		})
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Faltan Datos en la Solicitud",
		})
	}

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		uc.logger.Error("No se pudo obtener el archivo", map[string]interface{}{
			"error": err.Error(),
		})
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No se pudo obtener el archivo",
		})
	}

	file, err := fileHeader.Open()
	if err != nil {
		uc.logger.Error("No se pudo abrir el archivo", map[string]interface{}{
			"error": err.Error(),
		})
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "No se pudo procesar el archivo",
		})
	}
	defer file.Close()

	// Detectar content type
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	fileCommand := command.UploadFileCommand{
		FileData: file,
		FileMetadata: command.FileMetadata{
			UUID:          UploadFormDTO.UUID,
			FileName:      UploadFormDTO.FileName,
			FileSize:      int(UploadFormDTO.FileSize),
			ContentType:   contentType,
			Tags:          UploadFormDTO.Tags,
			Description:   UploadFormDTO.Description,
			RetentionDays: UploadFormDTO.RetentionDays,
			IsPublic:      UploadFormDTO.IsPublic,
		},
	}

	err = uc.storageApplication.UploadFile(
		ctx.Context(),
		fileCommand,
	)

	if err != nil {
		uc.logger.Error("Upload failed", map[string]interface{}{
			"error":    err.Error(),
			"fileName": fileHeader.Filename,
			"uuid":     UploadFormDTO.UUID,
		})
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error al subir el archivo",
		})
	}

	uc.logger.Info("File uploaded successfully", map[string]interface{}{
		"fileName": fileHeader.Filename,
		"size":     fileHeader.Size,
		"uuid":     UploadFormDTO.UUID,
	})

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":  "Archivo subido exitosamente",
		"fileName": fileHeader.Filename,
		"size":     fileHeader.Size,
	})
}

func (c *StorageController) DownloadFile() {
	// Lógica para manejar la descarga de archivos
}

func (c *StorageController) DeleteFile() {
	// Lógica para manejar la eliminación de archivos
}

func (c *StorageController) ListFiles() {
	// Lógica para manejar la lista de archivos
}

func (c *StorageController) GetPresignedURL(ctx fiber.Ctx) error {
	start := time.Now()
	var presignedURLRequest inbound_dto.PresignedURLRequestDTO
	if err := ctx.Bind().Query(&presignedURLRequest); err != nil {
		c.logger.Error("Error al parsear la solicitud de URL prefirmada", map[string]interface{}{
			"error": err.Error(),
			"path":  ctx.Path(),
		})
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Faltan Datos en la Solicitud",
		})
	}

	correlationId := strings.TrimSpace(ctx.Get("X-Correlation-Id"))
	if correlationId == "" {
		correlationId = "N/A"
	}

	c.logger.Info("Presigned URL request received", map[string]interface{}{
		"correlationId": correlationId,
		"path":          ctx.Path(),
		"method":        ctx.Method(),
	})

	presignedURLRequest.UUID = strings.TrimSpace(presignedURLRequest.UUID)
	presignedURLRequest.ObjectType = strings.TrimSpace(presignedURLRequest.ObjectType)
	presignedURLRequest.FileName = strings.TrimSpace(presignedURLRequest.FileName)
	presignedURLRequest.ContentType = strings.TrimSpace(presignedURLRequest.ContentType)

	if presignedURLRequest.UUID == "" || presignedURLRequest.ObjectType == "" || presignedURLRequest.FileName == "" || presignedURLRequest.ContentType == "" {
		c.logger.Warn("Presigned URL request validation failed", map[string]interface{}{
			"correlationId": correlationId,
			"uuid":          presignedURLRequest.UUID,
			"objectType":    presignedURLRequest.ObjectType,
			"contentType":   presignedURLRequest.ContentType,
		})
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "uuid, object_type, file_name y content_type son requeridos",
		})
	}

	url, err := c.storageApplication.ExecuteGetPresignedPutURL(
		ctx.Context(),
		command.GetPresignedPutURLCommand{
			UUID:          presignedURLRequest.UUID,
			ObjectType:    presignedURLRequest.ObjectType,
			FileName:      presignedURLRequest.FileName,
			ContentType:   presignedURLRequest.ContentType,
			CorrelationId: correlationId,
		},
	)

	if err != nil {
		c.logger.Error("Error al obtener la URL prefirmada", map[string]interface{}{
			"error":         err.Error(),
			"correlationId": correlationId,
			"durationMs":    time.Since(start).Milliseconds(),
		})
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "No se pudo generar la URL prefirmada",
		})
	}
	c.logger.Info("Presigned URL generated successfully", map[string]interface{}{
		"uuid":            presignedURLRequest.UUID,
		"objectType":      presignedURLRequest.ObjectType,
		"fileName":        presignedURLRequest.FileName,
		"correlationId":   correlationId,
		"durationMs":      time.Since(start).Milliseconds(),
		"presignedUrlSet": url != "",
	})
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"url": url,
	})

}

func (c *StorageController) MinioWebhookHandler(ctx fiber.Ctx) error {
	c.logger.Info("Received MinIO webhook event", map[string]interface{}{
		"body": string(ctx.Body()),
	})
	var event inbound_dto.MinIOEvent
	if err := ctx.Bind().Body(&event); err != nil {
		c.logger.Error("Error al parsear el evento de MinIO", map[string]interface{}{
			"error": err.Error(),
		})
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	c.logger.Info("Received MinIO webhook event", map[string]interface{}{
		"event": event,
	})
	if len(event.Records) == 0 {
		c.logger.Warn("MinIO webhook recibido sin records", map[string]interface{}{})
		return ctx.Status(fiber.StatusAccepted).JSON(fiber.Map{
			"message": "Webhook recibido sin records",
		})
	}

	processed := 0
	failed := make([]string, 0)

	for _, record := range event.Records {
		objectKey := strings.TrimSpace(record.S3.Object.Key)
		if objectKey == "" {
			c.logger.Warn("Record MinIO sin object key", map[string]interface{}{
				"bucket": record.S3.Bucket.Name,
			})
			continue
		}

		decodedKey, decodeErr := url.QueryUnescape(objectKey)
		if decodeErr == nil {
			objectKey = decodedKey
		}

		directory := extractTopLevelDirectory(objectKey)
		if err := c.storageApplication.ExecuteProcessFile(ctx.Context(), objectKey); err != nil {
			failed = append(failed, objectKey)
			c.logger.Error("Error al procesar el archivo desde el webhook de MinIO", map[string]interface{}{
				"error":     err.Error(),
				"objectKey": objectKey,
				"directory": directory,
				"bucket":    record.S3.Bucket.Name,
			})
			continue
		}

		processed++
		c.logger.Info("Archivo encolado desde webhook MinIO", map[string]interface{}{
			"objectKey": objectKey,
			"directory": directory,
			"bucket":    record.S3.Bucket.Name,
		})
	}

	if len(failed) > 0 {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message":         "Webhook procesado con errores parciales",
			"processed_count": processed,
			"failed_count":    len(failed),
			"failed_objects":  failed,
		})
	}

	// Extraer datos y publicar en RabbitMQ
	// for _, record := range event.Records {
	// bucket := record.S3.Bucket.Name
	// key := record.S3.Object.Key

	// Aquí tu lógica de Go:
	// 1. Identificar el AssetID desde el Key (o metadata)
	// 2. Enviar mensaje a RabbitMQ para que Rust trabaje

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":         "Webhook recibido correctamente",
		"processed_count": processed,
	})
}

func extractTopLevelDirectory(objectKey string) string {
	key := strings.Trim(objectKey, "/")
	if key == "" {
		return "root"
	}

	parts := strings.Split(key, "/")
	if len(parts) == 1 {
		return "root"
	}

	return parts[0]
}

func (c *StorageController) NotifyFileProcessedHandler(ctx fiber.Ctx) error {
	c.logger.Info("Received file processed notification", map[string]interface{}{
		"body": string(ctx.Body()),
	})
	var notification inbound_dto.NotifyProcessDTO
	if err := ctx.Bind().Body(&notification); err != nil {
		c.logger.Error("Error al parsear la notificación de archivo procesado", map[string]interface{}{
			"error": err.Error(),
		})
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	c.logger.Info("Received file processed notification", map[string]interface{}{
		"notification": notification,
	})

	notifyModel := AplicationModel.NotifyModel{
		Category:      notification.Category,
		Status:        notification.Status,
		Timestamp:     notification.Timestamp,
		CorrelationId: notification.CorrelationId,
		App:           notification.App,
		Payload: AplicationModel.NotifyPayload{
			NumeroFactura: notification.Payload.NumeroFactura,
			RutDeudor:     notification.Payload.RutDeudor,
			NombreDeudor:  notification.Payload.NombreDeudor,
			MontoTotal:    notification.Payload.MontoTotal,
		},
	}

	if err := c.storageApplication.ExecuteNotifyProcessObject(ctx.Context(), notifyModel); err != nil {
		c.logger.Error("Error al procesar la notificación de archivo procesado", map[string]interface{}{
			"error": err.Error(),
		})
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Notificación recibida correctamente",
	})
}
