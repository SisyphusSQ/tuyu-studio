import { describe, expect, it } from 'vitest'

import { DEFAULT_ALPHA_PROJECT_ROOT } from './projectRoot'
import {
  buildExportGenerationPackageCommand,
  DEFAULT_PACKAGE_CREATED_BY,
  DEFAULT_PROVIDER_PROFILE_ID,
} from './packageExport'

describe('Generation package API command builder', () => {
  it('builds package export commands with normalized root and defaults', () => {
    expect(buildExportGenerationPackageCommand({
      shotId: ' shot_002 ',
    }, 'corr-package')).toEqual({
      root: DEFAULT_ALPHA_PROJECT_ROOT,
      shotId: 'shot_002',
      providerProfileId: DEFAULT_PROVIDER_PROFILE_ID,
      createdBy: DEFAULT_PACKAGE_CREATED_BY,
      correlationId: 'corr-package',
    })
  })

  it('preserves explicit package export owner and provider profile', () => {
    expect(buildExportGenerationPackageCommand({
      root: ' /tmp/tuyu-project ',
      shotId: ' shot_003 ',
      providerProfileId: ' provider_x ',
      createdBy: ' reviewer ',
    }, 'corr-explicit')).toEqual({
      root: '/tmp/tuyu-project',
      shotId: 'shot_003',
      providerProfileId: 'provider_x',
      createdBy: 'reviewer',
      correlationId: 'corr-explicit',
    })
  })
})
