package controller

import (
	"time"

	"github.com/gofiber/fiber/v3"
)

type HealthcheckController struct{}

func NewHealthcheckController() *HealthcheckController {
	return &HealthcheckController{}
}

func (h *HealthcheckController) Check(ctx fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":    "ok",
		"service":   "ms-storage-orchestrator",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
