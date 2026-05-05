package inbound_dto

type FacturaDataDTO struct {
	NumeroFactura []string `json:"numero_factura"`
	RutDeudor     []string `json:"rut_deudor"`
	NombreDeudor  []string `json:"nombre_deudor"`
	MontoTotal    []string `json:"monto_total"`
	FullText      []string `json:"full_text"`
}
