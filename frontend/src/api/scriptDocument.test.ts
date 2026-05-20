import { describe, expect, it } from 'vitest'

import { DEFAULT_ALPHA_PROJECT_ROOT } from './projectRoot'
import {
  DEFAULT_SCRIPT_DOCUMENT_ID,
  DEFAULT_SCRIPT_SOURCE_ASSET_ID,
  buildLoadScriptDocumentCommand,
  buildSaveScriptDocumentCommand,
} from './scriptDocument'

describe('ScriptDocument API command builders', () => {
  it('uses the alpha example project and default script id', () => {
    expect(buildLoadScriptDocumentCommand(undefined, 'corr-load')).toEqual({
      root: DEFAULT_ALPHA_PROJECT_ROOT,
      scriptId: DEFAULT_SCRIPT_DOCUMENT_ID,
      correlationId: 'corr-load',
    })
  })

  it('preserves raw text exactly while trimming metadata fields', () => {
    const rawText = '  SCENE 1\nMina: keep going.\n\n'

    expect(buildSaveScriptDocumentCommand({
      title: '  Night Market  ',
      sourceAssetId: ` ${DEFAULT_SCRIPT_SOURCE_ASSET_ID} `,
      rawText,
      logline: '  A painter chooses a frame.  ',
      synopsis: '  A short script draft.  ',
    }, 'corr-save')).toEqual({
      root: DEFAULT_ALPHA_PROJECT_ROOT,
      scriptId: DEFAULT_SCRIPT_DOCUMENT_ID,
      title: 'Night Market',
      sourceAssetId: DEFAULT_SCRIPT_SOURCE_ASSET_ID,
      rawText,
      logline: 'A painter chooses a frame.',
      synopsis: 'A short script draft.',
      correlationId: 'corr-save',
    })
  })

  it('keeps explicit roots and script ids for later user-selected projects', () => {
    expect(buildSaveScriptDocumentCommand({
      root: ' /tmp/creative-project ',
      scriptId: ' episode_01 ',
      rawText: 'Script\n',
    }, 'corr-explicit')).toMatchObject({
      root: '/tmp/creative-project',
      scriptId: 'episode_01',
      correlationId: 'corr-explicit',
    })
  })
})
