package command

type GetPresignedPutURLCommand struct {
	UUID          string
	ObjectType    string
	FileName      string
	ContentType   string
	CorrelationId string
}
