export type ErrorSeverity = 'info' | 'warning' | 'error' | 'blocking'
export type RuntimeEventState = 'queued' | 'running' | 'blocked' | 'completed' | 'failed' | 'cancelled'
export type WorkbenchProbeMode = 'status' | 'structured_error'
export type ProjectOperationName = 'create' | 'open' | 'save' | 'health'
export type CanvasTheme = 'dark' | 'warm_light'
export type CanvasNodeCategory = 'source' | 'concept' | 'continuity' | 'production' | 'output' | 'review' | 'organizer' | 'handoff'
export type NodeSource = 'human' | 'agent' | 'system' | 'imported'
export type GraphEdgeValidity = 'valid' | 'invalid_relation' | 'missing_endpoint'

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

export interface ProjectSummaryDTO {
  projectId: string
  name: string
  type: string
  schemaVersion: string
  rootName: string
  openMode: string
  lockState: string
  lastCleanShutdown: boolean
  graphVersion: number
  updatedAt: string
  capabilities: string[]
}

export interface HealthItemDTO {
  severity: ErrorSeverity
  code: string
  path?: string
  affectedObjects: string[]
  userMessage: string
  technicalDetail?: string
  recoveryActions: string[]
}

export interface HealthReportDTO {
  status: 'clean' | 'warning' | 'blocking'
  checkedAt: string
  items: HealthItemDTO[]
}

export interface ProjectEventDTO {
  eventId: string
  eventType: string
  state: RuntimeEventState
  summary: string
  error?: AppErrorDTO
  nextActions: string[]
  createdAt: string
}

export interface ProjectOperationResultDTO {
  ok: boolean
  summary?: ProjectSummaryDTO
  health?: HealthReportDTO
  error?: AppErrorDTO
  events: ProjectEventDTO[]
}

export interface ProjectGraphViewCommandDTO {
  root: string
  expectedGraphVersion?: number
  correlationId: string
}

export interface ProjectGraphNodeLayoutCommandDTO {
  id: string
  position: CanvasPositionDTO
  size?: CanvasSizeDTO
  collapsed: boolean
}

export interface ProjectGraphLayoutSaveCommandDTO {
  root: string
  expectedGraphVersion: number
  viewport: CanvasViewportDTO
  theme: CanvasTheme
  grid: CanvasGridDTO
  nodes: ProjectGraphNodeLayoutCommandDTO[]
  correlationId: string
}

export interface CanvasViewportDTO {
  x: number
  y: number
  zoom: number
}

export interface CanvasGridDTO {
  visible: boolean
  size: number
  opacity: number
}

export interface CanvasPositionDTO {
  x: number
  y: number
}

export interface CanvasSizeDTO {
  width: number
  height: number
}

export interface CanvasLayoutDTO {
  x: number
  y: number
  width: number
  height: number
  collapsed: boolean
}

export interface GraphNodeDTO {
  id: string
  kind: string
  category: CanvasNodeCategory
  title: string
  refId?: string
  position: CanvasPositionDTO
  size: CanvasSizeDTO
  collapsed: boolean
  status?: string
  badges: string[]
  source: NodeSource
  sourceEventId?: string
  data?: Record<string, string>
}

export interface GraphEdgeDTO {
  id: string
  sourceNodeId: string
  targetNodeId: string
  relation: string
  label: string
  createdAt: string
  validity: GraphEdgeValidity
  error?: AppErrorDTO
}

export interface ReferenceGroupDTO {
  id: string
  title: string
  role: string
  inputNodeIds: string[]
  priority: number
  notes?: string
  layout: CanvasLayoutDTO
}

export interface FrameHistorySummaryDTO {
  currentRunId?: string
  favoriteRunIds: string[]
  latestSuccessfulRunId?: string
}

export interface ProductionFrameDTO {
  id: string
  title: string
  referenceGroupIds: string[]
  outputNodeIds: string[]
  taskIntent: string
  requiredCapabilities: string[]
  historySummary: FrameHistorySummaryDTO
  layout: CanvasLayoutDTO
}

export interface ProjectCanvasDTO {
  id: string
  projectId: string
  schemaVersion: string
  version: number
  viewport: CanvasViewportDTO
  theme: CanvasTheme
  grid: CanvasGridDTO
  nodes: GraphNodeDTO[]
  edges: GraphEdgeDTO[]
  frames: ProductionFrameDTO[]
  referenceGroups: ReferenceGroupDTO[]
  selectedNodeIds: string[]
  updatedAt: string
}

