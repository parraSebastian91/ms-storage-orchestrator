package controller

import (
	"github.com/gofiber/fiber/v3"
	"github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/application"
	"github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/application/storageApplication/command"
	ports "github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/domain/ports/outbound"
	"github.com/parraSebastian91/ms-storage-orchestrator.git/src/infrastructure/adapter/inbound/http/dto"
)

type StorageController struct {
	storageApplication application.IStorageApplicationUseCase
	logger             ports.ILoggerService
}

// NewStorageController crea una nueva instancia de StorageController
func NewStorageController(storageApplication application.IStorageApplicationUseCase, logger ports.ILoggerService) *StorageController {
	return &StorageController{
		storageApplication: storageApplication,
		logger:             logger,
	}
}
func (uc *StorageController) UploadFile(ctx fiber.Ctx) error {
	var UploadFormDTO dto.UploadFormDTO

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
