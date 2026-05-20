import { describe, expect, it } from 'vitest'

import { DEFAULT_ALPHA_PROJECT_ROOT } from './projectRoot'
import {
  buildMarkShotContextDirtyCommand,
  buildPromoteShotContextCommand,
  buildValidateShotContextCommand,
} from './shotContext'

describe('Shot context API command builders', () => {
  it('builds validate and promote commands with normalized root and shot id', () => {
    expect(buildValidateShotContextCommand({
      shotId: ' shot_001 ',
    }, 'corr-check')).toMatchObject({
      root: DEFAULT_ALPHA_PROJECT_ROOT,
      shotId: 'shot_001',
      correlationId: 'corr-check',
    })

    expect(buildPromoteShotContextCommand({
      shotId: ' shot_001 ',
    }, 'corr-ready')).toMatchObject({
      root: DEFAULT_ALPHA_PROJECT_ROOT,
      shotId: 'shot_001',
      correlationId: 'corr-ready',
    })
  })

  it('builds dirty commands with trimmed reason', () => {
    expect(buildMarkShotContextDirtyCommand({
      shotId: ' shot_001 ',
      reason: ' revised action ',
    }, 'corr-dirty')).toMatchObject({
      root: DEFAULT_ALPHA_PROJECT_ROOT,
      shotId: 'shot_001',
      reason: 'revised action',
      correlationId: 'corr-dirty',
    })
  })
})
