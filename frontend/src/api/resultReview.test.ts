import { describe, expect, it } from 'vitest'

import { DEFAULT_ALPHA_PROJECT_ROOT } from './projectRoot'
import {
  buildImportResultCommand,
  buildRebindResultCommand,
  buildTraceResultCommand,
  buildUpdateResultReviewCommand,
  DEFAULT_RESULT_CREATED_BY,
  DEFAULT_RESULT_REBIND_REASON,
  DEFAULT_RESULT_SOURCE_PATH,
} from './resultReview'

describe('Result review API command builder', () => {
  it('builds import commands with normalized root and duplicate policy', () => {
    expect(buildImportResultCommand({
      shotId: ' shot_002 ',
      packageId: ' pkg_scene001_shot001 ',
      duplicatePolicy: 'new_take',
    }, 'corr-import')).toEqual({
      root: DEFAULT_ALPHA_PROJECT_ROOT,
      sourcePath: DEFAULT_RESULT_SOURCE_PATH,
      shotId: 'shot_002',
      packageId: 'pkg_scene001_shot001',
      runId: undefined,
      duplicatePolicy: 'new_take',
      createdBy: DEFAULT_RESULT_CREATED_BY,
      correlationId: 'corr-import',
    })
  })

  it('can import directly from a mock run output without forcing a source path override', () => {
    expect(buildImportResultCommand({
      root: ' /tmp/tuyu-project ',
      sourcePath: ' ',
      runId: ' run_mock_shot_002_001 ',
    }, 'corr-run')).toEqual({
      root: '/tmp/tuyu-project',
      sourcePath: undefined,
      shotId: undefined,
      packageId: undefined,
      runId: 'run_mock_shot_002_001',
      duplicatePolicy: 'cancel',
      createdBy: DEFAULT_RESULT_CREATED_BY,
      correlationId: 'corr-run',
    })
  })

  it('builds review update commands with trimmed reason and notes', () => {
    expect(buildUpdateResultReviewCommand({
      resultId: ' result_001 ',
      reviewStatus: 'needs_revision',
      reason: ' continuity mismatch ',
      reviewNotes: ' redo lantern ',
    }, 'corr-review')).toEqual({
      root: DEFAULT_ALPHA_PROJECT_ROOT,
      resultId: 'result_001',
      reviewStatus: 'needs_revision',
      reviewNotes: 'redo lantern',
      reason: 'continuity mismatch',
      correlationId: 'corr-review',
    })
  })

  it('builds trace and rebind commands with recovery reason defaults', () => {
    expect(buildTraceResultCommand({
      resultId: ' result_001 ',
    }, 'corr-trace')).toEqual({
      root: DEFAULT_ALPHA_PROJECT_ROOT,
      resultId: 'result_001',
      correlationId: 'corr-trace',
    })

    expect(buildRebindResultCommand({
      resultId: ' result_001 ',
      shotId: ' shot_003 ',
      unbindPackage: true,
    }, 'corr-rebind')).toEqual({
      root: DEFAULT_ALPHA_PROJECT_ROOT,
      resultId: 'result_001',
      shotId: 'shot_003',
      packageId: undefined,
      unbindShot: false,
      unbindPackage: true,
      reason: DEFAULT_RESULT_REBIND_REASON,
      correlationId: 'corr-rebind',
    })
  })
})
