import {
  ProjectScriptSceneConfirm,
  ProjectShotCandidateConfirm,
  ProjectShotCandidateReject,
  ProjectShotCandidateSave,
  ProjectShotCandidatesList,
} from '../../wailsjs/go/main/App'
import { project } from '../../wailsjs/go/models'

import {
  createCorrelationId,
  normalizeScriptSceneCandidateResult,
  normalizeUnknownError,
  type ConfirmScriptSceneCommandDTO,
  type ConfirmShotCandidateCommandDTO,
  type ListShotCandidatesCommandDTO,
  type RejectShotCandidateCommandDTO,
  type SaveShotCandidateCommandDTO,
  type ScriptSceneCandidateResultDTO,
  type ScriptSourceRangeDTO,
  type ShotCharacterRefDTO,
  type ShotCandidateDTO,
} from './dto'
import { normalizeProjectRoot } from './projectRoot'
import { DEFAULT_SCRIPT_DOCUMENT_ID } from './scriptDocument'

export interface ConfirmScriptSceneInput {
  root?: string
  scriptId?: string
  sceneId?: string
  title?: string
  location?: string
  timeOfDay?: string
  characters?: string[]
  props?: string[]
  action?: string
  emotionalBeat?: string
  sourceRange: ScriptSourceRangeDTO
  allowOverlap?: boolean
}

export interface SaveShotCandidateInput {
  root?: string
  scriptId?: string
  candidateId?: string
  scriptSceneId: string
  index: number
  durationSeconds: number
  visualDescription?: string
  characterRefs?: ShotCharacterRefDTO[]
  sourceRange: ScriptSourceRangeDTO
}

export interface CandidateActionInput {
  root?: string
  scriptId?: string
  candidateId: string
  shotId?: string
  confirmedBy?: string
  rejectionReason?: string
}

export async function confirmScriptScene(input: ConfirmScriptSceneInput): Promise<ScriptSceneCandidateResultDTO> {
  const correlationId = createCorrelationId('script-scene-confirm')

  try {
    const result = await ProjectScriptSceneConfirm(project.ConfirmScriptSceneCommand.createFrom(
      buildConfirmScriptSceneCommand(input, correlationId),
    ))

    return normalizeScriptSceneCandidateResult(result as Partial<ScriptSceneCandidateResultDTO>, correlationId)
  } catch (error) {
    return transportFailure('script.scene.confirm', error, correlationId)
  }
}

export async function saveShotCandidate(input: SaveShotCandidateInput): Promise<ScriptSceneCandidateResultDTO> {
  const correlationId = createCorrelationId('shot-candidate-save')

  try {
    const result = await ProjectShotCandidateSave(project.SaveShotCandidateCommand.createFrom(
      buildSaveShotCandidateCommand(input, correlationId),
    ))

    return normalizeScriptSceneCandidateResult(result as Partial<ScriptSceneCandidateResultDTO>, correlationId)
  } catch (error) {
    return transportFailure('shot.candidate.save', error, correlationId)
  }
}

export async function listShotCandidates(root?: string, scriptId?: string): Promise<ScriptSceneCandidateResultDTO> {
  const correlationId = createCorrelationId('shot-candidates-list')

  try {
    const result = await ProjectShotCandidatesList(project.ListShotCandidatesCommand.createFrom(
      buildListShotCandidatesCommand(root, scriptId, correlationId),
    ))

    return normalizeScriptSceneCandidateResult(result as Partial<ScriptSceneCandidateResultDTO>, correlationId)
  } catch (error) {
    return transportFailure('shot.candidates.list', error, correlationId)
  }
}

export async function confirmShotCandidate(input: CandidateActionInput): Promise<ScriptSceneCandidateResultDTO> {
  const correlationId = createCorrelationId('shot-candidate-confirm')

  try {
    const result = await ProjectShotCandidateConfirm(project.ConfirmShotCandidateCommand.createFrom(
      buildConfirmShotCandidateCommand(input, correlationId),
    ))

    return normalizeScriptSceneCandidateResult(result as Partial<ScriptSceneCandidateResultDTO>, correlationId)
  } catch (error) {
    return transportFailure('shot.candidate.confirm', error, correlationId)
  }
}

