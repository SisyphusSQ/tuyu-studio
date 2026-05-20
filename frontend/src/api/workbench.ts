import { WorkbenchProbe } from '../../wailsjs/go/main/App'

import {
  createCorrelationId,
  normalizeProbeResult,
  normalizeUnknownError,
  type WorkbenchProbeMode,
  type WorkbenchProbeResultDTO,
} from './dto'

export async function runWorkbenchProbe(mode: WorkbenchProbeMode): Promise<WorkbenchProbeResultDTO> {
  const correlationId = createCorrelationId(`workbench-${mode}`)

  try {
    const result = await WorkbenchProbe({
      mode,
      correlationId,
    })

    return normalizeProbeResult(result as Partial<WorkbenchProbeResultDTO>, correlationId)
  } catch (error) {
    const appError = normalizeUnknownError(error, correlationId)

    return {
      ok: false,
      error: appError,
      events: [
        {
          eventId: `${correlationId}-transport-error`,
          eventType: 'workbench.probe',
          state: 'failed',
          progress: 0,
          summary: appError.userMessage,
          error: appError,
          nextActions: appError.recoveryActions,
          createdAt: new Date().toISOString(),
        },
      ],
    }
  }
}
