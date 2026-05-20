import { describe, expect, it } from 'vitest'

import type { ProjectCanvasDTO } from '../api/dto'
import {
  continuityRisks,
  firstErrorSummary,
  healthTone,
  highestHealthStatus,
  latestEventSummary,
  nodeDataRows,
  nodeStatusBadges,
  relationSummaries,
  selectionLabel,
  summarizeSaveState,
  statusBadgeView,
} from './workbenchFeedback'

describe('Workbench feedback model', () => {
  it('maps node status and badges into stable badge semantics', () => {
    const [status, missing] = nodeStatusBadges({
      ...sampleCanvas().nodes[0],
      status: 'context_ready',
      badges: ['context_ready', 'missing_asset'],
    })

    expect(status).toMatchObject({ label: 'Context ready', tone: 'success' })
    expect(missing).toMatchObject({ label: 'Missing asset', tone: 'error' })
  })

  it.each([
    ['missing_ref', 'Missing ref', 'error'],
    ['missing_asset', 'Missing asset', 'error'],
    ['missing_context', 'Missing context', 'error'],
    ['blocked', 'Blocked', 'error'],
    ['context_dirty', 'Context dirty', 'warning'],
    ['context_ready', 'Context ready', 'success'],
    ['package_ready', 'Package ready', 'success'],
    ['submitted', 'Submitted', 'processing'],
    ['pending_review', 'Pending review', 'processing'],
    ['approved', 'Approved', 'success'],
    ['needs_revision', 'Needs revision', 'warning'],
    ['invalid', 'Invalid', 'error'],
  ] as const)('maps %s into the expected badge view', (value, label, tone) => {
    expect(statusBadgeView(value)).toMatchObject({ label, tone })
  })

  it('summarizes selected relations with endpoint titles and validity tones', () => {
    const rows = relationSummaries(sampleCanvas(), ['node_shot_001'])

    expect(rows.map((row) => row.label)).toEqual([
      'Script Source -> Shot 001',
      'Shot 001 -> Missing Endpoint',
    ])
    expect(rows[0]).toMatchObject({ direction: 'incoming', tone: 'success' })
    expect(rows[1]).toMatchObject({ direction: 'outgoing', tone: 'warning' })
  })

  it('promotes dirty badges and invalid relations into continuity risks', () => {
    const canvas = sampleCanvas()
    const rows = relationSummaries(canvas, ['node_shot_001'])
    const risks = continuityRisks(canvas.nodes[1], rows)

    expect(risks.map((risk) => risk.label)).toContain('Context dirty')
    expect(risks.map((risk) => risk.label)).toContain('Missing Endpoint')
  })

  it('sorts node data into readable key/value rows', () => {
    const rows = nodeDataRows({
      ...sampleCanvas().nodes[0],
      data: { summary: 'Opening shot', path: 'shots/shot-001.json' },
    })

    expect(rows.map((row) => row.label)).toEqual(['Path', 'Summary'])
  })

  it('summarizes save state, errors, events, and selection', () => {
    expect(summarizeSaveState('failed')).toMatchObject({ label: 'Save failed', tone: 'error' })
    expect(firstErrorSummary([
      undefined,
      {
        code: 'save_failed',
        severity: 'error',
        retryable: true,
        userMessage: 'Save failed',
        recoveryActions: ['Retry'],
        correlationId: 'corr',
      },
    ])).toMatchObject({ code: 'save_failed', retryable: true })
    expect(latestEventSummary([
      {
        eventId: 'evt-1',
        eventType: 'graph.loaded',
        state: 'completed',
        summary: 'Graph loaded',
        nextActions: [],
        createdAt: '2026-05-20T05:40:00Z',
      },
    ], [], [])).toBe('graph.loaded · completed')
    expect(selectionLabel(sampleCanvas().nodes, undefined)).toBe('Selection: 3 nodes')
  })

  it('aggregates health from graph health, project health, and visible errors', () => {
    expect(highestHealthStatus([
      { status: 'clean', checkedAt: '2026-05-20T05:40:00Z', items: [] },
      { status: 'blocking', checkedAt: '2026-05-20T05:41:00Z', items: [] },
    ], [])).toBe('blocking')
    expect(highestHealthStatus([
      { status: 'clean', checkedAt: '2026-05-20T05:40:00Z', items: [] },
    ], [
      {
        code: 'save_failed',
        severity: 'error',
        retryable: true,
        userMessage: 'Save failed',
        recoveryActions: ['Retry'],
        correlationId: 'corr',
      },
    ])).toBe('blocking')
    expect(healthTone('warning')).toBe('warning')
  })
})

function sampleCanvas(): ProjectCanvasDTO {
  return {
    id: 'proj_alpha_fixture_canvas',
    projectId: 'proj_alpha_fixture',
    schemaVersion: '1.0.0',
    version: 3,
    viewport: { x: 0, y: 0, zoom: 1 },
    theme: 'dark',
    grid: { visible: true, size: 24, opacity: 0.24 },
    nodes: [
      {
        id: 'node_script_source',
        kind: 'script',
        category: 'source',
        title: 'Script Source',
        refId: 'assets/inputs/script-source.md',
        position: { x: 120, y: 120 },
        size: { width: 220, height: 96 },
        collapsed: false,
        status: 'context_ready',
        badges: ['context_ready'],
        source: 'imported',
      },
      {
        id: 'node_shot_001',
        kind: 'shot',
        category: 'production',
        title: 'Shot 001',
        refId: 'shots/shot-001.json',
        position: { x: 460, y: 120 },
        size: { width: 240, height: 112 },
        collapsed: false,
        status: 'draft',
        badges: ['context_dirty'],
        source: 'human',
      },
      {
        id: 'node_missing',
        kind: 'asset',
        category: 'source',
        title: 'Missing Endpoint',
        position: { x: 760, y: 120 },
        size: { width: 220, height: 96 },
        collapsed: false,
        badges: ['missing_ref'],
        source: 'system',
      },
    ],
    edges: [
      {
        id: 'edge_context',
        sourceNodeId: 'node_script_source',
        targetNodeId: 'node_shot_001',
        relation: 'uses',
        label: 'uses context',
        createdAt: '2026-05-20T05:40:00Z',
        validity: 'valid',
      },
      {
        id: 'edge_missing',
        sourceNodeId: 'node_shot_001',
        targetNodeId: 'node_missing',
        relation: 'uses',
        label: 'uses',
        createdAt: '2026-05-20T05:41:00Z',
        validity: 'missing_endpoint',
      },
    ],
    frames: [],
    referenceGroups: [],
    selectedNodeIds: [],
    updatedAt: '2026-05-20T05:40:00Z',
  }
}
