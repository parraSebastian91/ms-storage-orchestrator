package inbound_dto

type NotifyProcessDTO struct {
	Category      string         `json:"category" validate:"required"`
	Status        string         `json:"status" validate:"required"`
	Timestamp     string         `json:"timestamp" validate:"required"`
	App           string         `json:"app" validate:"required"`
	CorrelationId string         `json:"correlation_id"`
	OwnerUUID     string         `json:"owner_uuid" validate:"required"`
	Gestor        string         `json:"gestor" validate:"required"`
	Payload       FacturaDataDTO `json:"payload" validate:"required"`
}
