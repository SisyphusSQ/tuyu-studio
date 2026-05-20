import type { EdgeData, GraphData, NodeData } from '@antv/g6'

import {
  createCorrelationId,
  type CanvasGridDTO,
  type CanvasPositionDTO,
  type CanvasTheme,
  type CanvasViewportDTO,
  type GraphEdgeDTO,
  type GraphNodeDTO,
  type ProductionFrameDTO,
  type ProjectCanvasDTO,
  type ProjectGraphLayoutSaveCommandDTO,
  type ReferenceGroupDTO,
} from '../api/dto'

export type CanvasElementRole = 'domain_node' | 'reference_group' | 'production_frame'

export interface CanvasElementMeta {
  role: CanvasElementRole
  id: string
  title: string
  kind: string
  category: string
  persistLayout: boolean
}

export interface GraphLayoutSnapshot {
  positions: Record<string, CanvasPositionDTO>
  viewport: CanvasViewportDTO
}

export function buildG6GraphData(canvas: ProjectCanvasDTO, theme: CanvasTheme = canvas.theme): GraphData {
  const domainNodeIDs = new Set(canvas.nodes.map((node) => node.id))
  const organizerNodes = [
    ...canvas.frames.map((frame) => productionFrameNode(frame, theme)),
    ...canvas.referenceGroups.map((group) => referenceGroupNode(group, theme)),
  ]
  const nodes = [
    ...organizerNodes,
    ...canvas.nodes.map((node) => graphNode(node, theme)),
  ]
  const edges = canvas.edges
    .filter((edge) => domainNodeIDs.has(edge.sourceNodeId) && domainNodeIDs.has(edge.targetNodeId))
    .map((edge) => graphEdge(edge, theme))

  return { nodes, edges }
}

export function buildGraphLayoutSaveCommand(
  canvas: ProjectCanvasDTO,
  snapshot: GraphLayoutSnapshot,
  theme: CanvasTheme,
  grid: CanvasGridDTO,
): ProjectGraphLayoutSaveCommandDTO {
  return {
    root: '',
    expectedGraphVersion: canvas.version,
    viewport: snapshot.viewport,
    theme,
    grid,
    nodes: canvas.nodes.map((node) => ({
      id: node.id,
      position: snapshot.positions[node.id] || node.position,
      size: node.size,
      collapsed: node.collapsed,
    })),
    correlationId: createCorrelationId('project-graph-layout-save'),
  }
}

export function visibleCanvasIssues(canvas: ProjectCanvasDTO): GraphEdgeDTO[] {
  return canvas.edges.filter((edge) => edge.validity !== 'valid')
}

function graphNode(node: GraphNodeDTO, theme: CanvasTheme): NodeData {
  const colors = nodeColors(node.category, theme)
  const meta: CanvasElementMeta = {
    role: 'domain_node',
    id: node.id,
    title: node.title,
    kind: node.kind,
    category: node.category,
    persistLayout: true,
  }

  return {
    id: node.id,
    type: 'rect',
    data: { meta, dto: node },
    style: {
      x: node.position.x,
      y: node.position.y,
      size: [node.size.width, node.size.height],
      radius: 8,
      fill: colors.fill,
      stroke: colors.stroke,
      lineWidth: 1.4,
      shadowColor: theme === 'dark' ? 'rgba(0,0,0,0.28)' : 'rgba(69,71,60,0.12)',
      shadowBlur: 12,
      labelText: compactLabel(node.title),
      labelFill: colors.text,
      labelFontSize: 13,
      labelFontWeight: 700,
      labelWordWrap: true,
      labelMaxWidth: Math.max(120, node.size.width - 28),
      zIndex: 4,
    },
  }
}

