package main

import (
	"context"

	"github.com/SisyphusSQ/tuyu-studio/internal/shell"
)

// App is the Wails-facing facade. It should stay thin: validate boundary input,
// call application services, and return DTOs that the frontend can render.
type App struct {
	ctx     context.Context
	service *shell.Service
}

func NewApp(service *shell.Service) *App {
	return &App{service: service}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) AppInfo() shell.Info {
	return a.service.Info()
}

func (a *App) ShellHealth() shell.Health {
	return a.service.Health()
}

func (a *App) WorkbenchProbe(command shell.WorkbenchProbeCommand) shell.WorkbenchProbeResult {
	return a.service.WorkbenchProbe(command)
}
