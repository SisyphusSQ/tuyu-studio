import { describe, expect, it } from 'vitest'

import { normalizeProbeResult, normalizeUnknownError, type WorkbenchProbeResultDTO } from './dto'

describe('AppErrorDTO normalization', () => {
  it('converts transport errors into retryable AppErrorDTO values', () => {
    const error = normalizeUnknownError(new Error('bridge unavailable'), 'corr-test')

    expect(error.code).toBe('transport_error')
    expect(error.severity).toBe('error')
    expect(error.retryable).toBe(true)
    expect(error.correlationId).toBe('corr-test')
    expect(error.recoveryActions.length).toBeGreaterThan(0)
  })

  it('preserves structured service errors from probe results', () => {
    const result: WorkbenchProbeResultDTO = {
      ok: false,
      error: {
        code: 'workbench_probe_blocked',
        severity: 'blocking',
        retryable: false,
        userMessage: 'Blocked',
        recoveryActions: ['Run ready probe'],
        correlationId: 'service-correlation',
      },
      events: [],
    }

    const normalized = normalizeProbeResult(result, 'ui-correlation')

    expect(normalized.ok).toBe(false)
    expect(normalized.error?.code).toBe('workbench_probe_blocked')
    expect(normalized.error?.correlationId).toBe('service-correlation')
  })
})
