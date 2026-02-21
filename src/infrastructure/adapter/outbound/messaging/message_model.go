package messaging

type Message struct {
	Body        []byte
	ContentType string
	RoutingKey  string
}

type Work struct {
	Message
	Headers map[string]interface{}
}
