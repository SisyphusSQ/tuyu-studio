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

export interface SaveScriptDocumentCommandDTO {
  root: string
  scriptId?: string
  title: string
  sourceAssetId?: string
  rawText: string
  logline?: string
  synopsis?: string
  correlationId: string
}

export interface LoadScriptDocumentCommandDTO {
  root: string
  scriptId?: string
  correlationId: string
}

export interface ConfirmScriptSceneCommandDTO {
  root: string
  scriptId?: string
  sceneId?: string
  title: string
  location: string
  timeOfDay?: string
  characters: string[]
  props: string[]
  action: string
  dialogue: DialogueLineDTO[]
  emotionalBeat?: string
  sourceRange: ScriptSourceRangeDTO
  allowOverlap: boolean
  correlationId: string
}

export interface SaveShotCandidateCommandDTO {
  root: string
  scriptId?: string
  candidateId?: string
  scriptSceneId: string
  index: number
  durationSeconds: number
  visualDescription: string
  characterRefs: ShotCharacterRefDTO[]
  sourceRange: ScriptSourceRangeDTO
  correlationId: string
}

export interface ListShotCandidatesCommandDTO {
  root: string
  scriptId?: string
  correlationId: string
}

export interface ConfirmShotCandidateCommandDTO {
  root: string
  scriptId?: string
  candidateId: string
  shotId?: string
  confirmedBy?: string
  correlationId: string
}

