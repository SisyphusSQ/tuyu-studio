import {
  ProjectAssetImport,
  ProjectAssetsList,
} from '../../wailsjs/go/main/App'
import { project } from '../../wailsjs/go/models'

import {
  createCorrelationId,
  normalizeAssetLibraryResult,
  normalizeUnknownError,
  type AssetDuplicatePolicy,
  type AssetLibraryResultDTO,
  type ImportAssetCommandDTO,
  type ListAssetsCommandDTO,
} from './dto'
import { DEFAULT_ALPHA_PROJECT_ROOT, normalizeProjectRoot } from './projectRoot'

export const DEFAULT_ASSET_IMPORT_SOURCE = 'assets/inputs/script-source.md'
export const DEFAULT_ASSET_IMPORT_ROLE = 'script_source'

export interface ImportAssetInput {
  root?: string
  sourcePath: string
  role?: string
  duplicatePolicy?: AssetDuplicatePolicy
  managedReference?: boolean
}

export async function listAssets(root?: string): Promise<AssetLibraryResultDTO> {
  const correlationId = createCorrelationId('asset-list')

  try {
    const result = await ProjectAssetsList(project.ListAssetsCommand.createFrom(
      buildListAssetsCommand(root, correlationId),
    ))

    return normalizeAssetLibraryResult(result as Partial<AssetLibraryResultDTO>, correlationId)
  } catch (error) {
    return assetTransportFailure('asset.list', error, correlationId)
  }
}

export async function importAsset(input: ImportAssetInput): Promise<AssetLibraryResultDTO> {
  const correlationId = createCorrelationId('asset-import')

  try {
    const result = await ProjectAssetImport(project.ImportAssetCommand.createFrom(
      buildImportAssetCommand(input, correlationId),
    ))

    return normalizeAssetLibraryResult(result as Partial<AssetLibraryResultDTO>, correlationId)
  } catch (error) {
    return assetTransportFailure('asset.import', error, correlationId)
  }
}

export function buildListAssetsCommand(root: string | undefined, correlationId: string): ListAssetsCommandDTO {
  return {
    root: normalizeProjectRoot(root || DEFAULT_ALPHA_PROJECT_ROOT),
    correlationId,
  }
}

export function buildImportAssetCommand(
  input: ImportAssetInput,
  correlationId: string,
): ImportAssetCommandDTO {
  return {
    root: normalizeProjectRoot(input.root || DEFAULT_ALPHA_PROJECT_ROOT),
    sourcePath: input.sourcePath.trim(),
    role: input.role?.trim() || DEFAULT_ASSET_IMPORT_ROLE,
    duplicatePolicy: input.duplicatePolicy || 'cancel',
    managedReference: Boolean(input.managedReference),
    correlationId,
  }
}

function assetTransportFailure(eventType: string, error: unknown, correlationId: string): AssetLibraryResultDTO {
  const appError = normalizeUnknownError(error, correlationId)

  return {
    ok: false,
    assets: [],
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
