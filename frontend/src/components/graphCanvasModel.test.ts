import { describe, expect, it } from 'vitest'

import type { ProjectCanvasDTO } from '../api/dto'
import { DEFAULT_ALPHA_PROJECT_ROOT } from '../api/projectRoot'
import {
  buildG6GraphData,
  buildGraphLayoutSaveCommand,
  visibleCanvasIssues,
} from './graphCanvasModel'

describe('Graph Canvas model', () => {
  it('maps domain nodes, organizer nodes, and renderable edges into G6 data', () => {
    const canvas = sampleCanvas()
    const data = buildG6GraphData(canvas, 'dark')

    expect(data.nodes?.map((node) => node.id)).toContain('node_shot_001')
    expect(data.nodes?.map((node) => node.id)).toContain('frame:frame_001')
    expect(data.nodes?.map((node) => node.id)).toContain('group:group_001')
    expect(data.edges?.map((edge) => edge.id)).toEqual(['edge_context', 'edge_invalid_relation'])
  })

  it('builds save commands from live graph positions without adding organizer layout', () => {
    const canvas = sampleCanvas()
    const command = buildGraphLayoutSaveCommand(canvas, {
      positions: {
        node_script_source: { x: 111, y: 222 },
        node_shot_001: { x: 333, y: 444 },
      },
      viewport: { x: -25, y: 40, zoom: 1.2 },
    }, 'warm_light', { visible: false, size: 32, opacity: 0.2 })

    expect(command.expectedGraphVersion).toBe(3)
    expect(command.root).toBe(DEFAULT_ALPHA_PROJECT_ROOT)
    expect(command.viewport.zoom).toBe(1.2)
    expect(command.theme).toBe('warm_light')
    expect(command.grid.visible).toBe(false)
    expect(command.nodes).toHaveLength(2)
    expect(command.nodes[0]).toMatchObject({
      id: 'node_script_source',
      position: { x: 111, y: 222 },
    })
    expect(command.nodes.map((node) => node.id)).not.toContain('frame:frame_001')
  })

  it('maps fixture status badges into node-level G6 badge render data', () => {
    const data = buildG6GraphData(sampleCanvas(), 'warm_light')
    const shotNode = data.nodes?.find((node) => node.id === 'node_shot_001')

    expect(shotNode?.data?.badges).toEqual([{ kind: 'context_ready', text: 'CTX' }])
    expect(shotNode?.style?.badge).toBe(true)
    expect(shotNode?.style?.badges).toEqual([
      expect.objectContaining({
        text: 'CTX',
        placement: 'right-top',
        background: true,
      }),
    ])
  })

  it('keeps invalid edge records visible for Canvas overlays', () => {
    const issues = visibleCanvasIssues(sampleCanvas())

    expect(issues.map((edge) => edge.id)).toEqual(['edge_invalid_relation', 'edge_missing'])
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
        status: 'draft',
        badges: ['draft'],
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
        status: 'context_ready',
        badges: ['context_ready'],
        source: 'human',
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
        id: 'edge_invalid_relation',
        sourceNodeId: 'node_script_source',
        targetNodeId: 'node_shot_001',
        relation: 'result_of',
        label: 'result_of',
        createdAt: '2026-05-20T05:40:00Z',
        validity: 'invalid_relation',
      },
      {
        id: 'edge_missing',
        sourceNodeId: 'node_shot_001',
        targetNodeId: 'node_missing',
        relation: 'uses',
        label: 'uses',
        createdAt: '2026-05-20T05:40:00Z',
        validity: 'missing_endpoint',
      },
    ],
    frames: [
      {
        id: 'frame_001',
        title: 'Frame 001',
        referenceGroupIds: ['group_001'],
        outputNodeIds: ['node_shot_001'],
        taskIntent: 'shot package',
        requiredCapabilities: ['manual_handoff'],
        historySummary: { favoriteRunIds: [] },
        layout: { x: 80, y: 40, width: 600, height: 280, collapsed: false },
      },
    ],
    referenceGroups: [
      {
        id: 'group_001',
        title: 'Reference Group',
        role: 'continuity',
        inputNodeIds: ['node_script_source'],
        priority: 1,
        layout: { x: 100, y: 60, width: 260, height: 180, collapsed: false },
      },
    ],
    selectedNodeIds: [],
    updatedAt: '2026-05-20T05:40:00Z',
  }
}
