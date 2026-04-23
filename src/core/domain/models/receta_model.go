package domainModels

const (
	TYPE_SIZE_SM = "sm"
	TYPE_SIZE_MD = "md"
	TYPE_SIZE_LG = "lg"
)

var RECIPE_IMAGE = map[string]RecipeMediaModel{
	CATEGORY_PROCESS_USER_AVATAR: {
		Name: "UserAvatar",
		TargetSize: []MediaSizeModel{
			{Size: TYPE_SIZE_SM, Width: 150, Height: 150},
			{Size: TYPE_SIZE_MD, Width: 500, Height: 500},
			{Size: TYPE_SIZE_LG, Width: 1000, Height: 1000},
		},
		Radio:    1.0,
		Format:   "webp",
		Priority: 10,
	},
	CATEGORY_PROCESS_USER_BANNER: {
		Name: "UserBanner",
		TargetSize: []MediaSizeModel{
			{Size: TYPE_SIZE_SM, Width: 300, Height: 100},
			{Size: TYPE_SIZE_MD, Width: 600, Height: 200},
			{Size: TYPE_SIZE_LG, Width: 1200, Height: 400},
		},
		Format:   "webp",
		Radio:    1.0,
		Priority: 8,
	},
}

var RECIPE_DOCUMENT = map[string]DocumentRecipeModel{
	CATEGORY_PROCESS_DOCUMENT_DTO: {
		Name:        "DocumentUpload",
		OcrLanguage: "spa+eng",
		Category:    CATEGORY_PROCESS_DOCUMENT_DTO,
	},
}

type DocumentRecipeModel struct {
	Name        string `json:"name"`
	OcrLanguage string `json:"ocr_language"`
	Category    string `json:"category"`
}

type RecipeMediaModel struct {
	Name       string           `json:"name"`
	TargetSize []MediaSizeModel `json:"target_size"` // ["sm", "md", "lg"]
	Format     string           `json:"format"`      // "webp"
	Radio      float64          `json:"radio"`
	Priority   int              `json:"priority"`
}
type MediaSizeModel struct {
	Size     string `json:"size"` // "sm", "md", "lg"
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Format   string `json:"format"` // "webp"
	Priority int    `json:"priority"`
}
