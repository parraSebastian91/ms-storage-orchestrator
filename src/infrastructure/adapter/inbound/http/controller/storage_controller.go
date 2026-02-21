package controller

import (
	"github.com/gofiber/fiber/v3"
	"github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/application"
	ports "github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/domain/ports/outbound"
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
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to get file",
		})
	}

	file, err := fileHeader.Open()
	if err != nil {
		uc.logger.Error("Failed to open file", map[string]interface{}{
			"error": err.Error(),
		})
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to process file",
		})
	}
	defer file.Close()

	// Detectar content type
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	uc.storageApplication.UploadFile(ctx, file, fileHeader.Filename, fileHeader.Size)

	return nil
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
