export type ErrorSeverity = 'info' | 'warning' | 'error' | 'blocking'
export type RuntimeEventState = 'queued' | 'running' | 'blocked' | 'completed' | 'failed' | 'cancelled'
export type WorkbenchProbeMode = 'status' | 'structured_error'

export interface AppErrorDTO {
  code: string
  severity: ErrorSeverity
  retryable: boolean
  targetType?: string
  targetId?: string
  userMessage: string
  technicalDetail?: string
  recoveryActions: string[]
  correlationId: string
}

export interface RuntimeEventDTO {
  eventId: string
  runId?: string
  eventType: string
  state: RuntimeEventState
  progress: number
  targetType?: string
  targetId?: string
  summary: string
  error?: AppErrorDTO
  nextActions: string[]
  createdAt: string
}

export interface WorkbenchStatusDTO {
  serviceName: string
  status: string
  summary: string
  capabilities: string[]
  checkedAt: string
}

export interface WorkbenchProbeCommandDTO {
  mode: WorkbenchProbeMode
  correlationId: string
}

export interface WorkbenchProbeResultDTO {
  ok: boolean
  snapshot?: WorkbenchStatusDTO
  error?: AppErrorDTO
  events: RuntimeEventDTO[]
}

export function createCorrelationId(prefix = 'ui'): string {
  return `${prefix}-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 8)}`
}

export function normalizeUnknownError(error: unknown, correlationId: string): AppErrorDTO {
  if (isAppErrorDTO(error)) {
    return error
  }

  const technicalDetail = error instanceof Error ? error.message : String(error)

  return {
    code: 'transport_error',
    severity: 'error',
    retryable: true,
    userMessage: 'Workbench could not reach the local Go service.',
    technicalDetail,
    recoveryActions: [
      'Confirm the Wails desktop bridge is running.',
      'Retry the probe after the Workbench finishes loading.',
    ],
    correlationId,
  }
}

export function normalizeProbeResult(
  result: Partial<WorkbenchProbeResultDTO> | null | undefined,
  correlationId: string,
): WorkbenchProbeResultDTO {
  if (!result) {
    return {
      ok: false,
      error: normalizeUnknownError(new Error('empty probe response'), correlationId),
      events: [],
    }
  }

  const events = Array.isArray(result.events) ? result.events : []
  const error = result.error ? normalizeAppError(result.error, correlationId) : undefined

  return {
    ok: Boolean(result.ok),
    snapshot: result.snapshot,
    error,
    events,
  }
}

function normalizeAppError(error: Partial<AppErrorDTO>, correlationId: string): AppErrorDTO {
  return {
    code: error.code || 'unexpected_error',
    severity: error.severity || 'error',
    retryable: Boolean(error.retryable),
    targetType: error.targetType,
    targetId: error.targetId,
    userMessage: error.userMessage || 'Unexpected Workbench error.',
    technicalDetail: error.technicalDetail,
    recoveryActions: Array.isArray(error.recoveryActions) && error.recoveryActions.length > 0
      ? error.recoveryActions
      : ['Retry the action and inspect the event summary.'],
    correlationId: error.correlationId || correlationId,
  }
}

function isAppErrorDTO(value: unknown): value is AppErrorDTO {
  return Boolean(
    value &&
      typeof value === 'object' &&
      'code' in value &&
      'severity' in value &&
      'userMessage' in value &&
      'correlationId' in value,
  )
}
