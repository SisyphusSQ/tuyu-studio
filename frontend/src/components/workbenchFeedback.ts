import type {
  AppErrorDTO,
  ErrorSeverity,
  GraphEdgeDTO,
  GraphNodeDTO,
  HealthReportDTO,
  ProjectCanvasDTO,
  ProjectEventDTO,
  RuntimeEventDTO,
} from '../api/dto'

export type FeedbackTone = 'default' | 'processing' | 'success' | 'warning' | 'error'
export type SaveFeedbackState = 'idle' | 'saving' | 'saved' | 'failed'
export type HealthFeedbackStatus = HealthReportDTO['status'] | 'unknown'

export interface StatusBadgeView {
  key: string
  label: string
  tone: FeedbackTone
  detail: string
}

export interface RelationSummaryView {
  id: string
  label: string
  direction: 'incoming' | 'outgoing' | 'selected'
  peerTitle: string
  relation: string
  validity: GraphEdgeDTO['validity']
  tone: FeedbackTone
  detail: string
}

export interface KeyValueView {
  key: string
  label: string
  value: string
}

export interface FeedbackErrorSummary {
  code: string
  message: string
  severity: ErrorSeverity
  retryable: boolean
}

export function nodeStatusBadges(node: GraphNodeDTO | undefined): StatusBadgeView[] {
  if (!node) {
    return []
  }

  const values = dedupeStrings([
    node.status || '',
    ...node.badges,
  ])

  return values.map(statusBadgeView)
}

export function statusBadgeView(value: string): StatusBadgeView {
  const normalized = value.trim()
  const fallback = normalized || 'unknown'

  switch (fallback) {
    case 'missing_ref':
      return badge(fallback, 'Missing ref', 'error', 'Reference cannot be resolved in the project.')
    case 'missing_asset':
      return badge(fallback, 'Missing asset', 'error', 'Referenced asset is missing or unreadable.')
    case 'stale_binding':
      return badge(fallback, 'Stale binding', 'warning', 'Binding target or asset lineage needs review.')
    case 'managed_reference':
      return badge(fallback, 'Managed ref', 'warning', 'Asset points to an explicit managed reference.')
    case 'missing_context':
      return badge(fallback, 'Missing context', 'error', 'Shot context is missing required fields or references.')
    case 'main_reference_missing':
      return badge(fallback, 'Main ref missing', 'warning', 'Profile has no main reference asset selected.')
    case 'main_reference':
      return badge(fallback, 'Main reference', 'success', 'Profile has a visible main reference asset.')
    case 'blocked':
      return badge(fallback, 'Blocked', 'error', 'A blocking validation issue prevents the next action.')
    case 'context_dirty':
    case 'stale':
      return badge(fallback, 'Context dirty', 'warning', 'Upstream context changed and needs review.')
    case 'context_ready':
      return badge(fallback, 'Context ready', 'success', 'Context dependencies are ready.')
    case 'package_ready':
    case 'ready':
      return badge(fallback, 'Package ready', 'success', 'Package or handoff context is ready.')
    case 'submitted':
      return badge(fallback, 'Submitted', 'processing', 'Manual handoff has been submitted.')
    case 'pending':
    case 'pending_review':
      return badge(fallback, 'Pending review', 'processing', 'Result is waiting for review.')
    case 'approved':
      return badge(fallback, 'Approved', 'success', 'Review accepted the result.')
    case 'needs_revision':
      return badge(fallback, 'Needs revision', 'warning', 'Review requested changes.')
    case 'invalid':
      return badge(fallback, 'Invalid', 'error', 'The object failed validation.')
    case 'draft':
      return badge(fallback, 'Draft', 'warning', 'The object is not ready for downstream work.')
    default:
      return badge(fallback, humanizeToken(fallback), 'default', `Status ${fallback}.`)
  }
}

export function relationSummaries(
  canvas: ProjectCanvasDTO | undefined,
  selectedNodeIDs: string[],
): RelationSummaryView[] {
  if (!canvas || selectedNodeIDs.length === 0) {
    return []
  }

  const selected = new Set(selectedNodeIDs)
  const nodes = new Map(canvas.nodes.map((node) => [node.id, node]))

  return canvas.edges
    .filter((edge) => selected.has(edge.sourceNodeId) || selected.has(edge.targetNodeId))
    .map((edge) => {
      const sourceSelected = selected.has(edge.sourceNodeId)
      const targetSelected = selected.has(edge.targetNodeId)
      const direction = sourceSelected && targetSelected
        ? 'selected'
        : sourceSelected
          ? 'outgoing'
          : 'incoming'
      const peerID = sourceSelected ? edge.targetNodeId : edge.sourceNodeId
      const peerTitle = nodes.get(peerID)?.title || peerID
      const validityIssue = edge.validity !== 'valid'
      const detail = edge.error
        ? `${edge.error.code}: ${edge.error.userMessage}`
        : validityIssue
          ? edge.validity
          : edge.createdAt

      return {
        id: edge.id,
        label: `${nodes.get(edge.sourceNodeId)?.title || edge.sourceNodeId} -> ${nodes.get(edge.targetNodeId)?.title || edge.targetNodeId}`,
        direction,
        peerTitle,
        relation: edge.relation,
        validity: edge.validity,
        tone: validityIssue ? 'warning' : 'success',
        detail,
      }
    })
}

