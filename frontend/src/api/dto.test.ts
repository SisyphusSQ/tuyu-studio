import { describe, expect, it } from 'vitest'

import {
  normalizeProbeResult,
  normalizeProjectOperationResult,
  normalizeUnknownError,
  type ProjectOperationResultDTO,
  type WorkbenchProbeResultDTO,
} from './dto'

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

  it('normalizes project operation health and events', () => {
    const result: ProjectOperationResultDTO = {
      ok: true,
      health: {
        status: 'warning',
        checkedAt: '2026-05-20T05:30:00Z',
        items: [
          {
            severity: 'warning',
            code: 'project_stale_lock',
            path: 'locks/session.lock',
            affectedObjects: [],
            userMessage: 'Project lock is stale.',
            recoveryActions: ['Confirm takeover.'],
          },
        ],
      },
      events: [
        {
          eventId: 'evt-project-health',
          eventType: 'project.health_checked',
          state: 'completed',
          summary: 'Project health check completed.',
          nextActions: [],
          createdAt: '2026-05-20T05:30:00Z',
        },
      ],
    }

    const normalized = normalizeProjectOperationResult(result, 'project-correlation')

    expect(normalized.ok).toBe(true)
    expect(normalized.health?.status).toBe('warning')
    expect(normalized.health?.items[0].code).toBe('project_stale_lock')
    expect(normalized.events[0].eventType).toBe('project.health_checked')
  })
})