export async function rejectShotCandidate(input: CandidateActionInput): Promise<ScriptSceneCandidateResultDTO> {
  const correlationId = createCorrelationId('shot-candidate-reject')

  try {
    const result = await ProjectShotCandidateReject(project.RejectShotCandidateCommand.createFrom(
      buildRejectShotCandidateCommand(input, correlationId),
    ))

    return normalizeScriptSceneCandidateResult(result as Partial<ScriptSceneCandidateResultDTO>, correlationId)
  } catch (error) {
    return transportFailure('shot.candidate.reject', error, correlationId)
  }
}

export function buildConfirmScriptSceneCommand(
  input: ConfirmScriptSceneInput,
  correlationId: string,
): ConfirmScriptSceneCommandDTO {
  return {
    root: normalizeProjectRoot(input.root),
    scriptId: input.scriptId?.trim() || DEFAULT_SCRIPT_DOCUMENT_ID,
    sceneId: input.sceneId?.trim() || undefined,
    title: input.title?.trim() || '',
    location: input.location?.trim() || '',
    timeOfDay: input.timeOfDay?.trim() || undefined,
    characters: cleanList(input.characters),
    props: cleanList(input.props),
    action: input.action?.trim() || '',
    dialogue: [],
    emotionalBeat: input.emotionalBeat?.trim() || undefined,
    sourceRange: normalizeSourceRange(input.sourceRange),
    allowOverlap: Boolean(input.allowOverlap),
    correlationId,
  }
}

export function buildSaveShotCandidateCommand(
  input: SaveShotCandidateInput,
  correlationId: string,
): SaveShotCandidateCommandDTO {
  return {
    root: normalizeProjectRoot(input.root),
    scriptId: input.scriptId?.trim() || DEFAULT_SCRIPT_DOCUMENT_ID,
    candidateId: input.candidateId?.trim() || undefined,
    scriptSceneId: input.scriptSceneId.trim(),
    index: Number(input.index || 0),
    durationSeconds: Number(input.durationSeconds || 0),
    visualDescription: input.visualDescription?.trim() || '',
    characterRefs: input.characterRefs?.map((ref) => ({
      characterId: ref.characterId?.trim() || undefined,
      name: ref.name.trim(),
      description: ref.description?.trim() || undefined,
      referenceAssetId: ref.referenceAssetId?.trim() || undefined,
    })) || [],
    sourceRange: normalizeSourceRange(input.sourceRange),
    correlationId,
  }
}

export function buildListShotCandidatesCommand(
  root: string | undefined,
  scriptId: string | undefined,
  correlationId: string,
): ListShotCandidatesCommandDTO {
  return {
    root: normalizeProjectRoot(root),
    scriptId: scriptId?.trim() || DEFAULT_SCRIPT_DOCUMENT_ID,
    correlationId,
  }
}

export function buildConfirmShotCandidateCommand(
  input: CandidateActionInput,
  correlationId: string,
): ConfirmShotCandidateCommandDTO {
  return {
    root: normalizeProjectRoot(input.root),
    scriptId: input.scriptId?.trim() || DEFAULT_SCRIPT_DOCUMENT_ID,
    candidateId: input.candidateId.trim(),
    shotId: input.shotId?.trim() || undefined,
    confirmedBy: input.confirmedBy?.trim() || 'local_user',
    correlationId,
  }
}

export function buildRejectShotCandidateCommand(
  input: CandidateActionInput,
  correlationId: string,
): RejectShotCandidateCommandDTO {
  return {
    root: normalizeProjectRoot(input.root),
    scriptId: input.scriptId?.trim() || DEFAULT_SCRIPT_DOCUMENT_ID,
    candidateId: input.candidateId.trim(),
    rejectionReason: input.rejectionReason?.trim() || undefined,
    correlationId,
  }
}

export function nextShotCandidatesAfterResult(
  current: ShotCandidateDTO[],
  result: ScriptSceneCandidateResultDTO,
): ShotCandidateDTO[] {
  if (result.ok || result.candidates.length > 0) {
    return result.candidates
  }
  return current
}

function normalizeSourceRange(sourceRange: ScriptSourceRangeDTO): ScriptSourceRangeDTO {
  return {
    startLine: Number(sourceRange.startLine || 0),
    endLine: Number(sourceRange.endLine || 0),
  }
}

function cleanList(values?: string[]): string[] {
  return values?.map((value) => value.trim()).filter(Boolean) || []
}

function transportFailure(eventType: string, error: unknown, correlationId: string): ScriptSceneCandidateResultDTO {
  const appError = normalizeUnknownError(error, correlationId)

  return {
    ok: false,
    error: appError,
    candidates: [],
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
