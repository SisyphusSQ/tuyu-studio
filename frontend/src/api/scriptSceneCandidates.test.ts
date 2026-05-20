import { describe, expect, it } from 'vitest'

import { DEFAULT_ALPHA_PROJECT_ROOT } from './projectRoot'
import {
  buildConfirmScriptSceneCommand,
  buildConfirmShotCandidateCommand,
  buildRejectShotCandidateCommand,
  buildSaveShotCandidateCommand,
  nextShotCandidatesAfterResult,
} from './scriptSceneCandidates'

describe('ScriptScene and shot candidate API command builders', () => {
  it('builds a normalized scene confirm command', () => {
    expect(buildConfirmScriptSceneCommand({
      sceneId: ' scene_001 ',
      title: ' Opening ',
      location: ' Laneway ',
      characters: [' Mina ', ''],
      props: [' Lantern '],
      action: ' Enters ',
      sourceRange: { startLine: 2, endLine: 4 },
    }, 'corr-scene')).toMatchObject({
      root: DEFAULT_ALPHA_PROJECT_ROOT,
      scriptId: 'script_main',
      sceneId: 'scene_001',
      title: 'Opening',
      location: 'Laneway',
      characters: ['Mina'],
      props: ['Lantern'],
      action: 'Enters',
      sourceRange: { startLine: 2, endLine: 4 },
      allowOverlap: false,
      correlationId: 'corr-scene',
    })
  })

  it('builds candidate rows with character refs', () => {
    expect(buildSaveShotCandidateCommand({
      candidateId: ' candidate_001 ',
      scriptSceneId: 'scene_001',
      index: 1,
      durationSeconds: 6,
      visualDescription: ' Mina reaches the stall. ',
      characterRefs: [{ characterId: ' char_mina ', name: ' Mina ', description: ' yellow raincoat ' }],
      sourceRange: { startLine: 3, endLine: 3 },
    }, 'corr-candidate')).toMatchObject({
      root: DEFAULT_ALPHA_PROJECT_ROOT,
      candidateId: 'candidate_001',
      scriptSceneId: 'scene_001',
      index: 1,
      durationSeconds: 6,
      visualDescription: 'Mina reaches the stall.',
      characterRefs: [{ characterId: 'char_mina', name: 'Mina', description: 'yellow raincoat' }],
      sourceRange: { startLine: 3, endLine: 3 },
    })
  })

  it('builds confirm and reject commands without changing the default root', () => {
    expect(buildConfirmShotCandidateCommand({
      candidateId: ' candidate_001 ',
      shotId: ' shot_001 ',
      confirmedBy: ' tester ',
    }, 'corr-confirm')).toMatchObject({
      root: DEFAULT_ALPHA_PROJECT_ROOT,
      candidateId: 'candidate_001',
      shotId: 'shot_001',
      confirmedBy: 'tester',
    })

    expect(buildRejectShotCandidateCommand({
      candidateId: ' candidate_001 ',
      rejectionReason: ' needs edit ',
    }, 'corr-reject')).toMatchObject({
      root: DEFAULT_ALPHA_PROJECT_ROOT,
      candidateId: 'candidate_001',
      rejectionReason: 'needs edit',
    })
  })

  it('preserves displayed candidate rows when a failed action returns no rows', () => {
    const current = [{
      id: 'candidate_001',
      projectId: 'proj_alpha_fixture',
      scriptId: 'script_main',
      scriptSceneId: 'scene_001',
      index: 1,
      durationSeconds: 5,
      visualDescription: 'Existing row.',
      characterRefs: [],
      sourceRange: { startLine: 1, endLine: 1 },
      status: 'candidate',
      createdAt: '2026-05-20T00:00:00Z',
      updatedAt: '2026-05-20T00:00:00Z',
    }]

    expect(nextShotCandidatesAfterResult(current, {
      ok: false,
      candidates: [],
      events: [],
    })).toBe(current)

    expect(nextShotCandidatesAfterResult(current, {
      ok: true,
      candidates: [],
      events: [],
    })).toEqual([])
  })
})
