package AplicationModel

type StorageModel struct {
	AssetId         string `json:"asset_id"`
	OwnerUUID       string `json:"owner_uuid"`
	MediaType       string `json:"media_type"`
	CategoryProcess string `json:"category_process"`
	NameFile        string `json:"name_file"`
	FormatFile      string `json:"format_file"`
	StorageKey      string `json:"storage_key"`
}
