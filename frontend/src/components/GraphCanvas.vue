<script setup lang="ts">
import {
  CanvasEvent,
  Graph,
  GraphEvent,
  NodeEvent,
  type IEvent,
  type Point,
} from '@antv/g6'
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'

import type {
  CanvasGridDTO,
  CanvasPositionDTO,
  CanvasTheme,
  CanvasViewportDTO,
  GraphNodeDTO,
  ProjectCanvasDTO,
  ProjectGraphLayoutSaveCommandDTO,
} from '../api/dto'
import {
  buildG6GraphData,
  buildGraphLayoutSaveCommand,
  visibleCanvasIssues,
} from './graphCanvasModel'

const props = defineProps<{
  canvas?: ProjectCanvasDTO
  theme: CanvasTheme
  grid: CanvasGridDTO
  loading?: boolean
  saving?: boolean
}>()

const emit = defineEmits<{
  save: [command: ProjectGraphLayoutSaveCommandDTO]
  select: [node: GraphNodeDTO | undefined]
  selection: [nodes: GraphNodeDTO[]]
  viewport: [viewport: CanvasViewportDTO]
}>()

const surface = ref<HTMLDivElement>()
const graph = ref<Graph>()
const selectedNodeID = ref<string>()
const selectedNodeIDs = ref<string[]>([])
const renderError = ref<string>()
const currentViewport = ref<CanvasViewportDTO>({ x: 0, y: 0, zoom: 1 })

const issueEdges = computed(() => props.canvas ? visibleCanvasIssues(props.canvas) : [])
const selectedNode = computed(() => {
  const id = selectedNodeID.value || selectedNodeIDs.value.at(-1)
  if (!props.canvas || !id) {
    return undefined
  }
  return props.canvas.nodes.find((node) => node.id === id)
})
const selectedNodes = computed(() => {
  if (!props.canvas) {
    return []
  }
  const selected = new Set(selectedNodeIDs.value)
  return props.canvas.nodes.filter((node) => selected.has(node.id))
})
const canvasStats = computed(() => ({
  nodes: props.canvas?.nodes.length || 0,
  edges: props.canvas?.edges.length || 0,
  frames: props.canvas?.frames.length || 0,
  references: props.canvas?.referenceGroups.length || 0,
}))
const stageClass = computed(() => ({
  'graph-canvas--warm': props.theme === 'warm_light',
  'graph-canvas--loading': Boolean(props.loading || props.saving),
}))

watch(
  () => [props.canvas, props.theme, props.grid.visible, props.grid.size, props.grid.opacity] as const,
  (next, previous) => {
    const preserveCurrentLayout = Boolean(previous && next[0] && next[0] === previous[0])
    void renderCanvas({ preserveCurrentLayout })
  },
)

watch(selectedNode, (node) => {
  emit('select', node)
})

watch(selectedNodes, (nodes) => {
  emit('selection', nodes)
})

watch(selectedNodeIDs, () => {
  void syncSelectedElementStates()
})

onMounted(() => {
  void renderCanvas()
})

onBeforeUnmount(() => {
  graph.value?.destroy()
  graph.value = undefined
})

interface RenderOptions {
  preserveCurrentLayout?: boolean
}

async function renderCanvas(options: RenderOptions = {}) {
  if (!surface.value) {
    return
  }
  if (!props.canvas) {
    clearGraph()
    return
  }

  try {
    renderError.value = undefined
    const renderSource = options.preserveCurrentLayout
      ? canvasWithCurrentLayout(props.canvas)
      : props.canvas
    const data = buildG6GraphData(renderSource, props.theme)
    const plugins = props.grid.visible
      ? [{
        type: 'grid-line',
        key: 'grid-line',
        size: props.grid.size,
        stroke: gridStroke(props.theme, props.grid.opacity),
        lineWidth: 1,
      }]
      : []

    if (!graph.value) {
      graph.value = new Graph({
        container: surface.value,
        autoResize: true,
        animation: { duration: 120 },
        data,
        node: {
          type: 'rect',
          style: (datum) => datum.style || {},
          state: {
            selected: {
              stroke: '#f0c35b',
              lineWidth: 3,
            },
          },
        },
        edge: {
          type: 'polyline',
          style: (datum) => datum.style || {},
        },
        behaviors: [
          'drag-canvas',
          'zoom-canvas',
          'drag-element',
        ],
        plugins,
      })
      bindGraphEvents(graph.value)
    } else {
      graph.value.setPlugins(plugins)
      graph.value.setData(data)
    }

    await graph.value.render()
    await applyCanvasViewport(renderSource.viewport)
    await syncSelectedElementStates()
  } catch (error) {
    renderError.value = error instanceof Error ? error.message : String(error)
  }
}

