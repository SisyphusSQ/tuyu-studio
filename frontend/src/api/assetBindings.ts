import {
  ProjectAssetBind,
  ProjectAssetBindingsList,
  ProjectMainReferenceSet,
} from '../../wailsjs/go/main/App'
import { project } from '../../wailsjs/go/models'

import {
  createCorrelationId,
  normalizeContinuityLibraryResult,
  normalizeUnknownError,
  type AssetBindingDuplicatePolicy,
  type AssetBindingTargetType,
  type BindAssetCommandDTO,
  type ContinuityLibraryResultDTO,
  type ListAssetBindingsCommandDTO,
  type SetMainReferenceCommandDTO,
} from './dto'
import { DEFAULT_ALPHA_PROJECT_ROOT, normalizeProjectRoot } from './projectRoot'

export const DEFAULT_BINDING_TARGET_TYPE: AssetBindingTargetType = 'character'
export const DEFAULT_BINDING_TARGET_ID = 'char_mina'
export const DEFAULT_BINDING_PURPOSE = 'reference'

export interface BindAssetInput {
  root?: string
  assetId: string
  targetType?: AssetBindingTargetType
  targetId: string
  purpose?: string
  duplicatePolicy?: AssetBindingDuplicatePolicy
  createdBy?: string
}

export interface SetMainReferenceInput {
  root?: string
  assetId?: string
  targetType?: AssetBindingTargetType
  targetId: string
  clear?: boolean
  createdBy?: string
}

export async function listAssetBindings(root?: string): Promise<ContinuityLibraryResultDTO> {
  const correlationId = createCorrelationId('asset-binding-list')

  try {
    const result = await ProjectAssetBindingsList(project.ListAssetBindingsCommand.createFrom(
      buildListAssetBindingsCommand(root, correlationId),
    ))

    return normalizeContinuityLibraryResult(result as Partial<ContinuityLibraryResultDTO>, correlationId)
  } catch (error) {
    return continuityTransportFailure('asset.binding.list', error, correlationId)
  }
}

export async function bindAsset(input: BindAssetInput): Promise<ContinuityLibraryResultDTO> {
  const correlationId = createCorrelationId('asset-binding-bind')

  try {
    const result = await ProjectAssetBind(project.BindAssetCommand.createFrom(
      buildBindAssetCommand(input, correlationId),
    ))

    return normalizeContinuityLibraryResult(result as Partial<ContinuityLibraryResultDTO>, correlationId)
  } catch (error) {
    return continuityTransportFailure('asset.binding.bind', error, correlationId)
  }
}

export async function setMainReference(input: SetMainReferenceInput): Promise<ContinuityLibraryResultDTO> {
  const correlationId = createCorrelationId('asset-main-reference')

  try {
    const result = await ProjectMainReferenceSet(project.SetMainReferenceCommand.createFrom(
      buildSetMainReferenceCommand(input, correlationId),
    ))

    return normalizeContinuityLibraryResult(result as Partial<ContinuityLibraryResultDTO>, correlationId)
  } catch (error) {
    return continuityTransportFailure('asset.binding.main_reference', error, correlationId)
  }
}

export function buildListAssetBindingsCommand(
  root: string | undefined,
  correlationId: string,
): ListAssetBindingsCommandDTO {
  return {
    root: normalizeProjectRoot(root || DEFAULT_ALPHA_PROJECT_ROOT),
    correlationId,
  }
}

export function buildBindAssetCommand(input: BindAssetInput, correlationId: string): BindAssetCommandDTO {
  return {
    root: normalizeProjectRoot(input.root || DEFAULT_ALPHA_PROJECT_ROOT),
    assetId: input.assetId.trim(),
    targetType: input.targetType || DEFAULT_BINDING_TARGET_TYPE,
    targetId: input.targetId.trim(),
    purpose: input.purpose?.trim() || DEFAULT_BINDING_PURPOSE,
    duplicatePolicy: input.duplicatePolicy || 'cancel',
    createdBy: input.createdBy?.trim() || 'workbench_user',
    correlationId,
  }
}

export function buildSetMainReferenceCommand(
  input: SetMainReferenceInput,
  correlationId: string,
): SetMainReferenceCommandDTO {
  return {
    root: normalizeProjectRoot(input.root || DEFAULT_ALPHA_PROJECT_ROOT),
    targetType: input.targetType || DEFAULT_BINDING_TARGET_TYPE,
    targetId: input.targetId.trim(),
    assetId: input.clear ? undefined : input.assetId?.trim() || undefined,
    clear: Boolean(input.clear),
    createdBy: input.createdBy?.trim() || 'workbench_user',
    correlationId,
  }
}

function continuityTransportFailure(
  eventType: string,
  error: unknown,
  correlationId: string,
): ContinuityLibraryResultDTO {
  const appError = normalizeUnknownError(error, correlationId)

  return {
    ok: false,
    assets: [],
    profiles: [],
    lineage: [],
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
