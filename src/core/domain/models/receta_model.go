package domainModels

const (
	TYPE_SIZE_SM = "sm"
	TYPE_SIZE_MD = "md"
	TYPE_SIZE_LG = "lg"
)

var RECIPE_PRIORITY_HIGH = map[string]RecipeMediaModel{
	CATEGORY_PROCESS_USER_AVATAR: {
		Name: "UserAvatar",
		TargetSize: []MediaSizeModel{
			{Size: TYPE_SIZE_SM, Width: 150, Height: 150},
			{Size: TYPE_SIZE_MD, Width: 500, Height: 500},
			{Size: TYPE_SIZE_LG, Width: 1000, Height: 1000},
		},
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
		Priority: 8,
	},
}

type RecipeMediaModel struct {
	Name       string
	TargetSize []MediaSizeModel // ["sm", "md", "lg"]
	Format     string           // "webp"
	radio      float64
	Priority   int
}
type MediaSizeModel struct {
	Size     string // "sm", "md", "lg"
	Width    int
	Height   int
	Format   string // "webp"
	Priority int
}