function referenceGroupNode(group: ReferenceGroupDTO, theme: CanvasTheme): NodeData {
  const meta: CanvasElementMeta = {
    role: 'reference_group',
    id: group.id,
    title: group.title,
    kind: 'reference_group',
    category: 'organizer',
    persistLayout: false,
  }

  return {
    id: `group:${group.id}`,
    type: 'rect',
    data: { meta, dto: group },
    style: {
      x: group.layout.x + group.layout.width / 2,
      y: group.layout.y + group.layout.height / 2,
      size: [group.layout.width, group.layout.height],
      radius: 10,
      fill: theme === 'dark' ? 'rgba(65, 111, 96, 0.16)' : 'rgba(82, 137, 106, 0.13)',
      stroke: theme === 'dark' ? '#4c9b80' : '#5f9875',
      lineDash: [8, 8],
      lineWidth: 1,
      labelText: compactLabel(group.title),
      labelFill: theme === 'dark' ? '#9ed9c2' : '#37614d',
      labelFontSize: 12,
      labelFontWeight: 650,
      zIndex: 1,
    },
  }
}

function productionFrameNode(frame: ProductionFrameDTO, theme: CanvasTheme): NodeData {
  const meta: CanvasElementMeta = {
    role: 'production_frame',
    id: frame.id,
    title: frame.title,
    kind: 'production_frame',
    category: 'organizer',
    persistLayout: false,
  }

  return {
    id: `frame:${frame.id}`,
    type: 'rect',
    data: { meta, dto: frame },
    style: {
      x: frame.layout.x + frame.layout.width / 2,
      y: frame.layout.y + frame.layout.height / 2,
      size: [frame.layout.width, frame.layout.height],
      radius: 12,
      fill: theme === 'dark' ? 'rgba(246, 241, 229, 0.06)' : 'rgba(255, 255, 255, 0.58)',
      stroke: theme === 'dark' ? 'rgba(225, 229, 214, 0.28)' : 'rgba(85, 91, 77, 0.24)',
      lineWidth: 1.2,
      labelText: compactLabel(frame.title),
      labelFill: theme === 'dark' ? '#dbe7d8' : '#52604f',
      labelFontSize: 12,
      labelFontWeight: 700,
      zIndex: 0,
    },
  }
}

function graphEdge(edge: GraphEdgeDTO, theme: CanvasTheme): EdgeData {
  const invalid = edge.validity !== 'valid'
  return {
    id: edge.id,
    source: edge.sourceNodeId,
    target: edge.targetNodeId,
    type: 'polyline',
    data: { dto: edge },
    style: {
      stroke: invalid ? '#c45f5f' : theme === 'dark' ? '#8fcfb8' : '#477966',
      lineWidth: invalid ? 1.8 : 1.2,
      lineDash: invalid ? [6, 6] : undefined,
      endArrow: true,
      labelText: edge.label,
      labelFill: theme === 'dark' ? '#b9c6b7' : '#596253',
      labelFontSize: 11,
      labelBackground: true,
      labelBackgroundFill: theme === 'dark' ? '#141a16' : '#fffaf0',
      zIndex: 3,
    },
  }
}

function nodeColors(category: string, theme: CanvasTheme) {
  if (theme === 'warm_light') {
    switch (category) {
      case 'production':
        return { fill: '#fff1d1', stroke: '#c78931', text: '#4b3920' }
      case 'continuity':
        return { fill: '#e9f3e5', stroke: '#6fa26f', text: '#29472d' }
      case 'output':
      case 'handoff':
        return { fill: '#e8f0f6', stroke: '#668da3', text: '#2b4554' }
      case 'review':
        return { fill: '#f3e7ed', stroke: '#a8718a', text: '#513143' }
      default:
        return { fill: '#fffdf7', stroke: '#8fa28f', text: '#283027' }
    }
  }

  switch (category) {
    case 'production':
      return { fill: '#243828', stroke: '#d1a34d', text: '#f5ecd5' }
    case 'continuity':
      return { fill: '#1d3329', stroke: '#70b28f', text: '#e5f3e9' }
    case 'output':
    case 'handoff':
      return { fill: '#20323b', stroke: '#77a8c0', text: '#e7f2f6' }
    case 'review':
      return { fill: '#352632', stroke: '#b47b9a', text: '#f2e4ec' }
    default:
      return { fill: '#202820', stroke: '#8fb59d', text: '#eef2ec' }
  }
}

function compactLabel(label: string): string {
  return label.length > 48 ? `${label.slice(0, 45)}...` : label
}
