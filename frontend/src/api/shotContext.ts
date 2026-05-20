import {
  ProjectShotContextMarkDirty,
  ProjectShotContextPromote,
  ProjectShotContextValidate,
} from '../../wailsjs/go/main/App'
import { project } from '../../wailsjs/go/models'

import {
  createCorrelationId,
  normalizeShotContextResult,
  normalizeUnknownError,
  type MarkShotContextDirtyCommandDTO,
  type PromoteShotContextCommandDTO,
  type ShotContextResultDTO,
  type ValidateShotContextCommandDTO,
} from './dto'
import { normalizeProjectRoot } from './projectRoot'

export interface ShotContextInput {
  root?: string
  shotId: string
}

export interface ShotContextDirtyInput extends ShotContextInput {
  reason?: string
}

export async function validateShotContext(input: ShotContextInput): Promise<ShotContextResultDTO> {
  const correlationId = createCorrelationId('shot-context-check')

  try {
    const result = await ProjectShotContextValidate(project.ValidateShotContextCommand.createFrom(
      buildValidateShotContextCommand(input, correlationId),
    ))

    return normalizeShotContextResult(result as Partial<ShotContextResultDTO>, correlationId)
  } catch (error) {
    return transportFailure('shot.context.check', error, correlationId)
  }
}

export async function promoteShotContext(input: ShotContextInput): Promise<ShotContextResultDTO> {
  const correlationId = createCorrelationId('shot-context-ready')

  try {
    const result = await ProjectShotContextPromote(project.PromoteShotContextCommand.createFrom(
      buildPromoteShotContextCommand(input, correlationId),
    ))

    return normalizeShotContextResult(result as Partial<ShotContextResultDTO>, correlationId)
  } catch (error) {
    return transportFailure('shot.context.ready', error, correlationId)
  }
}

export async function markShotContextDirty(input: ShotContextDirtyInput): Promise<ShotContextResultDTO> {
  const correlationId = createCorrelationId('shot-context-dirty')

  try {
    const result = await ProjectShotContextMarkDirty(project.MarkShotContextDirtyCommand.createFrom(
      buildMarkShotContextDirtyCommand(input, correlationId),
    ))

    return normalizeShotContextResult(result as Partial<ShotContextResultDTO>, correlationId)
  } catch (error) {
    return transportFailure('shot.context.dirty', error, correlationId)
  }
}

export function buildValidateShotContextCommand(
  input: ShotContextInput,
  correlationId: string,
): ValidateShotContextCommandDTO {
  return {
    root: normalizeProjectRoot(input.root),
    shotId: input.shotId.trim(),
    correlationId,
  }
}

export function buildPromoteShotContextCommand(
  input: ShotContextInput,
  correlationId: string,
): PromoteShotContextCommandDTO {
  return {
    root: normalizeProjectRoot(input.root),
    shotId: input.shotId.trim(),
    correlationId,
  }
}

export function buildMarkShotContextDirtyCommand(
  input: ShotContextDirtyInput,
  correlationId: string,
): MarkShotContextDirtyCommandDTO {
  return {
    root: normalizeProjectRoot(input.root),
    shotId: input.shotId.trim(),
    reason: input.reason?.trim() || undefined,
    correlationId,
  }
}

function transportFailure(eventType: string, error: unknown, correlationId: string): ShotContextResultDTO {
  const appError = normalizeUnknownError(error, correlationId)

  return {
    ok: false,
    report: {
      shotId: '',
      status: '',
      canEnterContextReady: false,
      missingFields: [],
      blocking: [],
      warnings: [],
      references: [],
      checkedAt: '',
    },
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
