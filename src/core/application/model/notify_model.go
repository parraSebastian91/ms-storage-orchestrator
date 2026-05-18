package AplicationModel

type NotifyModel struct {
	Category      string        `json:"category"`
	Status        string        `json:"status"`
	Timestamp     string        `json:"timestamp"`
	CorrelationId string        `json:"correlationId"`
	OwnerUUID     string        `json:"ownerUUID"`
	Gestor        string        `json:"gestor"`
	App           string        `json:"app"`
	Payload       NotifyPayload `json:"payload"`
	Asset_id      string        `json:"asset_id"`
}

type NotifyPayload struct {
	NumeroFactura []string `json:"numeroFactura"`
	RutDeudor     []string `json:"rutDeudor"`
	NombreDeudor  []string `json:"nombreDeudor"`
	MontoTotal    []string `json:"montoTotal"`
}
