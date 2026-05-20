import { describe, expect, it } from 'vitest'

import {
  buildBindAssetCommand,
  buildListAssetBindingsCommand,
  buildSetMainReferenceCommand,
  DEFAULT_BINDING_PURPOSE,
  DEFAULT_BINDING_TARGET_ID,
  DEFAULT_BINDING_TARGET_TYPE,
} from './assetBindings'
import { normalizeContinuityLibraryResult, type ContinuityLibraryResultDTO } from './dto'
import { DEFAULT_ALPHA_PROJECT_ROOT } from './projectRoot'

describe('Asset binding API wrapper', () => {
  it('normalizes list commands onto the alpha fixture project root', () => {
    expect(buildListAssetBindingsCommand(undefined, 'binding-list-corr')).toEqual({
      root: DEFAULT_ALPHA_PROJECT_ROOT,
      correlationId: 'binding-list-corr',
    })
  })

  it('builds bind commands with stable defaults and trimmed identifiers', () => {
    expect(buildBindAssetCommand({
      assetId: ' asset_char_mina ',
      targetId: ` ${DEFAULT_BINDING_TARGET_ID} `,
      purpose: '',
    }, 'binding-bind-corr')).toEqual({
      root: DEFAULT_ALPHA_PROJECT_ROOT,
      assetId: 'asset_char_mina',
      targetType: DEFAULT_BINDING_TARGET_TYPE,
      targetId: DEFAULT_BINDING_TARGET_ID,
      purpose: DEFAULT_BINDING_PURPOSE,
      duplicatePolicy: 'cancel',
      createdBy: 'workbench_user',
      correlationId: 'binding-bind-corr',
    })
  })

  it('builds clear-main-reference commands without an asset id', () => {
    expect(buildSetMainReferenceCommand({
      assetId: 'asset_char_mina',
      targetType: 'character',
      targetId: 'char_mina',
      clear: true,
      createdBy: ' reviewer ',
    }, 'main-ref-corr')).toEqual({
      root: DEFAULT_ALPHA_PROJECT_ROOT,
      targetType: 'character',
      targetId: 'char_mina',
      assetId: undefined,
      clear: true,
      createdBy: 'reviewer',
      correlationId: 'main-ref-corr',
    })
  })

  it('normalizes continuity results with profiles, lineage, and duplicate binding metadata', () => {
    const result: ContinuityLibraryResultDTO = {
      ok: false,
      assets: [
        {
          id: 'asset_char_mina',
          projectId: 'proj_alpha_fixture',
          type: 'text',
          role: 'character_ref',
          relativePath: 'assets/refs/character-mina.txt',
          originalName: 'character-mina.txt',
          mimeType: 'text/plain',
          sizeBytes: 120,
          digest: 'abc1234567890',
          source: { kind: 'managed_reference' },
          bindings: [
            {
              id: 'bind_asset_char_mina_character_char_mina_reference',
              assetId: 'asset_char_mina',
              targetType: 'character',
              targetId: 'char_mina',
              purpose: 'reference',
              locked: false,
            },
          ],
          thumbnailStatus: 'placeholder',
          digestSummary: '',
          bindingCount: 0,
          missing: false,
          createdAt: '2026-05-20T09:00:00Z',
          updatedAt: '2026-05-20T09:00:00Z',
        },
      ],
      profiles: [
        {
          id: 'char_mina',
          type: 'character',
          name: 'Mina',
          relativePath: 'characters/char-mina.json',
          referenceAssetIds: ['asset_char_mina'],
          mainReferenceAssetId: 'asset_char_mina',
          lockedRules: [],
          continuityRules: [],
          bindings: [],
          bindingCount: 1,
          missingMainReference: false,
        },
      ],
      lineage: [
        {
          assetId: 'asset_char_mina',
          sourceKind: 'managed_reference',
          relativePath: 'assets/refs/character-mina.txt',
          bindings: [],
          targetSummaries: [
            {
              targetType: 'character',
              targetId: 'char_mina',
              name: 'Mina',
              relativePath: 'characters/char-mina.json',
            },
          ],
        },
      ],
      duplicate: {
        id: 'bind_asset_char_mina_character_char_mina_reference',
        assetId: 'asset_char_mina',
        targetType: 'character',
        targetId: 'char_mina',
        purpose: 'reference',
        locked: false,
      },
      error: {
        code: 'binding_duplicate',
        severity: 'warning',
        retryable: false,
        targetType: 'asset_binding',
        userMessage: 'Duplicate',
        recoveryActions: ['Choose reuse'],
        correlationId: 'service-corr',
      },
      events: [],
    }

    const normalized = normalizeContinuityLibraryResult(result, 'ui-corr')

    expect(normalized.error?.code).toBe('binding_duplicate')
    expect(normalized.duplicate?.purpose).toBe('reference')
    expect(normalized.assets[0].bindingCount).toBe(1)
    expect(normalized.profiles[0].mainReferenceAssetId).toBe('asset_char_mina')
    expect(normalized.lineage[0].targetSummaries[0].name).toBe('Mina')
  })
})
