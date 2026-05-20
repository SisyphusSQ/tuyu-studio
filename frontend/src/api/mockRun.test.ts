import { describe, expect, it } from 'vitest'

import { DEFAULT_ALPHA_PROJECT_ROOT } from './projectRoot'
import {
  buildCancelMockRunCommand,
  buildRetryMockRunCommand,
  buildStartMockRunCommand,
  DEFAULT_MOCK_RUN_CREATED_BY,
  DEFAULT_MOCK_RUN_TASK_MODE,
} from './mockRun'

describe('Mock run API command builder', () => {
  it('builds start commands with normalized root and mock_local mode', () => {
    expect(buildStartMockRunCommand({
      shotId: ' shot_002 ',
      selectionIds: [' node_shot_002 ', ''],
    }, 'corr-start')).toEqual({
      root: DEFAULT_ALPHA_PROJECT_ROOT,
      runId: undefined,
      shotId: 'shot_002',
      packageId: undefined,
      selectionIds: ['node_shot_002'],
      taskMode: DEFAULT_MOCK_RUN_TASK_MODE,
      retryOfRunId: undefined,
      cancelReason: undefined,
      createdBy: DEFAULT_MOCK_RUN_CREATED_BY,
      correlationId: 'corr-start',
    })
  })

  it('builds cancel commands with a deterministic local cancel reason', () => {
    expect(buildCancelMockRunCommand({
      root: ' /tmp/tuyu-project ',
      runId: ' run_mock_shot_002_001 ',
      shotId: ' shot_002 ',
    }, 'corr-cancel')).toEqual({
      root: '/tmp/tuyu-project',
      runId: 'run_mock_shot_002_001',
      shotId: 'shot_002',
      packageId: undefined,
      selectionIds: [],
      taskMode: DEFAULT_MOCK_RUN_TASK_MODE,
      retryOfRunId: undefined,
      cancelReason: 'user_cancelled',
      createdBy: DEFAULT_MOCK_RUN_CREATED_BY,
      correlationId: 'corr-cancel',
    })
  })

  it('builds retry commands without overwriting the target run id', () => {
    expect(buildRetryMockRunCommand({
      runId: ' run_mock_shot_002_001 ',
      packageId: ' pkg_001 ',
    }, 'corr-retry')).toMatchObject({
      root: DEFAULT_ALPHA_PROJECT_ROOT,
      runId: 'run_mock_shot_002_001',
      packageId: 'pkg_001',
      retryOfRunId: 'run_mock_shot_002_001',
      taskMode: DEFAULT_MOCK_RUN_TASK_MODE,
      correlationId: 'corr-retry',
    })
  })
})
