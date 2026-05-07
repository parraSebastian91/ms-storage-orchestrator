package AplicationModel

type NotifyModel struct {
	Category      string        `json:"category"`
	Status        string        `json:"status"`
	Timestamp     string        `json:"timestamp"`
	CorrelationId string        `json:"correlationId"`
	OwnerUUID     string        `json:"ownerUUID"`
	App           string        `json:"app"`
	Payload       NotifyPayload `json:"payload"`
}

type NotifyPayload struct {
	NumeroFactura []string `json:"numeroFactura"`
	RutDeudor     []string `json:"rutDeudor"`
	NombreDeudor  []string `json:"nombreDeudor"`
	MontoTotal    []string `json:"montoTotal"`
}
