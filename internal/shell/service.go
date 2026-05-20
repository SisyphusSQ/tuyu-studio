package shell

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/SisyphusSQ/tuyu-studio/internal/project"
)

type Info struct {
	AppName      string   `json:"appName"`
	Stage        string   `json:"stage"`
	ShellVersion string   `json:"shellVersion"`
	StartedAt    string   `json:"startedAt"`
	Capabilities []string `json:"capabilities"`
}

type Health struct {
	Status          string   `json:"status"`
	Severity        string   `json:"severity"`
	UserMessage     string   `json:"userMessage"`
	TechnicalDetail string   `json:"technicalDetail"`
	RecoveryActions []string `json:"recoveryActions"`
	CheckedAt       string   `json:"checkedAt"`
}

type AppError struct {
	Code            string   `json:"code"`
	Severity        string   `json:"severity"`
	Retryable       bool     `json:"retryable"`
	TargetType      string   `json:"targetType,omitempty"`
	TargetID        string   `json:"targetId,omitempty"`
	UserMessage     string   `json:"userMessage"`
	TechnicalDetail string   `json:"technicalDetail,omitempty"`
	RecoveryActions []string `json:"recoveryActions"`
	CorrelationID   string   `json:"correlationId"`
}

type RuntimeEvent struct {
	EventID     string    `json:"eventId"`
	RunID       string    `json:"runId,omitempty"`
	EventType   string    `json:"eventType"`
	State       string    `json:"state"`
	Progress    int       `json:"progress"`
	TargetType  string    `json:"targetType,omitempty"`
	TargetID    string    `json:"targetId,omitempty"`
	Summary     string    `json:"summary"`
	Error       *AppError `json:"error,omitempty"`
	NextActions []string  `json:"nextActions"`
	CreatedAt   string    `json:"createdAt"`
}

type WorkbenchStatus struct {
	ServiceName  string   `json:"serviceName"`
	Status       string   `json:"status"`
	Summary      string   `json:"summary"`
	Capabilities []string `json:"capabilities"`
	CheckedAt    string   `json:"checkedAt"`
}

type WorkbenchProbeCommand struct {
	Mode          string `json:"mode"`
	CorrelationID string `json:"correlationId"`
}

type WorkbenchProbeResult struct {
	OK       bool            `json:"ok"`
	Snapshot WorkbenchStatus `json:"snapshot"`
	Error    *AppError       `json:"error,omitempty"`
	Events   []RuntimeEvent  `json:"events"`
}

type Service struct {
	startedAt          time.Time
	projectStore       *project.Store
	defaultProjectRoot string
}

func NewService(startedAt time.Time) *Service {
	startedAt = startedAt.UTC()
	return &Service{
		startedAt: startedAt,
		projectStore: project.NewStore(project.StoreOptions{
			InstanceID: "shell-" + startedAt.Format("20060102T150405Z"),
		}),
		defaultProjectRoot: defaultProjectRoot(),
	}
}

func (s *Service) Info() Info {
	return Info{
		AppName:      "Tuyu Studio",
		Stage:        "alpha-shell",
		ShellVersion: "0.1.0-alpha",
		StartedAt:    s.startedAt.Format(time.RFC3339),
		Capabilities: []string{
			"wails_v2_shell",
			"app_facade",
			"frontend_asset_host",
			"project_create",
			"project_open",
			"project_save",
			"project_lock",
			"project_health_check",
			"project_graph_view",
			"project_graph_layout_save",
		},
	}
}

