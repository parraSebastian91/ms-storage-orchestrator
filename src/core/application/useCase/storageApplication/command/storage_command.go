package command

type GetPresignedPutURLCommand struct {
	UUID          string
	Gestor        string
	ObjectType    string
	FileName      string
	ContentType   string
	CorrelationId string
	Organization  string
}
