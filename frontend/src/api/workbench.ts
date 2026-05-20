import {
  ProjectCreate,
  ProjectGraphLayoutSave,
  ProjectGraphView,
  ProjectHealth,
  ProjectOpen,
  ProjectSave,
  WorkbenchProbe,
} from '../../wailsjs/go/main/App'
import { project } from '../../wailsjs/go/models'

import {
  createCorrelationId,
  normalizeProjectGraphViewResult,
  normalizeProjectOperationResult,
  normalizeProbeResult,
  normalizeUnknownError,
  type ProjectGraphLayoutSaveCommandDTO,
  type ProjectGraphViewCommandDTO,
  type ProjectGraphViewResultDTO,
  type ProjectOperationName,
  type ProjectOperationResultDTO,
  type WorkbenchProbeMode,
  type WorkbenchProbeResultDTO,
} from './dto'
import { DEFAULT_ALPHA_PROJECT_ROOT, normalizeProjectRoot } from './projectRoot'

export interface ProjectGraphViewOptions {
  root?: string
  expectedGraphVersion?: number
}

export async function runWorkbenchProbe(mode: WorkbenchProbeMode): Promise<WorkbenchProbeResultDTO> {
  const correlationId = createCorrelationId(`workbench-${mode}`)

  try {
    const result = await WorkbenchProbe({
      mode,
      correlationId,
    })

    return normalizeProbeResult(result as Partial<WorkbenchProbeResultDTO>, correlationId)
  } catch (error) {
    const appError = normalizeUnknownError(error, correlationId)

    return {
      ok: false,
      error: appError,
      events: [
        {
          eventId: `${correlationId}-transport-error`,
          eventType: 'workbench.probe',
          state: 'failed',
          progress: 0,
          summary: appError.userMessage,
          error: appError,
          nextActions: appError.recoveryActions,
          createdAt: new Date().toISOString(),
        },
      ],
    }
  }
}

export async function runProjectOperation(action: ProjectOperationName): Promise<ProjectOperationResultDTO> {
  const correlationId = createCorrelationId(`project-${action}`)

  try {
    const result = await callProjectOperation(action, correlationId)

    return normalizeProjectOperationResult(result as Partial<ProjectOperationResultDTO>, correlationId)
  } catch (error) {
    const appError = normalizeUnknownError(error, correlationId)

    return {
      ok: false,
      error: appError,
      events: [
        {
          eventId: `${correlationId}-transport-error`,
          eventType: `project.${action}`,
          state: 'failed',
          summary: appError.userMessage,
          error: appError,
          nextActions: appError.recoveryActions,
          createdAt: new Date().toISOString(),
        },
      ],
    }
  }
}

export async function runProjectGraphView(
  options?: number | ProjectGraphViewOptions,
): Promise<ProjectGraphViewResultDTO> {
  const correlationId = createCorrelationId('project-graph-view')

  try {
    const result = await ProjectGraphView(project.GraphViewCommand.createFrom(
      buildProjectGraphViewCommand(options, correlationId),
    ))

    return normalizeProjectGraphViewResult(result as Partial<ProjectGraphViewResultDTO>, correlationId)
  } catch (error) {
    const appError = normalizeUnknownError(error, correlationId)

    return {
      ok: false,
      error: appError,
      errors: [appError],
      events: [
        {
          eventId: `${correlationId}-transport-error`,
          eventType: 'project.graph_view',
          state: 'failed',
          summary: appError.userMessage,
          error: appError,
          nextActions: appError.recoveryActions,
          createdAt: new Date().toISOString(),
        },
      ],
    }
  }
}

export async function saveProjectGraphLayout(
  command: ProjectGraphLayoutSaveCommandDTO,
): Promise<ProjectGraphViewResultDTO> {
  const correlationId = command.correlationId || createCorrelationId('project-graph-layout-save')

  try {
    const result = await ProjectGraphLayoutSave(project.SaveGraphLayoutCommand.createFrom(
      normalizeProjectGraphLayoutSaveCommand(command, correlationId),
    ))

    return normalizeProjectGraphViewResult(result as Partial<ProjectGraphViewResultDTO>, correlationId)
  } catch (error) {
    const appError = normalizeUnknownError(error, correlationId)

    return {
      ok: false,
      error: appError,
      errors: [appError],
      events: [
        {
          eventId: `${correlationId}-transport-error`,
          eventType: 'project.graph_layout_save',
          state: 'failed',
          summary: appError.userMessage,
          error: appError,
          nextActions: appError.recoveryActions,
          createdAt: new Date().toISOString(),
        },
      ],
    }
  }
}

export function buildProjectGraphViewCommand(
  options: number | ProjectGraphViewOptions | undefined,
  correlationId: string,
): ProjectGraphViewCommandDTO {
  const resolvedOptions = typeof options === 'number'
    ? { expectedGraphVersion: options }
    : options || {}

  return {
    root: normalizeProjectRoot(resolvedOptions.root),
    expectedGraphVersion: resolvedOptions.expectedGraphVersion,
    correlationId,
  }
}

export function normalizeProjectGraphLayoutSaveCommand(
  command: ProjectGraphLayoutSaveCommandDTO,
  correlationId: string,
): ProjectGraphLayoutSaveCommandDTO {
  return {
    ...command,
    root: normalizeProjectRoot(command.root),
    correlationId,
  }
}

export function projectOperationRoot(action: ProjectOperationName): string {
  return action === 'create' ? '' : DEFAULT_ALPHA_PROJECT_ROOT
}

function callProjectOperation(action: ProjectOperationName, correlationId: string): Promise<unknown> {
  switch (action) {
    case 'create':
      return ProjectCreate(project.CreateProjectCommand.createFrom({
        root: projectOperationRoot(action),
        projectId: '',
        name: 'Tuyu Alpha Shell Project',
        type: 'series',
        correlationId,
      }))
    case 'open':
      return ProjectOpen(project.OpenProjectCommand.createFrom({
        root: projectOperationRoot(action),
        takeover: false,
        correlationId,
      }))
    case 'save':
      return ProjectSave(project.SaveProjectCommand.createFrom({
        root: projectOperationRoot(action),
        correlationId,
      }))
    case 'health':
      return ProjectHealth(project.CheckProjectHealthCommand.createFrom({
        root: projectOperationRoot(action),
        correlationId,
      }))
  }
}
