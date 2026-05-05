package dto

type NotifyProcessDTO struct {
	Category      string         `json:"category" validate:"required"`
	Status        string         `json:"status" validate:"required"`
	Timestamp     string         `json:"timestamp" validate:"required"`
	App           string         `json:"app" validate:"required"`
	CorrelationId string         `json:"correlationId"`
	Payload       FacturaDataDTO `json:"payload" validate:"required"`
}
