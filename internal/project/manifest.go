package project

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type Manifest struct {
	SchemaVersion string          `json:"schemaVersion"`
	Project       ProjectMetadata `json:"project"`
	Paths         Paths           `json:"paths"`
	Defaults      Defaults        `json:"defaults"`
	Integrity     Integrity       `json:"integrity"`
	Graph         Graph           `json:"graph"`
	Principles    Principles      `json:"principles"`
	StyleBible    StyleBible      `json:"styleBible"`
}

type ProjectMetadata struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type Paths struct {
	Root            string `json:"root"`
	Characters      string `json:"characters"`
	Scenes          string `json:"scenes"`
	Props           string `json:"props"`
	Shots           string `json:"shots"`
	Prompts         string `json:"prompts"`
	PromptRuns      string `json:"promptRuns"`
	Assets          string `json:"assets"`
	AssetInputs     string `json:"assetInputs"`
	AssetRefs       string `json:"assetRefs"`
	AssetOutputs    string `json:"assetOutputs"`
	AssetResults    string `json:"assetResults"`
	AssetThumbnails string `json:"assetThumbnails"`
	Packages        string `json:"packages"`
	Audit           string `json:"audit"`
	Backups         string `json:"backups"`
	Locks           string `json:"locks"`
}

type Defaults struct {
	ProviderProfileID string `json:"providerProfileId"`
	Language          string `json:"language"`
	AspectRatio       string `json:"aspectRatio"`
}

type Integrity struct {
	LastCleanShutdown bool   `json:"lastCleanShutdown"`
	LastGraphVersion  int    `json:"lastGraphVersion"`
	LastHealthCheckAt string `json:"lastHealthCheckAt,omitempty"`
}

type Graph struct {
	Version  int      `json:"version"`
	Nodes    []Node   `json:"nodes"`
	Edges    []Edge   `json:"edges"`
	Viewport Viewport `json:"viewport"`
}

type Node struct {
	ID     string `json:"id,omitempty"`
	Kind   string `json:"kind,omitempty"`
	RefID  string `json:"refId,omitempty"`
	Status string `json:"status,omitempty"`
}

type Edge struct {
	ID     string `json:"id,omitempty"`
	Source string `json:"source,omitempty"`
	Target string `json:"target,omitempty"`
	Kind   string `json:"kind,omitempty"`
}

type Viewport struct {
	X    int     `json:"x"`
	Y    int     `json:"y"`
	Zoom float64 `json:"zoom"`
}

type Principles struct {
	Summary     string   `json:"summary"`
	LockedRules []string `json:"lockedRules"`
}

type StyleBible struct {
	Summary     string   `json:"summary"`
	VisualRules []string `json:"visualRules"`
}

type ManifestInput struct {
	ProjectID string
	Name      string
	Type      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewBaselineManifest(input ManifestInput) (Manifest, error) {
	if input.ProjectID == "" {
		return Manifest{}, errors.New("project id is required")
	}
	if input.Name == "" {
		return Manifest{}, errors.New("project name is required")
	}

	projectType := input.Type
	if projectType == "" {
		projectType = "series"
	}

	createdAt := input.CreatedAt.UTC()
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	updatedAt := input.UpdatedAt.UTC()
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}

	return Manifest{
		SchemaVersion: CurrentSchemaVersion,
		Project: ProjectMetadata{
			ID:        input.ProjectID,
			Name:      input.Name,
			Type:      projectType,
			CreatedAt: createdAt.Format(time.RFC3339),
			UpdatedAt: updatedAt.Format(time.RFC3339),
		},
		Paths: DefaultPaths(),
		Defaults: Defaults{
			ProviderProfileID: "provider_manual_handoff",
			Language:          "zh",
			AspectRatio:       "9:16",
		},
		Integrity: Integrity{
			LastCleanShutdown: true,
			LastGraphVersion:  1,
		},
		Graph: Graph{
			Version:  1,
			Nodes:    []Node{},
			Edges:    []Edge{},
			Viewport: Viewport{X: 0, Y: 0, Zoom: 1},
		},
		Principles: Principles{
			LockedRules: []string{},
		},
		StyleBible: StyleBible{
			VisualRules: []string{},
		},
	}, nil
}

func EncodeManifest(manifest Manifest) ([]byte, error) {
	report := ValidateManifest(manifest)
	if report.HasBlocking() {
		return nil, ValidationError{Report: report}
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal project manifest: %w", err)
	}

	return append(data, '\n'), nil
}

func DecodeManifest(data []byte) (Manifest, ValidationReport) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return Manifest{}, ValidationReport{
			Items: []ValidationIssue{{
				Severity:        SeverityBlocking,
				Code:            CodeManifestJSONInvalid,
				Path:            ManifestFileName,
				UserMessage:     "Project manifest is not valid JSON.",
				TechnicalDetail: err.Error(),
				RecoveryActions: []string{"Restore project.tuyu.json from a backup or keep the project in diagnostic mode."},
			}},
		}
	}

	issues := validateRequiredRawFields(raw)

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		issues = append(issues, ValidationIssue{
			Severity:        SeverityBlocking,
			Code:            CodeManifestJSONInvalid,
			Path:            ManifestFileName,
			UserMessage:     "Project manifest shape is invalid.",
			TechnicalDetail: err.Error(),
			RecoveryActions: []string{"Fix the manifest shape or restore project.tuyu.json from a backup."},
		})
		return Manifest{}, ValidationReport{Items: issues}
	}

	report := ValidateManifest(manifest)
	issues = append(issues, report.Items...)

	return manifest, ValidationReport{Items: dedupeIssues(issues)}
}

func (paths Paths) All() map[string]string {
	return map[string]string{
		PathRoot:            paths.Root,
		PathCharacters:      paths.Characters,
		PathScenes:          paths.Scenes,
		PathProps:           paths.Props,
		PathShots:           paths.Shots,
		PathPrompts:         paths.Prompts,
		PathPromptRuns:      paths.PromptRuns,
		PathAssets:          paths.Assets,
		PathAssetInputs:     paths.AssetInputs,
		PathAssetRefs:       paths.AssetRefs,
		PathAssetOutputs:    paths.AssetOutputs,
		PathAssetResults:    paths.AssetResults,
		PathAssetThumbnails: paths.AssetThumbnails,
		PathPackages:        paths.Packages,
		PathAudit:           paths.Audit,
		PathBackups:         paths.Backups,
		PathLocks:           paths.Locks,
	}
}