export function nodeDataRows(node: GraphNodeDTO | undefined): KeyValueView[] {
  if (!node?.data) {
    return []
  }

  return Object.entries(node.data)
    .filter(([, value]) => value.trim().length > 0)
    .sort(([left], [right]) => left.localeCompare(right))
    .map(([key, value]) => ({
      key,
      label: humanizeToken(key),
      value,
    }))
}

export function continuityRisks(
  node: GraphNodeDTO | undefined,
  relations: RelationSummaryView[],
): StatusBadgeView[] {
  const badges = nodeStatusBadges(node).filter((item) => item.tone === 'warning' || item.tone === 'error')
  const relationIssues = relations
    .filter((relation) => relation.validity !== 'valid')
    .map((relation) => badge(
      relation.id,
      humanizeToken(relation.validity),
      'warning',
      `${relation.relation}: ${relation.detail}`,
    ))

  return [...badges, ...relationIssues]
}

export function summarizeSaveState(state: SaveFeedbackState, savedAt?: string): StatusBadgeView {
  switch (state) {
    case 'saving':
      return badge('saving', 'Saving', 'processing', 'Layout save is in progress.')
    case 'saved':
      return badge('saved', savedAt ? `Saved ${savedAt}` : 'Saved', 'success', 'Latest Canvas layout save completed.')
    case 'failed':
      return badge('failed', 'Save failed', 'error', 'Latest Canvas layout save failed; current input is preserved.')
    default:
      return badge('idle', 'Not saved', 'default', 'No Canvas layout save has completed in this session.')
  }
}

export function highestHealthTone(health: HealthReportDTO | undefined, errors: AppErrorDTO[] = []): FeedbackTone {
  const status = highestHealthStatus([health], errors)

  switch (status) {
    case 'blocking':
      return 'error'
    case 'warning':
      return 'warning'
    case 'clean':
      return 'success'
    default:
      return 'default'
  }
}

export function highestHealthStatus(
  reports: Array<HealthReportDTO | undefined>,
  errors: AppErrorDTO[] = [],
): HealthFeedbackStatus {
  if (
    reports.some((report) => report?.status === 'blocking') ||
    errors.some((error) => error.severity === 'blocking' || error.severity === 'error')
  ) {
    return 'blocking'
  }
  if (reports.some((report) => report?.status === 'warning') || errors.length > 0) {
    return 'warning'
  }
  if (reports.some((report) => report?.status === 'clean')) {
    return 'clean'
  }
  return 'unknown'
}

export function healthTone(status: HealthFeedbackStatus): FeedbackTone {
  switch (status) {
    case 'blocking':
      return 'error'
    case 'warning':
      return 'warning'
    case 'clean':
      return 'success'
    default:
      return 'default'
  }
}

export function firstErrorSummary(
  errors: Array<AppErrorDTO | undefined>,
): FeedbackErrorSummary | undefined {
  const error = errors.find(Boolean)
  if (!error) {
    return undefined
  }

  return {
    code: error.code,
    message: error.userMessage,
    severity: error.severity,
    retryable: error.retryable,
  }
}

export function latestEventSummary(
  graphEvents: ProjectEventDTO[] = [],
  projectEvents: ProjectEventDTO[] = [],
  runtimeEvents: RuntimeEventDTO[] = [],
): string {
  const latest = [...graphEvents, ...projectEvents, ...runtimeEvents]
    .sort((left, right) => Date.parse(right.createdAt) - Date.parse(left.createdAt))[0]

  if (!latest) {
    return 'No events'
  }

  return `${latest.eventType} · ${latest.state}`
}

export function selectionLabel(nodes: GraphNodeDTO[], fallback?: GraphNodeDTO): string {
  if (nodes.length > 1) {
    return `Selection: ${nodes.length} nodes`
  }
  if (nodes.length === 1) {
    return `Selection: ${nodes[0].title}`
  }
  if (fallback) {
    return `Selection: ${fallback.title}`
  }
  return 'Selection: Canvas'
}

function badge(key: string, label: string, tone: FeedbackTone, detail: string): StatusBadgeView {
  return { key, label, tone, detail }
}

function dedupeStrings(values: string[]): string[] {
  const seen = new Set<string>()
  const result: string[] = []
  for (const value of values) {
    const normalized = value.trim()
    if (!normalized || seen.has(normalized)) {
      continue
    }
    seen.add(normalized)
    result.push(normalized)
  }
  return result
}

function humanizeToken(value: string): string {
  return value
    .replace(/[_-]+/g, ' ')
    .replace(/\b\w/g, (letter) => letter.toUpperCase())
}
