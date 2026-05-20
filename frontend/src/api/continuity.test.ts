import { describe, expect, it } from 'vitest'

import {
  buildListContinuityCommand,
  buildSaveContinuityRuleCommand,
  buildUnlockAssetBindingCommand,
  buildUnlockContinuityRuleCommand,
} from './continuity'
import { normalizeContinuityResult, type ContinuityResultDTO } from './dto'
import { DEFAULT_ALPHA_PROJECT_ROOT } from './projectRoot'

describe('Continuity API wrapper', () => {
  it('builds continuity commands with trimmed target and reason fields', () => {
    expect(buildListContinuityCommand(undefined, 'continuity-list-corr')).toEqual({
      root: DEFAULT_ALPHA_PROJECT_ROOT,
      correlationId: 'continuity-list-corr',
    })

    expect(buildSaveContinuityRuleCommand({
      id: ' rule_char_mina_raincoat ',
      targetType: 'character',
      targetId: ' char_mina ',
      rule: ' yellow raincoat stays visible ',
      severity: 'warning',
      locked: false,
      createdBy: ' reviewer ',
    }, 'continuity-save-corr')).toEqual({
      root: DEFAULT_ALPHA_PROJECT_ROOT,
      id: 'rule_char_mina_raincoat',
      targetType: 'character',
      targetId: 'char_mina',
      rule: 'yellow raincoat stays visible',
      severity: 'warning',
      locked: false,
      createdBy: 'reviewer',
      correlationId: 'continuity-save-corr',
    })

    expect(buildUnlockContinuityRuleCommand({
      ruleId: ' rule_char_mina_raincoat ',
      reason: ' approved exception ',
    }, 'continuity-unlock-corr')).toEqual({
      root: DEFAULT_ALPHA_PROJECT_ROOT,
      ruleId: 'rule_char_mina_raincoat',
      reason: 'approved exception',
      correlationId: 'continuity-unlock-corr',
    })

    expect(buildUnlockAssetBindingCommand({
      bindingId: ' binding_asset_ref_character_mina_character_char_mina_main_reference ',
      reason: ' unlock exact binding ',
    }, 'binding-id-unlock-corr')).toEqual({
      root: DEFAULT_ALPHA_PROJECT_ROOT,
      bindingId: 'binding_asset_ref_character_mina_character_char_mina_main_reference',
      assetId: undefined,
      targetType: undefined,
      targetId: undefined,
      purpose: undefined,
      reason: 'unlock exact binding',
      correlationId: 'binding-id-unlock-corr',
    })

    expect(buildUnlockAssetBindingCommand({
      assetId: ' asset_ref_character_mina ',
      targetType: 'character',
      targetId: ' char_mina ',
      purpose: ' main_reference ',
      reason: ' replace main reference ',
    }, 'binding-unlock-corr')).toEqual({
      root: DEFAULT_ALPHA_PROJECT_ROOT,
      bindingId: undefined,
      assetId: 'asset_ref_character_mina',
      targetType: 'character',
      targetId: 'char_mina',
      purpose: 'main_reference',
      reason: 'replace main reference',
      correlationId: 'binding-unlock-corr',
    })
  })

  it('normalizes continuity rules, impact report, health, errors, and events', () => {
    const result: ContinuityResultDTO = {
      ok: true,
      rule: {
        id: 'rule_char_mina_raincoat',
        targetType: 'character',
        targetId: 'char_mina',
        rule: 'Keep the yellow raincoat visible.',
        severity: 'blocking',
        locked: true,
        createdBy: 'user',
      },
      rules: [
        {
          id: 'rule_char_mina_raincoat',
          targetType: 'character',
          targetId: 'char_mina',
          rule: 'Keep the yellow raincoat visible.',
          severity: 'blocking',
          locked: true,
          createdBy: 'user',
        },
      ],
      impact: {
        affectedAssets: ['asset_ref_character_mina'],
        affectedBindings: ['binding_asset_ref_character_mina_character_char_mina_main_reference'],
        affectedProfiles: ['char_mina'],
        affectedShots: ['shot_001'],
        affectedPackages: ['pkg_scene001_shot001'],
        issues: [
          {
            code: 'continuity_reference_missing',
            severity: 'blocking',
            targetType: 'asset',
            targetId: 'asset_ref_character_mina',
            userMessage: 'Reference missing',
            recoveryActions: ['Relink asset'],
          },
        ],
        recoveryActions: ['Review affected Shots'],
        checkedAt: '2026-05-20T10:00:00Z',
      },
      health: {
        status: 'warning',
        checkedAt: '2026-05-20T10:00:00Z',
        items: [],
      },
      events: [
        {
          eventId: 'evt-continuity',
          eventType: 'continuity.rule.saved',
          state: 'completed',
          summary: 'Saved',
          nextActions: undefined as unknown as string[],
          createdAt: '2026-05-20T10:00:00Z',
        },
      ],
    }

    const normalized = normalizeContinuityResult(result, 'ui-corr')

    expect(normalized.rule?.locked).toBe(true)
    expect(normalized.rules[0].severity).toBe('blocking')
    expect(normalized.impact?.affectedShots).toEqual(['shot_001'])
    expect(normalized.impact?.issues[0].recoveryActions).toEqual(['Relink asset'])
    expect(normalized.events[0].nextActions).toEqual([])
  })
})
