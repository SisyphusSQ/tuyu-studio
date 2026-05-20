import { describe, expect, it } from 'vitest'

import {
  normalizeProjectGraphViewResult,
  normalizeMockRunResult,
  normalizeProbeResult,
  normalizeProjectOperationResult,
  normalizeUnknownError,
  type ProjectGraphViewResultDTO,
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

  it('normalizes graph view DTO collections and relation errors', () => {
    const result: ProjectGraphViewResultDTO = {
      ok: true,
      canvas: {
        id: 'proj_alpha_fixture_canvas',
        projectId: 'proj_alpha_fixture',
        schemaVersion: '1.0.0',
        version: 3,
        viewport: { x: 0, y: 0, zoom: 1 },
        theme: 'dark',
        grid: { visible: true, size: 24, opacity: 0.24 },
        nodes: [
          {
            id: 'node_shot_001',
            kind: 'shot',
            category: 'production',
            title: 'Shot 001',
            refId: 'shots/shot-001.json',
            position: { x: 620, y: 96 },
            size: { width: 240, height: 112 },
            collapsed: false,
            status: 'context_ready',
            badges: ['context_ready'],
            source: 'human',
          },
        ],
        edges: [
          {
            id: 'edge_invalid',
            sourceNodeId: 'node_shot_001',
            targetNodeId: 'node_missing',
            relation: 'uses',
            label: 'uses',
            createdAt: '2026-05-20T05:40:00Z',
            validity: 'missing_endpoint',
            error: {
              code: 'graph_view_missing_endpoint',
              severity: 'warning',
              retryable: true,
              userMessage: 'Missing endpoint',
              recoveryActions: ['Restore node'],
              correlationId: 'service-correlation',
            },
          },
        ],
        frames: [],
        referenceGroups: [],
        selectedNodeIds: [],
        updatedAt: '2026-05-20T05:40:00Z',
      },
      errors: [],
      events: [],
    }

    const normalized = normalizeProjectGraphViewResult(result, 'graph-correlation')

    expect(normalized.ok).toBe(true)
    expect(normalized.canvas?.nodes[0].badges).toEqual(['context_ready'])
    expect(normalized.canvas?.edges[0].validity).toBe('missing_endpoint')
    expect(normalized.canvas?.edges[0].error?.code).toBe('graph_view_missing_endpoint')
  })

  it('normalizes mock run DTOs with runtime events and placeholder output', () => {
    const normalized = normalizeMockRunResult({
      ok: true,
      run: {
        runId: 'run_mock_shot_002_001',
        projectId: 'proj_alpha_fixture',
        shotId: 'shot_002',
        selectionIds: ['shot_002'],
        taskMode: 'mock_local',
        providerMode: 'mock_local',
        contextDigest: 'sha256:test',
        status: 'completed',
        attempt: 1,
        runPath: 'prompts/runs/run_mock_shot_002_001/run.json',
        eventsPath: 'prompts/runs/run_mock_shot_002_001/events.jsonl',
        output: {
          runId: 'run_mock_shot_002_001',
          relativePath: 'assets/outputs/mock-run/run_mock_shot_002_001/placeholder-output.txt',
          digest: 'sha256-output',
          mimeType: 'text/plain',
          sizeBytes: 42,
          summary: 'placeholder',
        },
        createdAt: '2026-05-20T15:00:00Z',
        updatedAt: '2026-05-20T15:00:00Z',
      },
      events: [
        {
          eventId: 'evt-run-complete',
          runId: 'run_mock_shot_002_001',
          eventType: 'run.complete',
          state: 'completed',
          progress: 100,
          summary: 'Done',
          nextActions: [],
          createdAt: '2026-05-20T15:00:00Z',
        },
      ],
    }, 'mock-correlation')

    expect(normalized.ok).toBe(true)
    expect(normalized.run?.providerMode).toBe('mock_local')
    expect(normalized.run?.output?.relativePath).toContain('assets/outputs/mock-run')
    expect(normalized.events[0].progress).toBe(100)
  })
})
