import { describe, expect, it } from 'vitest'

import {
  buildImportAssetCommand,
  buildListAssetsCommand,
  DEFAULT_ASSET_IMPORT_ROLE,
  DEFAULT_ASSET_IMPORT_SOURCE,
} from './assets'
import { normalizeAssetLibraryResult, type AssetLibraryResultDTO } from './dto'
import { DEFAULT_ALPHA_PROJECT_ROOT } from './projectRoot'

describe('Asset API wrapper', () => {
  it('normalizes list commands onto the alpha fixture project root', () => {
    expect(buildListAssetsCommand(undefined, 'asset-list-corr')).toEqual({
      root: DEFAULT_ALPHA_PROJECT_ROOT,
      correlationId: 'asset-list-corr',
    })
  })

  it('builds import commands with explicit duplicate policy and managed reference flag', () => {
    expect(buildImportAssetCommand({
      sourcePath: ` ${DEFAULT_ASSET_IMPORT_SOURCE} `,
      role: '',
      duplicatePolicy: 'reuse',
      managedReference: true,
    }, 'asset-import-corr')).toEqual({
      root: DEFAULT_ALPHA_PROJECT_ROOT,
      sourcePath: DEFAULT_ASSET_IMPORT_SOURCE,
      role: DEFAULT_ASSET_IMPORT_ROLE,
      duplicatePolicy: 'reuse',
      managedReference: true,
      correlationId: 'asset-import-corr',
    })
  })

  it('normalizes asset list rows and duplicate errors from the service', () => {
    const result: AssetLibraryResultDTO = {
      ok: false,
      duplicate: {
        existingAssetId: 'asset_text_abc123',
        digest: 'abc123',
        mimeType: 'text/plain',
        policy: 'cancel',
      },
      error: {
        code: 'asset_duplicate_found',
        severity: 'warning',
        retryable: false,
        targetType: 'asset',
        userMessage: 'Duplicate',
        recoveryActions: ['Choose reuse or copy'],
        correlationId: 'service-corr',
      },
      assets: [
        {
          id: 'asset_text_abc123',
          projectId: 'proj_alpha_fixture',
          type: 'text',
          role: 'script_source',
          relativePath: 'assets/inputs/script-source-asset_text_abc123.md',
          originalName: 'script-source.md',
          mimeType: 'text/plain',
          sizeBytes: 120,
          digest: 'abc1234567890',
          source: { kind: 'imported_file' },
          bindings: [],
          thumbnailStatus: 'placeholder',
          bindingCount: 0,
          digestSummary: '',
          missing: false,
          createdAt: '2026-05-20T09:00:00Z',
          updatedAt: '2026-05-20T09:00:00Z',
        },
      ],
      events: [],
    }

    const normalized = normalizeAssetLibraryResult(result, 'ui-corr')

    expect(normalized.ok).toBe(false)
    expect(normalized.error?.code).toBe('asset_duplicate_found')
    expect(normalized.error?.correlationId).toBe('service-corr')
    expect(normalized.duplicate?.existingAssetId).toBe('asset_text_abc123')
    expect(normalized.assets[0].digestSummary).toBe('abc123456789')
    expect(normalized.assets[0].bindingCount).toBe(0)
  })
})
