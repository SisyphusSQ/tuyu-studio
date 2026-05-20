import {
  ProjectResultImport,
  ProjectResultRebind,
  ProjectResultReviewUpdate,
  ProjectResultsList,
  ProjectResultTrace,
} from '../../wailsjs/go/main/App'
import { project } from '../../wailsjs/go/models'

import {
  createCorrelationId,
  normalizeResultReviewResult,
  normalizeUnknownError,
  type ImportResultCommandDTO,
  type ListResultsCommandDTO,
  type RebindResultCommandDTO,
  type ResultDuplicatePolicy,
  type ResultReviewResultDTO,
  type ResultReviewStatus,
  type TraceResultCommandDTO,
  type UpdateResultReviewCommandDTO,
} from './dto'
import { normalizeProjectRoot } from './projectRoot'

export const DEFAULT_RESULT_SOURCE_PATH = 'assets/results/mock-result-summary.md'
export const DEFAULT_RESULT_SHOT_ID = 'shot_002'
export const DEFAULT_RESULT_REVIEW_STATUS: ResultReviewStatus = 'approved'
export const DEFAULT_RESULT_REBIND_REASON = 'correct result binding'
export const DEFAULT_RESULT_CREATED_BY = 'local_user'

export interface ImportResultInput {
  root?: string
  sourcePath?: string
  shotId?: string
  packageId?: string
  runId?: string
  duplicatePolicy?: ResultDuplicatePolicy
}

export interface UpdateResultReviewInput {
  root?: string
  resultId: string
  reviewStatus: ResultReviewStatus
  reviewNotes?: string
  reason?: string
}

export interface RebindResultInput {
  root?: string
  resultId: string
  shotId?: string
  packageId?: string
  unbindShot?: boolean
  unbindPackage?: boolean
  reason?: string
}

export async function importResult(input: ImportResultInput): Promise<ResultReviewResultDTO> {
  const correlationId = createCorrelationId('result-import')

  try {
    const result = await ProjectResultImport(project.ImportResultCommand.createFrom(
      buildImportResultCommand(input, correlationId),
    ))

    return normalizeResultReviewResult(result as Partial<ResultReviewResultDTO>, correlationId)
  } catch (error) {
    return resultReviewTransportFailure(error, correlationId, 'result.import')
  }
}

export async function listResults(input: Partial<ListResultsCommandDTO> = {}): Promise<ResultReviewResultDTO> {
  const correlationId = createCorrelationId('result-list')

  try {
    const result = await ProjectResultsList(project.ListResultsCommand.createFrom(
      buildListResultsCommand(input, correlationId),
    ))

    return normalizeResultReviewResult(result as Partial<ResultReviewResultDTO>, correlationId)
  } catch (error) {
    return resultReviewTransportFailure(error, correlationId, 'result.list')
  }
}

export async function traceResult(input: TraceResultCommandDTO): Promise<ResultReviewResultDTO> {
  const correlationId = createCorrelationId('result-trace')

  try {
    const result = await ProjectResultTrace(project.TraceResultCommand.createFrom(
      buildTraceResultCommand(input, correlationId),
    ))

    return normalizeResultReviewResult(result as Partial<ResultReviewResultDTO>, correlationId)
  } catch (error) {
    return resultReviewTransportFailure(error, correlationId, 'result.trace')
  }
}

export async function updateResultReview(input: UpdateResultReviewInput): Promise<ResultReviewResultDTO> {
  const correlationId = createCorrelationId('result-review')

  try {
    const result = await ProjectResultReviewUpdate(project.UpdateResultReviewCommand.createFrom(
      buildUpdateResultReviewCommand(input, correlationId),
    ))

    return normalizeResultReviewResult(result as Partial<ResultReviewResultDTO>, correlationId)
  } catch (error) {
    return resultReviewTransportFailure(error, correlationId, 'result.review')
  }
}

export async function rebindResult(input: RebindResultInput): Promise<ResultReviewResultDTO> {
  const correlationId = createCorrelationId('result-rebind')

  try {
    const result = await ProjectResultRebind(project.RebindResultCommand.createFrom(
      buildRebindResultCommand(input, correlationId),
    ))

    return normalizeResultReviewResult(result as Partial<ResultReviewResultDTO>, correlationId)
  } catch (error) {
    return resultReviewTransportFailure(error, correlationId, 'result.rebind')
  }
}

export function buildImportResultCommand(
  input: ImportResultInput,
  correlationId: string,
): ImportResultCommandDTO {
  const runId = input.runId?.trim() || undefined
  const sourcePath = input.sourcePath?.trim() || (runId ? undefined : DEFAULT_RESULT_SOURCE_PATH)

  return {
    root: normalizeProjectRoot(input.root),
    sourcePath,
    shotId: input.shotId?.trim() || undefined,
    packageId: input.packageId?.trim() || undefined,
    runId,
    duplicatePolicy: input.duplicatePolicy || 'cancel',
    createdBy: DEFAULT_RESULT_CREATED_BY,
    correlationId,
  }
}

export function buildListResultsCommand(
  input: Partial<ListResultsCommandDTO>,
  correlationId: string,
): ListResultsCommandDTO {
  return {
    root: normalizeProjectRoot(input.root),
    resultId: input.resultId?.trim() || undefined,
    shotId: input.shotId?.trim() || undefined,
    packageId: input.packageId?.trim() || undefined,
    correlationId,
  }
}

export function buildTraceResultCommand(
  input: TraceResultCommandDTO,
  correlationId: string,
): TraceResultCommandDTO {
  return {
    root: normalizeProjectRoot(input.root),
    resultId: input.resultId.trim(),
    correlationId,
  }
}

export function buildUpdateResultReviewCommand(
  input: UpdateResultReviewInput,
  correlationId: string,
): UpdateResultReviewCommandDTO {
  return {
    root: normalizeProjectRoot(input.root),
    resultId: input.resultId.trim(),
    reviewStatus: input.reviewStatus || DEFAULT_RESULT_REVIEW_STATUS,
    reviewNotes: input.reviewNotes?.trim() || undefined,
    reason: input.reason?.trim() || undefined,
    correlationId,
  }
}

export function buildRebindResultCommand(
  input: RebindResultInput,
  correlationId: string,
): RebindResultCommandDTO {
  return {
    root: normalizeProjectRoot(input.root),
    resultId: input.resultId.trim(),
    shotId: input.shotId?.trim() || undefined,
    packageId: input.packageId?.trim() || undefined,
    unbindShot: Boolean(input.unbindShot),
    unbindPackage: Boolean(input.unbindPackage),
    reason: input.reason?.trim() || DEFAULT_RESULT_REBIND_REASON,
    correlationId,
  }
}

function resultReviewTransportFailure(
  error: unknown,
  correlationId: string,
  eventType: string,
): ResultReviewResultDTO {
  const appError = normalizeUnknownError(error, correlationId)

  return {
    ok: false,
    results: [],
    error: appError,
    events: [
      {
        eventId: `${correlationId}-transport-error`,
        eventType,
        state: 'failed',
        summary: appError.userMessage,
        error: appError,
        nextActions: appError.recoveryActions,
        createdAt: new Date().toISOString(),
      },
    ],
  }
}
