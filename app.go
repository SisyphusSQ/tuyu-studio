package main

import (
	"context"

	"github.com/SisyphusSQ/tuyu-studio/internal/project"
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

func (a *App) ProjectCreate(command project.CreateProjectCommand) project.OperationResult {
	return a.service.ProjectCreate(command)
}

func (a *App) ProjectOpen(command project.OpenProjectCommand) project.OperationResult {
	return a.service.ProjectOpen(command)
}

func (a *App) ProjectSave(command project.SaveProjectCommand) project.OperationResult {
	return a.service.ProjectSave(command)
}

func (a *App) ProjectHealth(command project.CheckProjectHealthCommand) project.OperationResult {
	return a.service.ProjectHealth(command)
}

func (a *App) ProjectGraphView(command project.GraphViewCommand) project.GraphViewResult {
	return a.service.ProjectGraphView(command)
}

func (a *App) ProjectGraphLayoutSave(command project.SaveGraphLayoutCommand) project.GraphViewResult {
	return a.service.ProjectGraphLayoutSave(command)
}

func (a *App) ProjectAssetImport(command project.ImportAssetCommand) project.AssetLibraryResult {
	return a.service.ProjectAssetImport(command)
}

func (a *App) ProjectAssetsList(command project.ListAssetsCommand) project.AssetLibraryResult {
	return a.service.ProjectAssetsList(command)
}

func (a *App) ProjectAssetBindingsList(command project.ListAssetBindingsCommand) project.ContinuityLibraryResult {
	return a.service.ProjectAssetBindingsList(command)
}

func (a *App) ProjectAssetBind(command project.BindAssetCommand) project.ContinuityLibraryResult {
	return a.service.ProjectAssetBind(command)
}

func (a *App) ProjectMainReferenceSet(command project.SetMainReferenceCommand) project.ContinuityLibraryResult {
	return a.service.ProjectMainReferenceSet(command)
}

func (a *App) ProjectContinuityList(command project.ListContinuityCommand) project.ContinuityResult {
	return a.service.ProjectContinuityList(command)
}

func (a *App) ProjectContinuityRuleSave(command project.SaveContinuityRuleCommand) project.ContinuityResult {
	return a.service.ProjectContinuityRuleSave(command)
}

func (a *App) ProjectContinuityRuleUnlock(command project.UnlockContinuityRuleCommand) project.ContinuityResult {
	return a.service.ProjectContinuityRuleUnlock(command)
}

func (a *App) ProjectAssetBindingUnlock(command project.UnlockAssetBindingCommand) project.ContinuityResult {
	return a.service.ProjectAssetBindingUnlock(command)
}

func (a *App) ProjectScriptDocumentSave(command project.SaveScriptDocumentCommand) project.ScriptDocumentResult {
	return a.service.ProjectScriptDocumentSave(command)
}

func (a *App) ProjectScriptDocumentLoad(command project.LoadScriptDocumentCommand) project.ScriptDocumentResult {
	return a.service.ProjectScriptDocumentLoad(command)
}

func (a *App) ProjectScriptSceneConfirm(command project.ConfirmScriptSceneCommand) project.ScriptSceneCandidateResult {
	return a.service.ProjectScriptSceneConfirm(command)
}

func (a *App) ProjectShotCandidateSave(command project.SaveShotCandidateCommand) project.ScriptSceneCandidateResult {
	return a.service.ProjectShotCandidateSave(command)
}

func (a *App) ProjectShotCandidatesList(command project.ListShotCandidatesCommand) project.ScriptSceneCandidateResult {
	return a.service.ProjectShotCandidatesList(command)
}

func (a *App) ProjectShotCandidateConfirm(command project.ConfirmShotCandidateCommand) project.ScriptSceneCandidateResult {
	return a.service.ProjectShotCandidateConfirm(command)
}

func (a *App) ProjectShotCandidateReject(command project.RejectShotCandidateCommand) project.ScriptSceneCandidateResult {
	return a.service.ProjectShotCandidateReject(command)
}

func (a *App) ProjectShotContextValidate(command project.ValidateShotContextCommand) project.ShotContextResult {
	return a.service.ProjectShotContextValidate(command)
}

func (a *App) ProjectShotContextPromote(command project.PromoteShotContextCommand) project.ShotContextResult {
	return a.service.ProjectShotContextPromote(command)
}

func (a *App) ProjectShotContextMarkDirty(command project.MarkShotContextDirtyCommand) project.ShotContextResult {
	return a.service.ProjectShotContextMarkDirty(command)
}
