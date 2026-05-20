package shell

import "time"

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

type Service struct {
	startedAt time.Time
}

func NewService(startedAt time.Time) *Service {
	return &Service{startedAt: startedAt.UTC()}
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
		},
	}
}

func (s *Service) Health() Health {
	return Health{
		Status:          "ready",
		Severity:        "info",
		UserMessage:     "Desktop shell is ready.",
		TechnicalDetail: "Domain services are intentionally not mounted in TOO-160.",
		RecoveryActions: []string{
			"Continue with TOO-161 to mount the Workbench first screen.",
		},
		CheckedAt: time.Now().UTC().Format(time.RFC3339),
	}
}
