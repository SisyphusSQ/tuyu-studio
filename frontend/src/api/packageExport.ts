import { ProjectGenerationPackageExport } from '../../wailsjs/go/main/App'
import { project } from '../../wailsjs/go/models'

import {
  createCorrelationId,
  normalizeGenerationPackageResult,
  normalizeUnknownError,
  type ExportGenerationPackageCommandDTO,
  type GenerationPackageResultDTO,
} from './dto'
import { normalizeProjectRoot } from './projectRoot'

export const DEFAULT_PROVIDER_PROFILE_ID = 'provider_manual_handoff'
export const DEFAULT_PACKAGE_CREATED_BY = 'local_user'

export interface ExportGenerationPackageInput {
  root?: string
  shotId: string
  providerProfileId?: string
  createdBy?: string
}

export async function exportGenerationPackage(
  input: ExportGenerationPackageInput,
): Promise<GenerationPackageResultDTO> {
  const correlationId = createCorrelationId('generation-package-export')

  try {
    const result = await ProjectGenerationPackageExport(project.ExportGenerationPackageCommand.createFrom(
      buildExportGenerationPackageCommand(input, correlationId),
    ))

    return normalizeGenerationPackageResult(result as Partial<GenerationPackageResultDTO>, correlationId)
  } catch (error) {
    const appError = normalizeUnknownError(error, correlationId)

    return {
      ok: false,
      error: appError,
      events: [
        {
          eventId: `${correlationId}-transport-error`,
          eventType: 'generation_package.export',
          state: 'failed',
          summary: appError.userMessage,
          error: appError,
          nextActions: appError.recoveryActions,
          createdAt: new Date().toISOString(),
        },
      ],
    }
  }
}

export function buildExportGenerationPackageCommand(
  input: ExportGenerationPackageInput,
  correlationId: string,
): ExportGenerationPackageCommandDTO {
  return {
    root: normalizeProjectRoot(input.root),
    shotId: input.shotId.trim(),
    providerProfileId: input.providerProfileId?.trim() || DEFAULT_PROVIDER_PROFILE_ID,
    createdBy: input.createdBy?.trim() || DEFAULT_PACKAGE_CREATED_BY,
    correlationId,
  }
}
