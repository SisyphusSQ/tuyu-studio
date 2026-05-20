import {
  ProjectAssetBindingUnlock,
  ProjectContinuityList,
  ProjectContinuityRuleSave,
  ProjectContinuityRuleUnlock,
} from '../../wailsjs/go/main/App'
import { project } from '../../wailsjs/go/models'

import {
  createCorrelationId,
  normalizeContinuityResult,
  normalizeUnknownError,
  type AssetBindingTargetType,
  type ContinuityResultDTO,
  type ContinuityRuleSeverity,
  type ListContinuityCommandDTO,
  type SaveContinuityRuleCommandDTO,
  type UnlockAssetBindingCommandDTO,
  type UnlockContinuityRuleCommandDTO,
} from './dto'
import { DEFAULT_ALPHA_PROJECT_ROOT, normalizeProjectRoot } from './projectRoot'

export interface SaveContinuityRuleInput {
  root?: string
  id?: string
  targetType?: AssetBindingTargetType
  targetId: string
  rule: string
  severity?: ContinuityRuleSeverity
  locked?: boolean
  createdBy?: string
}

export interface UnlockContinuityRuleInput {
  root?: string
  ruleId: string
  reason: string
}

export interface UnlockAssetBindingInput {
  root?: string
  bindingId?: string
  assetId?: string
  targetType?: AssetBindingTargetType
  targetId?: string
  purpose?: string
  reason: string
}

export async function listContinuity(root?: string): Promise<ContinuityResultDTO> {
  const correlationId = createCorrelationId('continuity-list')

  try {
    const result = await ProjectContinuityList(project.ListContinuityCommand.createFrom(
      buildListContinuityCommand(root, correlationId),
    ))

    return normalizeContinuityResult(result as Partial<ContinuityResultDTO>, correlationId)
  } catch (error) {
    return continuityTransportFailure('continuity.list', error, correlationId)
  }
}

export async function saveContinuityRule(input: SaveContinuityRuleInput): Promise<ContinuityResultDTO> {
  const correlationId = createCorrelationId('continuity-rule-save')

  try {
    const result = await ProjectContinuityRuleSave(project.SaveContinuityRuleCommand.createFrom(
      buildSaveContinuityRuleCommand(input, correlationId),
    ))

    return normalizeContinuityResult(result as Partial<ContinuityResultDTO>, correlationId)
  } catch (error) {
    return continuityTransportFailure('continuity.rule.save', error, correlationId)
  }
}

export async function unlockContinuityRule(input: UnlockContinuityRuleInput): Promise<ContinuityResultDTO> {
  const correlationId = createCorrelationId('continuity-rule-unlock')

  try {
    const result = await ProjectContinuityRuleUnlock(project.UnlockContinuityRuleCommand.createFrom(
      buildUnlockContinuityRuleCommand(input, correlationId),
    ))

    return normalizeContinuityResult(result as Partial<ContinuityResultDTO>, correlationId)
  } catch (error) {
    return continuityTransportFailure('continuity.rule.unlock', error, correlationId)
  }
}

export async function unlockAssetBinding(input: UnlockAssetBindingInput): Promise<ContinuityResultDTO> {
  const correlationId = createCorrelationId('asset-binding-unlock')

  try {
    const result = await ProjectAssetBindingUnlock(project.UnlockAssetBindingCommand.createFrom(
      buildUnlockAssetBindingCommand(input, correlationId),
    ))

    return normalizeContinuityResult(result as Partial<ContinuityResultDTO>, correlationId)
  } catch (error) {
    return continuityTransportFailure('continuity.binding.unlock', error, correlationId)
  }
}

export function buildListContinuityCommand(
  root: string | undefined,
  correlationId: string,
): ListContinuityCommandDTO {
  return {
    root: normalizeProjectRoot(root || DEFAULT_ALPHA_PROJECT_ROOT),
    correlationId,
  }
}

export function buildSaveContinuityRuleCommand(
  input: SaveContinuityRuleInput,
  correlationId: string,
): SaveContinuityRuleCommandDTO {
  return {
    root: normalizeProjectRoot(input.root || DEFAULT_ALPHA_PROJECT_ROOT),
    id: input.id?.trim() || undefined,
    targetType: input.targetType || 'character',
    targetId: input.targetId.trim(),
    rule: input.rule.trim(),
    severity: input.severity || 'blocking',
    locked: input.locked ?? true,
    createdBy: input.createdBy?.trim() || 'workbench_user',
    correlationId,
  }
}

export function buildUnlockContinuityRuleCommand(
  input: UnlockContinuityRuleInput,
  correlationId: string,
): UnlockContinuityRuleCommandDTO {
  return {
    root: normalizeProjectRoot(input.root || DEFAULT_ALPHA_PROJECT_ROOT),
    ruleId: input.ruleId.trim(),
    reason: input.reason.trim(),
    correlationId,
  }
}

export function buildUnlockAssetBindingCommand(
  input: UnlockAssetBindingInput,
  correlationId: string,
): UnlockAssetBindingCommandDTO {
  return {
    root: normalizeProjectRoot(input.root || DEFAULT_ALPHA_PROJECT_ROOT),
    bindingId: input.bindingId?.trim() || undefined,
    assetId: input.assetId?.trim() || undefined,
    targetType: input.targetType,
    targetId: input.targetId?.trim() || undefined,
    purpose: input.purpose?.trim() || undefined,
    reason: input.reason.trim(),
    correlationId,
  }
}

function continuityTransportFailure(
  eventType: string,
  error: unknown,
  correlationId: string,
): ContinuityResultDTO {
  const appError = normalizeUnknownError(error, correlationId)

  return {
    ok: false,
    rules: [],
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