export interface RejectShotCandidateCommandDTO {
  root: string
  scriptId?: string
  candidateId: string
  rejectionReason?: string
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

export interface ScriptSourceRangeDTO {
  startLine: number
  endLine: number
}

export interface DialogueLineDTO {
  characterName: string
  text: string
  intent?: string
}

export interface ScriptSceneDTO {
  id: string
  index: number
  title: string
  location: string
  timeOfDay: string
  characters: string[]
  props: string[]
  action: string
  dialogue: DialogueLineDTO[]
  emotionalBeat: string
  sourceRange?: ScriptSourceRangeDTO
}

export interface ScriptDocumentDTO {
  id: string
  projectId: string
  title: string
  sourceAssetId?: string
  rawText: string
  logline?: string
  synopsis?: string
  scenes: ScriptSceneDTO[]
  createdAt: string
  updatedAt: string
}

export interface ScriptDocumentResultDTO {
  ok: boolean
  document?: ScriptDocumentDTO
  error?: AppErrorDTO
  events: ProjectEventDTO[]
}

export interface ShotCharacterRefDTO {
  characterId?: string
  name: string
  description?: string
  referenceAssetId?: string
}

export interface ShotCandidateDTO {
  id: string
  projectId: string
  scriptId: string
  scriptSceneId: string
  index: number
  durationSeconds: number
  visualDescription: string
  characterRefs: ShotCharacterRefDTO[]
  sourceRange: ScriptSourceRangeDTO
  status: 'candidate' | 'accepted' | 'rejected' | string
  shotId?: string
  rejectionReason?: string
  createdAt: string
  updatedAt: string
}

export interface ShotCardDTO {
  id: string
  projectId: string
  sceneId: string
  sceneProfileId?: string
  sceneIdRefs: string[]
  sourceCandidateId: string
  scriptSceneId: string
  index: number
  title: string
  description: string
  durationSeconds: number
  aspectRatio: string
  shotType: string
  cameraMovement: string
  action: string
  emotion: string
  emptySceneReason?: string
  characterIds: string[]
  characterRefs: ShotCharacterRefDTO[]
  propIds: string[]
  referenceAssetIds: string[]
  sourceRange: ScriptSourceRangeDTO
  status: string
  confirmedBy: string
  confirmedAt: string
  overwrittenFields: string[]
  packageIds: string[]
  resultIds: string[]
  continuityRuleIds: string[]
  createdAt: string
  updatedAt: string
}

export interface ShotContextIssueDTO {
  code: string
  severity: ErrorSeverity
  field: string
  referenceId?: string
  userMessage: string
  recoveryActions: string[]
}

export interface ShotReferenceDTO {
  field: string
  kind: string
  referenceId: string
  status: string
  path?: string
}

export interface ShotContextReportDTO {
  shotId: string
  status: string
  canEnterContextReady: boolean
  missingFields: string[]
  blocking: ShotContextIssueDTO[]
  warnings: ShotContextIssueDTO[]
  references: ShotReferenceDTO[]
  checkedAt: string
}

export interface ValidateShotContextCommandDTO {
  root?: string
  shotId: string
  correlationId?: string
}

export interface PromoteShotContextCommandDTO {
  root?: string
  shotId: string
  correlationId?: string
}

export interface MarkShotContextDirtyCommandDTO {
  root?: string
  shotId: string
  reason?: string
  correlationId?: string
}

export interface ShotContextResultDTO {
  ok: boolean
  shot?: ShotCardDTO
  report: ShotContextReportDTO
  error?: AppErrorDTO
  events: ProjectEventDTO[]
}

export interface ScriptSceneCandidateResultDTO {
  ok: boolean
  document?: ScriptDocumentDTO
  candidate?: ShotCandidateDTO
  candidates: ShotCandidateDTO[]
  shot?: ShotCardDTO
  error?: AppErrorDTO
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

export function normalizeScriptDocumentResult(
  result: Partial<ScriptDocumentResultDTO> | null | undefined,
  correlationId: string,
): ScriptDocumentResultDTO {
  if (!result) {
    return {
      ok: false,
      error: normalizeUnknownError(new Error('empty script document response'), correlationId),
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

  return {
    ok: Boolean(result.ok),
    document: result.document ? normalizeScriptDocument(result.document) : undefined,
    error: result.error ? normalizeAppError(result.error, correlationId) : undefined,
    events,
  }
}

export function normalizeScriptSceneCandidateResult(
  result: Partial<ScriptSceneCandidateResultDTO> | null | undefined,
  correlationId: string,
): ScriptSceneCandidateResultDTO {
  if (!result) {
    return {
      ok: false,
      error: normalizeUnknownError(new Error('empty script scene candidate response'), correlationId),
      candidates: [],
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

  return {
    ok: Boolean(result.ok),
    document: result.document ? normalizeScriptDocument(result.document) : undefined,
    candidate: result.candidate ? normalizeShotCandidate(result.candidate) : undefined,
    candidates: Array.isArray(result.candidates)
      ? result.candidates.map((candidate) => normalizeShotCandidate(candidate))
      : [],
    shot: result.shot ? normalizeShotCard(result.shot) : undefined,
    error: result.error ? normalizeAppError(result.error, correlationId) : undefined,
    events,
  }
}

export function normalizeShotContextResult(
  result: Partial<ShotContextResultDTO> | null | undefined,
  correlationId: string,
): ShotContextResultDTO {
  if (!result) {
    return {
      ok: false,
      report: emptyShotContextReport(),
      error: normalizeUnknownError(new Error('empty shot context response'), correlationId),
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

  return {
    ok: Boolean(result.ok),
    shot: result.shot ? normalizeShotCard(result.shot) : undefined,
    report: normalizeShotContextReport(result.report),
    error: result.error ? normalizeAppError(result.error, correlationId) : undefined,
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

function normalizeShotCandidate(candidate: Partial<ShotCandidateDTO>): ShotCandidateDTO {
  return {
    id: candidate.id || '',
    projectId: candidate.projectId || '',
    scriptId: candidate.scriptId || '',
    scriptSceneId: candidate.scriptSceneId || '',
    index: Number(candidate.index || 0),
    durationSeconds: Number(candidate.durationSeconds || 0),
    visualDescription: candidate.visualDescription || '',
    characterRefs: Array.isArray(candidate.characterRefs) ? candidate.characterRefs : [],
    sourceRange: candidate.sourceRange || { startLine: 0, endLine: 0 },
    status: candidate.status || 'candidate',
    shotId: candidate.shotId,
    rejectionReason: candidate.rejectionReason,
    createdAt: candidate.createdAt || '',
    updatedAt: candidate.updatedAt || '',
  }
}

function normalizeShotCard(shot: Partial<ShotCardDTO>): ShotCardDTO {
  return {
    id: shot.id || '',
    projectId: shot.projectId || '',
    sceneId: shot.sceneId || '',
    sceneProfileId: shot.sceneProfileId,
    sceneIdRefs: Array.isArray(shot.sceneIdRefs) ? shot.sceneIdRefs : [],
    sourceCandidateId: shot.sourceCandidateId || '',
    scriptSceneId: shot.scriptSceneId || '',
    index: Number(shot.index || 0),
    title: shot.title || '',
    description: shot.description || '',
    durationSeconds: Number(shot.durationSeconds || 0),
    aspectRatio: shot.aspectRatio || '',
    shotType: shot.shotType || '',
    cameraMovement: shot.cameraMovement || '',
    action: shot.action || '',
    emotion: shot.emotion || '',
    emptySceneReason: shot.emptySceneReason,
    characterIds: Array.isArray(shot.characterIds) ? shot.characterIds : [],
    characterRefs: Array.isArray(shot.characterRefs) ? shot.characterRefs : [],
    propIds: Array.isArray(shot.propIds) ? shot.propIds : [],
    referenceAssetIds: Array.isArray(shot.referenceAssetIds) ? shot.referenceAssetIds : [],
    sourceRange: shot.sourceRange || { startLine: 0, endLine: 0 },
    status: shot.status || '',
    confirmedBy: shot.confirmedBy || '',
    confirmedAt: shot.confirmedAt || '',
    overwrittenFields: Array.isArray(shot.overwrittenFields) ? shot.overwrittenFields : [],
    packageIds: Array.isArray(shot.packageIds) ? shot.packageIds : [],
    resultIds: Array.isArray(shot.resultIds) ? shot.resultIds : [],
    continuityRuleIds: Array.isArray(shot.continuityRuleIds) ? shot.continuityRuleIds : [],
    createdAt: shot.createdAt || '',
    updatedAt: shot.updatedAt || '',
  }
}

function normalizeShotContextReport(report: Partial<ShotContextReportDTO> | undefined): ShotContextReportDTO {
  if (!report) {
    return emptyShotContextReport()
  }
  return {
    shotId: report.shotId || '',
    status: report.status || '',
    canEnterContextReady: Boolean(report.canEnterContextReady),
    missingFields: Array.isArray(report.missingFields) ? report.missingFields : [],
    blocking: Array.isArray(report.blocking) ? report.blocking.map(normalizeShotContextIssue) : [],
    warnings: Array.isArray(report.warnings) ? report.warnings.map(normalizeShotContextIssue) : [],
    references: Array.isArray(report.references)
      ? report.references.map((reference) => ({
        field: reference.field || '',
        kind: reference.kind || '',
        referenceId: reference.referenceId || '',
        status: reference.status || 'missing',
        path: reference.path,
      }))
      : [],
    checkedAt: report.checkedAt || '',
  }
}

function normalizeShotContextIssue(issue: Partial<ShotContextIssueDTO>): ShotContextIssueDTO {
  return {
    code: issue.code || 'shot_context_issue',
    severity: issue.severity || 'warning',
    field: issue.field || '',
    referenceId: issue.referenceId,
    userMessage: issue.userMessage || '',
    recoveryActions: Array.isArray(issue.recoveryActions) ? issue.recoveryActions : [],
  }
}

function emptyShotContextReport(): ShotContextReportDTO {
  return {
    shotId: '',
    status: '',
    canEnterContextReady: false,
    missingFields: [],
    blocking: [],
    warnings: [],
    references: [],
    checkedAt: '',
  }
}

function normalizeScriptDocument(document: Partial<ScriptDocumentDTO>): ScriptDocumentDTO {
  return {
    id: document.id || 'script_main',
    projectId: document.projectId || '',
    title: document.title || 'Untitled Script',
    sourceAssetId: document.sourceAssetId,
    rawText: document.rawText || '',
    logline: document.logline,
    synopsis: document.synopsis,
    scenes: Array.isArray(document.scenes)
      ? document.scenes.map((scene) => ({
        ...scene,
        characters: Array.isArray(scene.characters) ? scene.characters : [],
        props: Array.isArray(scene.props) ? scene.props : [],
        dialogue: Array.isArray(scene.dialogue) ? scene.dialogue : [],
      }))
      : [],
    createdAt: document.createdAt || '',
    updatedAt: document.updatedAt || '',
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