export interface ProjectGraphViewResultDTO {
  ok: boolean
  canvas?: ProjectCanvasDTO
  health?: HealthReportDTO
  error?: AppErrorDTO
  errors: AppErrorDTO[]
  events: ProjectEventDTO[]
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

export function normalizeProjectOperationResult(
  result: Partial<ProjectOperationResultDTO> | null | undefined,
  correlationId: string,
): ProjectOperationResultDTO {
  if (!result) {
    return {
      ok: false,
      error: normalizeUnknownError(new Error('empty project operation response'), correlationId),
      events: [],
    }
  }

  const events = Array.isArray(result.events)
    ? result.events.map((event) => ({
      ...event,
      error: event.error ? normalizeAppError(event.error, correlationId) : undefined,
      nextActions: Array.isArray(event.nextActions) ? event.nextActions : [],
    }))
    : []
  const error = result.error ? normalizeAppError(result.error, correlationId) : undefined
  const health = result.health
    ? {
      ...result.health,
      items: Array.isArray(result.health.items) ? result.health.items : [],
    }
    : undefined

  return {
    ok: Boolean(result.ok),
    summary: result.summary,
    health,
    error,
    events,
  }
}

export function normalizeProjectGraphViewResult(
  result: Partial<ProjectGraphViewResultDTO> | null | undefined,
  correlationId: string,
): ProjectGraphViewResultDTO {
  if (!result) {
    return {
      ok: false,
      error: normalizeUnknownError(new Error('empty graph view response'), correlationId),
      errors: [],
      events: [],
    }
  }

  const errors = Array.isArray(result.errors)
    ? result.errors.map((error) => normalizeAppError(error, correlationId))
    : []
  const events = Array.isArray(result.events)
    ? result.events.map((event) => ({
      ...event,
      error: event.error ? normalizeAppError(event.error, correlationId) : undefined,
      nextActions: Array.isArray(event.nextActions) ? event.nextActions : [],
    }))
    : []
  const health = result.health
    ? {
      ...result.health,
      items: Array.isArray(result.health.items) ? result.health.items : [],
    }
    : undefined

  return {
    ok: Boolean(result.ok),
    canvas: result.canvas ? normalizeProjectCanvas(result.canvas) : undefined,
    health,
    error: result.error ? normalizeAppError(result.error, correlationId) : undefined,
    errors,
    events,
  }
}

function normalizeProjectCanvas(canvas: ProjectCanvasDTO): ProjectCanvasDTO {
  return {
    ...canvas,
    nodes: Array.isArray(canvas.nodes)
      ? canvas.nodes.map((node) => ({
        ...node,
        badges: Array.isArray(node.badges) ? node.badges : [],
        data: node.data || {},
      }))
      : [],
    edges: Array.isArray(canvas.edges)
      ? canvas.edges.map((edge) => ({
        ...edge,
        error: edge.error ? normalizeAppError(edge.error, edge.error.correlationId) : undefined,
      }))
      : [],
    frames: Array.isArray(canvas.frames)
      ? canvas.frames.map((frame) => ({
        ...frame,
        referenceGroupIds: Array.isArray(frame.referenceGroupIds) ? frame.referenceGroupIds : [],
        outputNodeIds: Array.isArray(frame.outputNodeIds) ? frame.outputNodeIds : [],
        requiredCapabilities: Array.isArray(frame.requiredCapabilities) ? frame.requiredCapabilities : [],
        historySummary: {
          ...frame.historySummary,
          favoriteRunIds: Array.isArray(frame.historySummary?.favoriteRunIds)
            ? frame.historySummary.favoriteRunIds
            : [],
        },
      }))
      : [],
    referenceGroups: Array.isArray(canvas.referenceGroups)
      ? canvas.referenceGroups.map((group) => ({
        ...group,
        inputNodeIds: Array.isArray(group.inputNodeIds) ? group.inputNodeIds : [],
      }))
      : [],
    selectedNodeIds: Array.isArray(canvas.selectedNodeIds) ? canvas.selectedNodeIds : [],
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
