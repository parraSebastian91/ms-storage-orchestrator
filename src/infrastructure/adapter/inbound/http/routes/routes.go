package routes

import (
	"github.com/gofiber/fiber/v3"
	"github.com/parraSebastian91/ms-storage-orchestrator.git/src/infrastructure/adapter/inbound/http/controller"
)

func SetupRoutes(app *fiber.App, storageController *controller.StorageController) {

	api := app.Group("/api/v1")
	api.Post("/upload", storageController.UploadFile)
	//api.Get("/download/:objectName", storageController.DownloadFile)
}