func (s *Service) Health() Health {
	return Health{
		Status:          "ready",
		Severity:        "info",
		UserMessage:     "Desktop shell is ready.",
		TechnicalDetail: "Alpha shell service boundary is available with local project operation DTOs mounted.",
		RecoveryActions: []string{
			"Use the Workbench probe or project operation buttons to verify the local Go service and DTO boundary.",
		},
		CheckedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func (s *Service) ProjectCreate(command project.CreateProjectCommand) project.OperationResult {
	command.Root = s.projectRoot(command.Root)
	if strings.TrimSpace(command.Name) == "" {
		command.Name = "Tuyu Alpha Shell Project"
	}
	return s.projectStore.CreateProject(command)
}

func (s *Service) ProjectOpen(command project.OpenProjectCommand) project.OperationResult {
	command.Root = s.projectRoot(command.Root)
	return s.projectStore.OpenProject(command)
}

func (s *Service) ProjectSave(command project.SaveProjectCommand) project.OperationResult {
	command.Root = s.projectRoot(command.Root)
	return s.projectStore.SaveProject(command)
}

func (s *Service) ProjectHealth(command project.CheckProjectHealthCommand) project.OperationResult {
	command.Root = s.projectRoot(command.Root)
	return s.projectStore.CheckHealth(command)
}

func (s *Service) ProjectGraphView(command project.GraphViewCommand) project.GraphViewResult {
	command.Root = s.projectRoot(command.Root)
	return s.projectStore.GraphView(command)
}

func (s *Service) ProjectGraphLayoutSave(command project.SaveGraphLayoutCommand) project.GraphViewResult {
	command.Root = s.projectRoot(command.Root)
	return s.projectStore.SaveGraphLayout(command)
}

func (s *Service) ProjectScriptDocumentSave(command project.SaveScriptDocumentCommand) project.ScriptDocumentResult {
	command.Root = s.projectRoot(command.Root)
	return s.projectStore.SaveScriptDocument(command)
}

func (s *Service) ProjectScriptDocumentLoad(command project.LoadScriptDocumentCommand) project.ScriptDocumentResult {
	command.Root = s.projectRoot(command.Root)
	return s.projectStore.LoadScriptDocument(command)
}

func (s *Service) WorkbenchProbe(command WorkbenchProbeCommand) WorkbenchProbeResult {
	now := time.Now().UTC()
	checkedAt := now.Format(time.RFC3339)
	mode := strings.TrimSpace(command.Mode)
	correlationID := strings.TrimSpace(command.CorrelationID)
	if correlationID == "" {
		correlationID = "probe-" + now.Format("20060102T150405Z")
	}

	switch mode {
	case "", "status":
		snapshot := WorkbenchStatus{
			ServiceName: "shell.workbench",
			Status:      "ready",
			Summary:     "Go service responded through the Wails facade and API wrapper.",
			Capabilities: []string{
				"facade_probe",
				"structured_error_dto",
				"runtime_event_projection",
			},
			CheckedAt: checkedAt,
		}

		return WorkbenchProbeResult{
			OK:       true,
			Snapshot: snapshot,
			Events: []RuntimeEvent{
				{
					EventID:     "evt-" + correlationID + "-completed",
					EventType:   "workbench.probe",
					State:       "completed",
					Progress:    100,
					TargetType:  "workbench",
					TargetID:    "alpha-shell",
					Summary:     "Workbench probe completed with a structured status DTO.",
					NextActions: []string{"Use the structured-error action to verify AppErrorDTO rendering."},
					CreatedAt:   checkedAt,
				},
			},
		}
	case "structured_error":
		err := AppError{
			Code:            "workbench_probe_blocked",
			Severity:        "blocking",
			Retryable:       false,
			TargetType:      "workbench",
			TargetID:        "alpha-shell",
			UserMessage:     "The probe was intentionally blocked to verify structured error rendering.",
			TechnicalDetail: "TOO-162 structured_error mode returns AppErrorDTO instead of a raw string error.",
			RecoveryActions: []string{
				"Inspect the recovery action list in the Workbench UI.",
				"Run the ready probe to verify the success path.",
			},
			CorrelationID: correlationID,
		}

		return WorkbenchProbeResult{
			OK:    false,
			Error: &err,
			Events: []RuntimeEvent{
				{
					EventID:     "evt-" + correlationID + "-blocked",
					EventType:   "workbench.probe",
					State:       "blocked",
					Progress:    0,
					TargetType:  "workbench",
					TargetID:    "alpha-shell",
					Summary:     "Workbench probe produced a structured blocking error.",
					Error:       &err,
					NextActions: err.RecoveryActions,
					CreatedAt:   checkedAt,
				},
			},
		}
	default:
		err := AppError{
			Code:            "workbench_probe_mode_invalid",
			Severity:        "error",
			Retryable:       true,
			TargetType:      "workbench",
			TargetID:        "alpha-shell",
			UserMessage:     "Unsupported probe mode.",
			TechnicalDetail: "mode=" + mode,
			RecoveryActions: []string{"Use status or structured_error as the probe mode."},
			CorrelationID:   correlationID,
		}

		return WorkbenchProbeResult{
			OK:    false,
			Error: &err,
			Events: []RuntimeEvent{
				{
					EventID:     "evt-" + correlationID + "-failed",
					EventType:   "workbench.probe",
					State:       "failed",
					Progress:    0,
					TargetType:  "workbench",
					TargetID:    "alpha-shell",
					Summary:     "Workbench probe rejected an unsupported mode.",
					Error:       &err,
					NextActions: err.RecoveryActions,
					CreatedAt:   checkedAt,
				},
			},
		}
	}
}

func (s *Service) projectRoot(root string) string {
	root = strings.TrimSpace(root)
	if root != "" {
		return resolveProjectRoot(root)
	}
	return s.defaultProjectRoot
}

func resolveProjectRoot(root string) string {
	if filepath.IsAbs(root) {
		return root
	}

	for _, base := range projectRootBases() {
		candidate := filepath.Clean(filepath.Join(base, root))
		if hasProjectManifest(candidate) {
			return candidate
		}
	}

	return root
}

func projectRootBases() []string {
	bases := make([]string, 0, 2)
	if cwd, err := os.Getwd(); err == nil && strings.TrimSpace(cwd) != "" {
		bases = append(bases, cwd)
	}
	if _, filename, _, ok := runtime.Caller(0); ok {
		bases = append(bases, filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..")))
	}
	return bases
}

func hasProjectManifest(root string) bool {
	info, err := os.Stat(filepath.Join(root, project.ManifestFileName))
	return err == nil && !info.IsDir()
}

func defaultProjectRoot() string {
	base, err := os.UserCacheDir()
	if err != nil || strings.TrimSpace(base) == "" {
		base = os.TempDir()
	}
	return filepath.Join(base, "tuyu-studio", "alpha-shell-project")
}