function clearGraph() {
  graph.value?.destroy()
  graph.value = undefined
  selectedNodeID.value = undefined
  selectedNodeIDs.value = []
  renderError.value = undefined
}

function bindGraphEvents(instance: Graph) {
  instance.on(NodeEvent.CLICK, (event: IEvent) => {
    const id = eventTargetID(event)
    updateSelection(id, event)
  })
  instance.on(NodeEvent.DRAG_END, () => {
    currentViewport.value = readViewport()
    emit('viewport', currentViewport.value)
  })
  instance.on(CanvasEvent.CLICK, () => {
    clearSelection()
  })
  instance.on(GraphEvent.AFTER_TRANSFORM, () => {
    currentViewport.value = readViewport()
    emit('viewport', currentViewport.value)
  })
}

async function applyCanvasViewport(viewport: CanvasViewportDTO) {
  if (!graph.value) {
    return
  }
  await graph.value.zoomTo(Math.max(0.2, viewport.zoom), { duration: 0 })
  await graph.value.translateTo([viewport.x, viewport.y], { duration: 0 })
  currentViewport.value = readViewport()
  emit('viewport', currentViewport.value)
}

function fitView() {
  void graph.value?.fitView({ when: 'always', direction: 'both' }, { duration: 160 })
}

function focusSelected() {
  if (!graph.value || selectedNodeIDs.value.length === 0) {
    return
  }
  const ids = selectedNodeIDs.value.filter((id) => graph.value?.hasNode(id))
  if (ids.length > 0) {
    void graph.value.focusElement(ids, { duration: 160 })
  }
}

async function nudgeSelected() {
  const id = selectedNodeID.value || props.canvas?.nodes[0]?.id
  if (!graph.value || !id || !graph.value.hasNode(id)) {
    return
  }
  replaceSelection(id)
  await graph.value.translateElementBy(id, [32, 0], false)
}

function saveLayout() {
  if (!props.canvas) {
    return
  }
  emit('save', buildGraphLayoutSaveCommand(props.canvas, {
    positions: readNodePositions(),
    viewport: readViewport(),
  }, props.theme, props.grid))
}

function updateSelection(id: string | undefined, event: IEvent) {
  if (!id || !props.canvas?.nodes.some((node) => node.id === id)) {
    clearSelection()
    return
  }
  if (selectionModifierPressed(event)) {
    toggleSelectedNode(id)
    return
  }
  replaceSelection(id)
}

function replaceSelection(id: string) {
  selectedNodeID.value = id
  selectedNodeIDs.value = [id]
}

function toggleSelectedNode(id: string) {
  if (selectedNodeIDs.value.includes(id)) {
    selectedNodeIDs.value = selectedNodeIDs.value.filter((value) => value !== id)
    selectedNodeID.value = selectedNodeIDs.value.at(-1)
    return
  }
  selectedNodeIDs.value = [...selectedNodeIDs.value, id]
  selectedNodeID.value = id
}

function clearSelection() {
  selectedNodeID.value = undefined
  selectedNodeIDs.value = []
}

function selectionModifierPressed(event: IEvent): boolean {
  return Boolean('shiftKey' in event && event.shiftKey)
}

