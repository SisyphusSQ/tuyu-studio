import {
  ProjectMockRunCancel,
  ProjectMockRunRetry,
  ProjectMockRunStart,
} from '../../wailsjs/go/main/App'
import { project } from '../../wailsjs/go/models'

import {
  createCorrelationId,
  normalizeMockRunResult,
  normalizeUnknownError,
  type MockRunCommandDTO,
  type MockRunResultDTO,
} from './dto'
import { normalizeProjectRoot } from './projectRoot'

export const DEFAULT_MOCK_RUN_TASK_MODE = 'mock_local'
export const DEFAULT_MOCK_RUN_CREATED_BY = 'local_user'

export interface MockRunInput {
  root?: string
  runId?: string
  shotId?: string
  packageId?: string
  selectionIds?: string[]
  cancelReason?: string
}

export async function startMockRun(input: MockRunInput): Promise<MockRunResultDTO> {
  const correlationId = createCorrelationId('mock-run-start')

  try {
    const result = await ProjectMockRunStart(project.MockRunCommand.createFrom(
      buildStartMockRunCommand(input, correlationId),
    ))

    return normalizeMockRunResult(result as Partial<MockRunResultDTO>, correlationId)
  } catch (error) {
    return mockRunTransportFailure(error, correlationId, 'run.start')
  }
}

export async function cancelMockRun(input: MockRunInput): Promise<MockRunResultDTO> {
  const correlationId = createCorrelationId('mock-run-cancel')

  try {
    const result = await ProjectMockRunCancel(project.MockRunCommand.createFrom(
      buildCancelMockRunCommand(input, correlationId),
    ))

    return normalizeMockRunResult(result as Partial<MockRunResultDTO>, correlationId)
  } catch (error) {
    return mockRunTransportFailure(error, correlationId, 'run.cancel')
  }
}

export async function retryMockRun(input: MockRunInput): Promise<MockRunResultDTO> {
  const correlationId = createCorrelationId('mock-run-retry')

  try {
    const result = await ProjectMockRunRetry(project.MockRunCommand.createFrom(
      buildRetryMockRunCommand(input, correlationId),
    ))

    return normalizeMockRunResult(result as Partial<MockRunResultDTO>, correlationId)
  } catch (error) {
    return mockRunTransportFailure(error, correlationId, 'run.retry')
  }
}

export function buildStartMockRunCommand(input: MockRunInput, correlationId: string): MockRunCommandDTO {
  return normalizeMockRunCommand({
    ...input,
    correlationId,
  })
}

export function buildCancelMockRunCommand(input: MockRunInput, correlationId: string): MockRunCommandDTO {
  return normalizeMockRunCommand({
    ...input,
    cancelReason: input.cancelReason?.trim() || 'user_cancelled',
    correlationId,
  })
}

export function buildRetryMockRunCommand(input: MockRunInput, correlationId: string): MockRunCommandDTO {
  return normalizeMockRunCommand({
    ...input,
    retryOfRunId: input.runId?.trim(),
    correlationId,
  })
}

function normalizeMockRunCommand(input: MockRunCommandDTO, correlationId = ''): MockRunCommandDTO {
  return {
    root: normalizeProjectRoot(input.root),
    runId: input.runId?.trim() || undefined,
    shotId: input.shotId?.trim() || undefined,
    packageId: input.packageId?.trim() || undefined,
    selectionIds: Array.isArray(input.selectionIds)
      ? input.selectionIds.map((item) => item.trim()).filter(Boolean)
      : [],
    taskMode: DEFAULT_MOCK_RUN_TASK_MODE,
    retryOfRunId: input.retryOfRunId?.trim() || undefined,
    cancelReason: input.cancelReason?.trim() || undefined,
    createdBy: DEFAULT_MOCK_RUN_CREATED_BY,
    correlationId: input.correlationId || correlationId,
  }
}

function mockRunTransportFailure(error: unknown, correlationId: string, eventType: string): MockRunResultDTO {
  const appError = normalizeUnknownError(error, correlationId)

  return {
    ok: false,
    error: appError,
    events: [
      {
        eventId: `${correlationId}-transport-error`,
        eventType,
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
