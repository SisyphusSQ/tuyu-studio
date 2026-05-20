import {
  ProjectScriptDocumentLoad,
  ProjectScriptDocumentSave,
} from '../../wailsjs/go/main/App'
import { project } from '../../wailsjs/go/models'

import {
  createCorrelationId,
  normalizeScriptDocumentResult,
  normalizeUnknownError,
  type LoadScriptDocumentCommandDTO,
  type SaveScriptDocumentCommandDTO,
  type ScriptDocumentResultDTO,
} from './dto'
import { normalizeProjectRoot } from './projectRoot'

export const DEFAULT_SCRIPT_DOCUMENT_ID = 'script_main'
export const DEFAULT_SCRIPT_SOURCE_ASSET_ID = 'assets/inputs/script-source.md'

export interface SaveScriptDocumentInput {
  root?: string
  scriptId?: string
  title?: string
  sourceAssetId?: string
  rawText?: string
  logline?: string
  synopsis?: string
}

export interface LoadScriptDocumentInput {
  root?: string
  scriptId?: string
}

export async function saveScriptDocument(input: SaveScriptDocumentInput): Promise<ScriptDocumentResultDTO> {
  const correlationId = createCorrelationId('script-document-save')

  try {
    const result = await ProjectScriptDocumentSave(project.SaveScriptDocumentCommand.createFrom(
      buildSaveScriptDocumentCommand(input, correlationId),
    ))

    return normalizeScriptDocumentResult(result as Partial<ScriptDocumentResultDTO>, correlationId)
  } catch (error) {
    return transportFailure('script.document.save', error, correlationId)
  }
}

export async function loadScriptDocument(input?: LoadScriptDocumentInput): Promise<ScriptDocumentResultDTO> {
  const correlationId = createCorrelationId('script-document-load')

  try {
    const result = await ProjectScriptDocumentLoad(project.LoadScriptDocumentCommand.createFrom(
      buildLoadScriptDocumentCommand(input, correlationId),
    ))

    return normalizeScriptDocumentResult(result as Partial<ScriptDocumentResultDTO>, correlationId)
  } catch (error) {
    return transportFailure('script.document.load', error, correlationId)
  }
}

export function buildSaveScriptDocumentCommand(
  input: SaveScriptDocumentInput,
  correlationId: string,
): SaveScriptDocumentCommandDTO {
  return {
    root: normalizeProjectRoot(input.root),
    scriptId: normalizeScriptID(input.scriptId),
    title: input.title?.trim() || 'Untitled Script',
    sourceAssetId: input.sourceAssetId?.trim() || undefined,
    rawText: input.rawText ?? '',
    logline: input.logline?.trim() || undefined,
    synopsis: input.synopsis?.trim() || undefined,
    correlationId,
  }
}

export function buildLoadScriptDocumentCommand(
  input: LoadScriptDocumentInput | undefined,
  correlationId: string,
): LoadScriptDocumentCommandDTO {
  return {
    root: normalizeProjectRoot(input?.root),
    scriptId: normalizeScriptID(input?.scriptId),
    correlationId,
  }
}

function normalizeScriptID(scriptId?: string): string {
  return scriptId?.trim() || DEFAULT_SCRIPT_DOCUMENT_ID
}

function transportFailure(eventType: string, error: unknown, correlationId: string): ScriptDocumentResultDTO {
  const appError = normalizeUnknownError(error, correlationId)

  return {
    ok: false,
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
