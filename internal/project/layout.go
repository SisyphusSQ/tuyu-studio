package project

const (
	ManifestFileName     = "project.tuyu.json"
	CurrentSchemaVersion = "1.0.0"
)

const (
	PathRoot            = "root"
	PathCharacters      = "characters"
	PathScenes          = "scenes"
	PathProps           = "props"
	PathShots           = "shots"
	PathPrompts         = "prompts"
	PathPromptRuns      = "promptRuns"
	PathAssets          = "assets"
	PathAssetInputs     = "assetInputs"
	PathAssetRefs       = "assetRefs"
	PathAssetOutputs    = "assetOutputs"
	PathAssetResults    = "assetResults"
	PathAssetThumbnails = "assetThumbnails"
	PathPackages        = "packages"
	PathAudit           = "audit"
	PathBackups         = "backups"
	PathLocks           = "locks"
)

var requiredProjectDirectories = []string{
	"characters",
	"scenes",
	"props",
	"shots",
	"prompts",
	"prompts/runs",
	"assets",
	"assets/inputs",
	"assets/refs",
	"assets/outputs",
	"assets/results",
	"assets/thumbnails",
	"packages",
	"audit",
	"backups",
	"locks",
}

func RequiredProjectDirectories() []string {
	dirs := make([]string, len(requiredProjectDirectories))
	copy(dirs, requiredProjectDirectories)
	return dirs
}

func DefaultPaths() Paths {
	return Paths{
		Root:            ".",
		Characters:      "characters",
		Scenes:          "scenes",
		Props:           "props",
		Shots:           "shots",
		Prompts:         "prompts",
		PromptRuns:      "prompts/runs",
		Assets:          "assets",
		AssetInputs:     "assets/inputs",
		AssetRefs:       "assets/refs",
		AssetOutputs:    "assets/outputs",
		AssetResults:    "assets/results",
		AssetThumbnails: "assets/thumbnails",
		Packages:        "packages",
		Audit:           "audit",
		Backups:         "backups",
		Locks:           "locks",
	}
}