async function syncSelectedElementStates() {
  if (!graph.value || !props.canvas) {
    return
  }
  const selected = new Set(selectedNodeIDs.value)
  const states: Record<string, string[]> = {}
  for (const node of props.canvas.nodes) {
    if (graph.value.hasNode(node.id)) {
      states[node.id] = selected.has(node.id) ? ['selected'] : []
    }
  }
  await graph.value.setElementState(states, false)
}

function readNodePositions(): Record<string, CanvasPositionDTO> {
  const positions: Record<string, CanvasPositionDTO> = {}
  if (!graph.value || !props.canvas) {
    return positions
  }

  for (const node of props.canvas.nodes) {
    positions[node.id] = graph.value.hasNode(node.id)
      ? pointToCanvasPosition(graph.value.getElementPosition(node.id), node.position)
      : node.position
  }
  return positions
}

function readViewport(): CanvasViewportDTO {
  if (!graph.value) {
    return props.canvas?.viewport || { x: 0, y: 0, zoom: 1 }
  }
  try {
    const point = graph.value.getPosition()
    return {
      x: Math.round(point[0] || 0),
      y: Math.round(point[1] || 0),
      zoom: Number(graph.value.getZoom().toFixed(3)),
    }
  } catch {
    return props.canvas?.viewport || { x: 0, y: 0, zoom: 1 }
  }
}

function canvasWithCurrentLayout(canvas: ProjectCanvasDTO): ProjectCanvasDTO {
  const positions = readNodePositions()
  const viewport = readViewport()
  return {
    ...canvas,
    viewport,
    nodes: canvas.nodes.map((node) => ({
      ...node,
      position: positions[node.id] || node.position,
    })),
  }
}

function gridStroke(theme: CanvasTheme, opacity: number): string {
  const alpha = Math.min(1, Math.max(0, opacity))
  return theme === 'warm_light'
    ? `rgba(80, 103, 84, ${alpha})`
    : `rgba(219, 225, 210, ${alpha})`
}

function pointToCanvasPosition(point: Point, fallback: CanvasPositionDTO): CanvasPositionDTO {
  return {
    x: Math.round(point[0] ?? fallback.x),
    y: Math.round(point[1] ?? fallback.y),
  }
}

function eventTargetID(event: IEvent): string | undefined {
  const target = 'target' in event ? event.target : undefined
  const id = target && 'id' in target ? target.id : undefined
  return typeof id === 'string' ? id : undefined
}

defineExpose({
  fitView,
  focusSelected,
  nudgeSelected,
  saveLayout,
})
</script>

<template>
  <section class="graph-canvas" :class="stageClass" data-testid="project-graph-canvas">
    <div ref="surface" class="graph-canvas-surface" />

    <div v-if="!canvas && !loading" class="graph-canvas-empty">
      <strong>Graph unavailable</strong>
      <span>Project Canvas DTO is not loaded.</span>
    </div>

    <div v-if="loading || saving" class="graph-canvas-busy">
      <span>{{ saving ? 'Saving layout' : 'Loading graph' }}</span>
    </div>

    <div v-if="canvas" class="graph-canvas-overlay">
      <div class="graph-canvas-stats">
        <span>{{ canvasStats.nodes }} nodes</span>
        <span>{{ canvasStats.edges }} edges</span>
        <span>{{ canvasStats.frames }} frames</span>
        <span>{{ canvasStats.references }} refs</span>
      </div>

      <div v-if="selectedNode" class="graph-canvas-selection">
        <strong :title="selectedNode.title">{{ selectedNode.title }}</strong>
        <span>{{ selectedNode.kind }} · {{ selectedNode.status || 'ready' }}</span>
      </div>

      <div v-if="issueEdges.length" class="graph-canvas-issues">
        <strong>{{ issueEdges.length }} relation issues</strong>
        <span v-for="edge in issueEdges.slice(0, 2)" :key="edge.id">
          {{ edge.id }} · {{ edge.validity }}
        </span>
      </div>
    </div>

    <div v-if="renderError" class="graph-canvas-error">
      <strong>Canvas render failed</strong>
      <span>{{ renderError }}</span>
    </div>
  </section>
</template>
