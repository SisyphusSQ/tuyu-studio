import { ProjectCreate, ProjectHealth, ProjectOpen, ProjectSave, WorkbenchProbe } from '../../wailsjs/go/main/App'
import { project } from '../../wailsjs/go/models'

import {
  createCorrelationId,
  normalizeProjectOperationResult,
  normalizeProbeResult,
  normalizeUnknownError,
  type ProjectOperationName,
  type ProjectOperationResultDTO,
  type WorkbenchProbeMode,
  type WorkbenchProbeResultDTO,
} from './dto'

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

function callProjectOperation(action: ProjectOperationName, correlationId: string): Promise<unknown> {
  switch (action) {
    case 'create':
      return ProjectCreate(project.CreateProjectCommand.createFrom({
        root: '',
        projectId: '',
        name: 'Tuyu Alpha Shell Project',
        type: 'series',
        correlationId,
      }))
    case 'open':
      return ProjectOpen(project.OpenProjectCommand.createFrom({
        root: '',
        takeover: false,
        correlationId,
      }))
    case 'save':
      return ProjectSave(project.SaveProjectCommand.createFrom({
        root: '',
        correlationId,
      }))
    case 'health':
      return ProjectHealth(project.CheckProjectHealthCommand.createFrom({
        root: '',
        correlationId,
      }))
  }
}
