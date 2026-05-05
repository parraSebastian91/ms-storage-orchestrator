package AplicationModel

type NotifyModel struct {
	Category      string
	Status        string
	Timestamp     string
	CorrelationId string
	App           string
	Payload       NotifyPayload
}

type NotifyPayload struct {
	NumeroFactura []string
	RutDeudor     []string
	NombreDeudor  []string
	MontoTotal    []string
}
