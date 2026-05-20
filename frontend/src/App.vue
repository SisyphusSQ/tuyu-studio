<script setup lang="ts">
import {
  Alert,
  Badge,
  Button,
  Layout,
  List,
  Menu,
  Progress,
  Segmented,
  Tabs,
  Tag,
  Tooltip,
} from 'ant-design-vue'
import {
  AppstoreOutlined,
  BgColorsOutlined,
  BranchesOutlined,
  CloudUploadOutlined,
  CopyOutlined,
  DatabaseOutlined,
  FileTextOutlined,
  FolderOpenOutlined,
  HistoryOutlined,
  ReloadOutlined,
  SaveOutlined,
  SearchOutlined,
  SettingOutlined,
  ShareAltOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons-vue'
import { computed, h, onMounted, ref } from 'vue'

import {
  type AppErrorDTO,
  type CanvasGridDTO,
  type CanvasTheme,
  type GraphNodeDTO,
  type ProjectCanvasDTO,
  type ProjectGraphLayoutSaveCommandDTO,
  type ProjectGraphViewResultDTO,
  type ProjectOperationName,
  type ProjectOperationResultDTO,
  type RuntimeEventDTO,
  type ScriptSceneCandidateResultDTO,
  type ScriptDocumentDTO,
  type ScriptDocumentResultDTO,
  type ShotCandidateDTO,
  type WorkbenchProbeMode,
  type WorkbenchStatusDTO,
} from './api/dto'
import {
  DEFAULT_SCRIPT_DOCUMENT_ID,
  DEFAULT_SCRIPT_SOURCE_ASSET_ID,
  loadScriptDocument,
  saveScriptDocument,
} from './api/scriptDocument'
import {
  confirmScriptScene,
  confirmShotCandidate,
  listShotCandidates,
  nextShotCandidatesAfterResult,
  rejectShotCandidate,
  saveShotCandidate,
} from './api/scriptSceneCandidates'
import { runProjectGraphView, runProjectOperation, runWorkbenchProbe, saveProjectGraphLayout } from './api/workbench'
import GraphCanvas from './components/GraphCanvas.vue'
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
  type SaveFeedbackState,
} from './components/workbenchFeedback'

const LayoutHeader = Layout.Header
const LayoutContent = Layout.Content
const LayoutSider = Layout.Sider
const ListItem = List.Item
const ListItemMeta = List.Item.Meta
const TabPane = Tabs.TabPane

const themeMode = ref<'dark' | 'warm'>('dark')
const activeRailTab = ref('assets')
const activeInspectorTab = ref('properties')
const probeLoading = ref(false)
const probeSnapshot = ref<WorkbenchStatusDTO>()
const probeError = ref<AppErrorDTO>()
const runtimeEvents = ref<RuntimeEventDTO[]>([])
const projectLoading = ref<ProjectOperationName>()
const projectResult = ref<ProjectOperationResultDTO>()
const graphCanvasRef = ref<InstanceType<typeof GraphCanvas>>()
const graphLoading = ref(false)
const graphSaving = ref(false)
const graphResult = ref<ProjectGraphViewResultDTO>()
const activeGraphCanvas = ref<ProjectCanvasDTO>()
const selectedGraphNode = ref<GraphNodeDTO>()
const selectedGraphNodes = ref<GraphNodeDTO[]>([])
const graphViewportLabel = ref('100%')
const canvasGrid = ref<CanvasGridDTO>({ visible: true, size: 24, opacity: 0.24 })
const saveState = ref<SaveFeedbackState>('idle')
const lastSavedAt = ref<string>()
const copiedInspectorText = ref('')
const scriptDocument = ref<ScriptDocumentDTO>()
const scriptResult = ref<ScriptDocumentResultDTO>()
const scriptTitle = ref('Alpha Script')
const scriptSourceAssetId = ref(DEFAULT_SCRIPT_SOURCE_ASSET_ID)
const scriptRawText = ref('')
const scriptLogline = ref('')
const scriptSynopsis = ref('')
const scriptLoading = ref(false)
const scriptSaving = ref(false)
const scriptDirty = ref(false)
const sceneResult = ref<ScriptSceneCandidateResultDTO>()
const shotCandidates = ref<ShotCandidateDTO[]>([])
const sceneId = ref('scene_001')
const sceneTitle = ref('Opening Scene')
const sceneLocation = ref('Laneway market')
const sceneTimeOfDay = ref('evening')
const sceneCharacters = ref('Mina')
const sceneProps = ref('Lantern')
const sceneAction = ref('')
const sceneEmotionalBeat = ref('')
const sceneStartLine = ref(1)
const sceneEndLine = ref(1)
const sceneAllowOverlap = ref(false)
const sceneSaving = ref(false)
const candidateId = ref('candidate_scene_001_001')
const candidateIndex = ref(1)
const candidateDuration = ref(5)
const candidateDescription = ref('')
const candidateCharacters = ref('Mina')
const candidateStartLine = ref(1)
const candidateEndLine = ref(1)
const candidateSaving = ref(false)
const candidateAction = ref('')

const railItems = [
  { key: 'assets', icon: () => h(DatabaseOutlined), label: 'Assets' },
  { key: 'script', icon: () => h(FileTextOutlined), label: 'Script' },
  { key: 'blueprint', icon: () => h(BranchesOutlined), label: 'Blueprint' },
  { key: 'search', icon: () => h(SearchOutlined), label: 'Search' },
  { key: 'history', icon: () => h(HistoryOutlined), label: 'History' },
]

const assets = [
  { name: 'hero-reference.png', type: 'image', status: 'ready' },
  { name: 'night-market-scene.mov', type: 'video', status: 'missing link' },
  { name: 'dialogue-source.txt', type: 'script', status: 'draft' },
]

const queueItems = [
  { label: 'Package export', value: 'idle' },
  { label: 'Mock run', value: 'not configured' },
  { label: 'Review import', value: 'waiting for TOO-180' },
]

const graphCanvas = computed(() => activeGraphCanvas.value)
const graphErrors = computed(() => graphResult.value?.errors || [])
const graphTheme = computed<CanvasTheme>(() => themeMode.value === 'warm' ? 'warm_light' : 'dark')
const selectedGraphNodeIDs = computed(() => selectedGraphNodes.value.length > 0
  ? selectedGraphNodes.value.map((node) => node.id)
  : selectedGraphNode.value
    ? [selectedGraphNode.value.id]
    : [])
const selectedNodeBadges = computed(() => nodeStatusBadges(selectedGraphNode.value))
const selectedNodeDataRows = computed(() => nodeDataRows(selectedGraphNode.value))
const selectedRelations = computed(() => relationSummaries(graphCanvas.value, selectedGraphNodeIDs.value))
const continuityRiskRows = computed(() => continuityRisks(selectedGraphNode.value, selectedRelations.value))
const saveBadge = computed(() => summarizeSaveState(saveState.value, lastSavedAt.value))
const visibleErrors = computed(() => [
  graphResult.value?.error,
  projectResult.value?.error,
  scriptResult.value?.error,
  sceneResult.value?.error,
  probeError.value,
  ...graphErrors.value,
].filter((error): error is AppErrorDTO => Boolean(error)))
const healthStatus = computed(() => highestHealthStatus([
  graphResult.value?.health,
  projectResult.value?.health,
], visibleErrors.value))
const healthStatusTone = computed(() => healthTone(healthStatus.value))
const errorSummary = computed(() => firstErrorSummary([
  ...visibleErrors.value,
]))
const latestEventLabel = computed(() => latestEventSummary(
  graphResult.value?.events,
  [
    ...(projectResult.value?.events || []),
    ...(scriptResult.value?.events || []),
    ...(sceneResult.value?.events || []),
  ],
  runtimeEvents.value,
))
const selectedSummary = computed(() => selectionLabel(selectedGraphNodes.value, selectedGraphNode.value))
const inspectorSummary = computed(() => {
  if (!selectedGraphNode.value) {
    return `${graphCanvas.value?.projectId || 'project'} · ${graphCanvas.value?.nodes.length || 0} nodes`
  }
  const node = selectedGraphNode.value
  return [
    node.title,
    `kind=${node.kind}`,
    `status=${node.status || 'none'}`,
    `ref=${node.refId || 'none'}`,
  ].join(' · ')
})
const canvasHealthType = computed(() => {
  if (graphResult.value?.health?.status === 'blocking') {
    return 'error'
  }
  if (graphResult.value?.health?.status === 'warning' || graphErrors.value.length > 0) {
    return 'warning'
  }
  return 'success'
})
const scriptByteCount = computed(() => new TextEncoder().encode(scriptRawText.value).length)
const scriptStatusLabel = computed(() => {
  if (scriptResult.value?.ok && scriptDocument.value) {
    return `${scriptDocument.value.title} · ${scriptDocument.value.updatedAt || 'saved'}`
  }
  if (scriptResult.value?.error) {
    return `${scriptResult.value.error.code} · ${scriptResult.value.error.correlationId}`
  }
  return scriptDirty.value ? 'Unsaved' : 'Ready'
})
const scriptStatusType = computed(() => {
  if (scriptResult.value?.ok) {
    return 'success'
  }
  if (scriptResult.value?.error?.severity === 'blocking' || scriptResult.value?.error?.severity === 'error') {
    return 'error'
  }
  return scriptDirty.value ? 'warning' : 'info'
})
const scriptRecoveryActions = computed(() => scriptResult.value?.error?.recoveryActions || [])
const sceneRecoveryActions = computed(() => sceneResult.value?.error?.recoveryActions || [])
const scriptScenes = computed(() => scriptDocument.value?.scenes || [])
const activeSceneId = computed(() => sceneId.value.trim() || scriptScenes.value[0]?.id || '')
const candidateSummary = computed(() => {
  const accepted = shotCandidates.value.filter((candidate) => candidate.status === 'accepted').length
  const rejected = shotCandidates.value.filter((candidate) => candidate.status === 'rejected').length
  return `${shotCandidates.value.length} rows · ${accepted} accepted · ${rejected} rejected`
})

onMounted(() => {
  void loadProjectGraph()
})

async function runProbe(mode: WorkbenchProbeMode) {
  probeLoading.value = true
  const result = await runWorkbenchProbe(mode)

  probeSnapshot.value = result.ok ? result.snapshot : undefined
  probeError.value = result.error
  runtimeEvents.value = result.events
  activeInspectorTab.value = 'tasks'
  probeLoading.value = false
}

async function runProjectAction(action: ProjectOperationName) {
  projectLoading.value = action
  projectResult.value = await runProjectOperation(action)
  activeInspectorTab.value = 'tasks'
  projectLoading.value = undefined
}

async function loadProjectGraph(expectedGraphVersion?: number) {
  graphLoading.value = true
  graphResult.value = await runProjectGraphView(expectedGraphVersion)
  const canvas = graphResult.value.canvas
  if (canvas) {
    activeGraphCanvas.value = canvas
    themeMode.value = canvas.theme === 'warm_light' ? 'warm' : 'dark'
    canvasGrid.value = canvas.grid
    graphViewportLabel.value = `${Math.round(canvas.viewport.zoom * 100)}%`
  } else {
    activeGraphCanvas.value = undefined
    selectedGraphNode.value = undefined
    selectedGraphNodes.value = []
  }
  activeInspectorTab.value = graphResult.value.ok ? activeInspectorTab.value : 'tasks'
  graphLoading.value = false
}

async function saveCanvasLayout(command: ProjectGraphLayoutSaveCommandDTO) {
  graphSaving.value = true
  saveState.value = 'saving'
  graphResult.value = await saveProjectGraphLayout(command)
  const canvas = graphResult.value.canvas
  if (canvas) {
    activeGraphCanvas.value = canvas
    canvasGrid.value = canvas.grid
    graphViewportLabel.value = `${Math.round(canvas.viewport.zoom * 100)}%`
  }
  if (graphResult.value.ok && canvas) {
    saveState.value = 'saved'
    lastSavedAt.value = new Date().toLocaleTimeString([], {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    })
  } else {
    saveState.value = 'failed'
    activeInspectorTab.value = 'tasks'
  }
  graphSaving.value = false
}

function saveCurrentCanvasLayout() {
  graphCanvasRef.value?.saveLayout()
}

function updateCanvasViewport(viewport: { zoom: number }) {
  graphViewportLabel.value = `${Math.round(viewport.zoom * 100)}%`
}

function toggleCanvasGrid() {
  canvasGrid.value = {
    ...canvasGrid.value,
    visible: !canvasGrid.value.visible,
  }
}

async function copyInspectorSummary() {
  copiedInspectorText.value = inspectorSummary.value
  try {
    await navigator.clipboard?.writeText(inspectorSummary.value)
  } catch {
    // Clipboard permissions are browser dependent; the visible copied state still confirms the action.
  }
}

function markScriptDirty() {
  scriptDirty.value = true
}

function hydrateScriptForm(document: ScriptDocumentDTO) {
  scriptDocument.value = document
  scriptTitle.value = document.title
  scriptSourceAssetId.value = document.sourceAssetId || scriptSourceAssetId.value || DEFAULT_SCRIPT_SOURCE_ASSET_ID
  scriptRawText.value = document.rawText
  scriptLogline.value = document.logline || ''
  scriptSynopsis.value = document.synopsis || ''
  scriptDirty.value = false
}

async function saveCurrentScriptDocument() {
  scriptSaving.value = true
  scriptResult.value = await saveScriptDocument({
    scriptId: scriptDocument.value?.id || DEFAULT_SCRIPT_DOCUMENT_ID,
    title: scriptTitle.value,
    sourceAssetId: scriptSourceAssetId.value,
    rawText: scriptRawText.value,
    logline: scriptLogline.value,
    synopsis: scriptSynopsis.value,
  })
  if (scriptResult.value.document) {
    hydrateScriptForm(scriptResult.value.document)
    await loadCandidateRows()
  } else {
    scriptDirty.value = true
  }
  activeInspectorTab.value = 'tasks'
  scriptSaving.value = false
}

async function importScriptSourceAsset() {
  scriptSaving.value = true
  scriptResult.value = await saveScriptDocument({
    scriptId: scriptDocument.value?.id || DEFAULT_SCRIPT_DOCUMENT_ID,
    title: scriptTitle.value,
    sourceAssetId: scriptSourceAssetId.value,
    rawText: '',
    logline: scriptLogline.value,
    synopsis: scriptSynopsis.value,
  })
  if (scriptResult.value.document) {
    hydrateScriptForm(scriptResult.value.document)
    await loadCandidateRows()
  }
  activeInspectorTab.value = 'tasks'
  scriptSaving.value = false
}

async function loadCurrentScriptDocument() {
  scriptLoading.value = true
  scriptResult.value = await loadScriptDocument({
    scriptId: scriptDocument.value?.id || DEFAULT_SCRIPT_DOCUMENT_ID,
  })
  if (scriptResult.value.document) {
    hydrateScriptForm(scriptResult.value.document)
    await loadCandidateRows()
  }
  activeInspectorTab.value = 'tasks'
  scriptLoading.value = false
}

async function confirmCurrentScriptScene() {
  sceneSaving.value = true
  sceneResult.value = await confirmScriptScene({
    sceneId: sceneId.value,
    title: sceneTitle.value,
    location: sceneLocation.value,
    timeOfDay: sceneTimeOfDay.value,
    characters: splitList(sceneCharacters.value),
    props: splitList(sceneProps.value),
    action: sceneAction.value,
    emotionalBeat: sceneEmotionalBeat.value,
    sourceRange: { startLine: sceneStartLine.value, endLine: sceneEndLine.value },
    allowOverlap: sceneAllowOverlap.value,
  })
  ingestSceneCandidateResult(sceneResult.value)
  activeInspectorTab.value = 'tasks'
  sceneSaving.value = false
}

async function saveCurrentShotCandidate() {
  candidateSaving.value = true
  sceneResult.value = await saveShotCandidate({
    candidateId: candidateId.value,
    scriptSceneId: activeSceneId.value,
    index: candidateIndex.value,
    durationSeconds: candidateDuration.value,
    visualDescription: candidateDescription.value,
    characterRefs: splitList(candidateCharacters.value).map((name) => ({ name })),
    sourceRange: { startLine: candidateStartLine.value, endLine: candidateEndLine.value },
  })
  ingestSceneCandidateResult(sceneResult.value)
  activeInspectorTab.value = 'tasks'
  candidateSaving.value = false
}

async function loadCandidateRows() {
  candidateAction.value = 'load'
  sceneResult.value = await listShotCandidates(undefined, scriptDocument.value?.id || DEFAULT_SCRIPT_DOCUMENT_ID)
  ingestSceneCandidateResult(sceneResult.value)
  candidateAction.value = ''
}

async function confirmCandidateRow(candidate: ShotCandidateDTO) {
  candidateAction.value = candidate.id
  sceneResult.value = await confirmShotCandidate({
    candidateId: candidate.id,
    confirmedBy: 'local_user',
  })
  ingestSceneCandidateResult(sceneResult.value)
  activeInspectorTab.value = 'tasks'
  candidateAction.value = ''
}

async function rejectCandidateRow(candidate: ShotCandidateDTO) {
  candidateAction.value = candidate.id
  sceneResult.value = await rejectShotCandidate({
    candidateId: candidate.id,
    rejectionReason: 'manual reject',
  })
  ingestSceneCandidateResult(sceneResult.value)
  activeInspectorTab.value = 'tasks'
  candidateAction.value = ''
}

function ingestSceneCandidateResult(result: ScriptSceneCandidateResultDTO) {
  if (result.document) {
    hydrateScriptForm(result.document)
  }
  shotCandidates.value = nextShotCandidatesAfterResult(shotCandidates.value, result)
}

function splitList(value: string): string[] {
  return value.split(',').map((item) => item.trim()).filter(Boolean)
}
</script>

<template>
  <main class="workbench" :class="`theme-${themeMode}`">
    <Layout class="workbench-layout">
      <LayoutHeader class="top-bar">
        <section class="project-summary" aria-label="Project summary">
          <div class="project-mark" aria-hidden="true">
            <AppstoreOutlined />
          </div>
          <div class="project-copy">
            <p class="eyebrow">Alpha shell workbench</p>
            <h1 title="Tuyu Studio Alpha Project">
              Tuyu Studio Alpha Project
            </h1>
          </div>
        </section>

        <section class="top-status" aria-label="Workspace status">
          <Tag color="processing">mock_local</Tag>
          <Tag :color="saveBadge.tone" :title="saveBadge.detail">
            {{ saveBadge.label }}
          </Tag>
          <Badge :status="graphCanvas ? 'success' : 'default'" :text="`Graph ${graphCanvas?.version ?? '-'}`" />
          <Tag :color="healthStatusTone">Health {{ healthStatus }}</Tag>
          <Badge status="processing" :text="`Queue ${runtimeEvents.length + (projectResult?.events.length || 0) + (scriptResult?.events.length || 0) + (sceneResult?.events.length || 0)}`" />
          <Segmented
            v-model:value="themeMode"
            class="theme-switch"
            size="small"
            aria-label="Canvas theme"
            :options="[
              { label: 'Dark', value: 'dark' },
              { label: 'Warm', value: 'warm' },
            ]"
          />
          <Tooltip title="Save Canvas layout">
            <Button
              size="small"
              aria-label="Save Canvas layout"
              :loading="graphSaving"
              :disabled="!graphCanvas"
              @click="saveCurrentCanvasLayout"
            >
              <template #icon>
                <SaveOutlined />
              </template>
            </Button>
          </Tooltip>
          <Tooltip title="Export handoff package">
            <Button size="small" type="primary" aria-label="Export handoff package">
              <template #icon>
                <CloudUploadOutlined />
              </template>
              Export
            </Button>
          </Tooltip>
          <Tooltip title="Open project settings">
            <Button size="small" aria-label="Open project settings">
              <template #icon>
                <SettingOutlined />
              </template>
            </Button>
          </Tooltip>
        </section>
      </LayoutHeader>

      <Layout class="workbench-body">
        <LayoutSider class="left-panel" width="280" theme="light">
          <Menu
            class="rail-menu"
            mode="inline"
            :selected-keys="[activeRailTab]"
            :items="railItems"
            @click="({ key }) => (activeRailTab = String(key))"
          />

          <Tabs v-model:activeKey="activeRailTab" size="small" class="rail-tabs">
            <TabPane key="assets" tab="Assets">
              <div class="rail-toolbar">
                <Button size="small" type="primary">
                  <template #icon>
                    <FolderOpenOutlined />
                  </template>
                  Import
                </Button>
                <Button size="small">
                  <template #icon>
                    <SearchOutlined />
                  </template>
                  Filter
                </Button>
              </div>
              <List class="asset-list" item-layout="horizontal" size="small" :data-source="assets">
                <template #renderItem="{ item }">
                  <ListItem>
                    <ListItemMeta>
                      <template #avatar>
                        <div class="asset-thumb">{{ item.type }}</div>
                      </template>
                      <template #title>
                        <strong class="asset-title" :title="item.name">{{ item.name }}</strong>
                      </template>
                      <template #description>
                        <span class="asset-status">{{ item.status }}</span>
                      </template>
                    </ListItemMeta>
                  </ListItem>
                </template>
              </List>
            </TabPane>
            <TabPane key="script" tab="Script">
              <div class="script-editor">
                <div class="script-editor-bar">
                  <Tag :color="scriptDirty ? 'warning' : scriptResult?.ok ? 'success' : 'default'">
                    {{ scriptDocument?.id || DEFAULT_SCRIPT_DOCUMENT_ID }}
                  </Tag>
                  <Tag>{{ scriptByteCount }} bytes</Tag>
                </div>

                <label class="script-field">
                  <span>Title</span>
                  <input
                    v-model="scriptTitle"
                    class="script-input"
                    type="text"
                    autocomplete="off"
                    @input="markScriptDirty"
                  >
                </label>

                <label class="script-field">
                  <span>Source asset</span>
                  <input
                    v-model="scriptSourceAssetId"
                    class="script-input"
                    type="text"
                    autocomplete="off"
                    @input="markScriptDirty"
                  >
                </label>

                <div class="script-actions">
                  <Button size="small" :loading="scriptSaving" :disabled="!scriptSourceAssetId.trim()" @click="importScriptSourceAsset">
                    <template #icon>
                      <FolderOpenOutlined />
                    </template>
                    Import asset
                  </Button>
                  <Button size="small" :loading="scriptLoading" @click="loadCurrentScriptDocument">
                    <template #icon>
                      <ReloadOutlined />
                    </template>
                    Load
                  </Button>
                  <Button size="small" type="primary" :loading="scriptSaving" :disabled="!scriptRawText.trim()" @click="saveCurrentScriptDocument">
                    <template #icon>
                      <SaveOutlined />
                    </template>
                    Save
                  </Button>
                </div>

                <label class="script-field">
                  <span>Logline</span>
                  <textarea
                    v-model="scriptLogline"
                    class="script-textarea script-textarea--short"
                    spellcheck="false"
                    @input="markScriptDirty"
                  />
                </label>

                <label class="script-field">
                  <span>Synopsis</span>
                  <textarea
                    v-model="scriptSynopsis"
                    class="script-textarea script-textarea--short"
                    spellcheck="false"
                    @input="markScriptDirty"
                  />
                </label>

                <label class="script-field">
                  <span>Raw text</span>
                  <textarea
                    v-model="scriptRawText"
                    class="script-textarea script-textarea--raw"
                    spellcheck="false"
                    @input="markScriptDirty"
                  />
                </label>

                <Alert
                  v-if="scriptResult || scriptDirty"
                  class="service-alert"
                  :type="scriptStatusType"
                  show-icon
                  :message="scriptResult?.error?.userMessage || scriptStatusLabel"
                  :description="scriptStatusLabel"
                />
                <List
                  v-if="scriptRecoveryActions.length"
                  class="recovery-list"
                  size="small"
                  :data-source="scriptRecoveryActions"
                >
                  <template #renderItem="{ item }">
                    <ListItem>
                      <span>{{ item }}</span>
                    </ListItem>
                  </template>
                </List>

                <div class="scene-editor">
                  <div class="script-editor-bar">
                    <Tag>{{ scriptScenes.length }} scenes</Tag>
                    <Tag>{{ candidateSummary }}</Tag>
                  </div>

                  <label class="script-field">
                    <span>Scene id</span>
                    <input v-model="sceneId" class="script-input" type="text" autocomplete="off">
                  </label>
                  <label class="script-field">
                    <span>Scene title</span>
                    <input v-model="sceneTitle" class="script-input" type="text" autocomplete="off">
                  </label>
                  <label class="script-field">
                    <span>Location</span>
                    <input v-model="sceneLocation" class="script-input" type="text" autocomplete="off">
                  </label>
                  <div class="line-range-grid">
                    <label class="script-field">
                      <span>Start line</span>
                      <input v-model.number="sceneStartLine" class="script-input" type="number" min="1">
                    </label>
                    <label class="script-field">
                      <span>End line</span>
                      <input v-model.number="sceneEndLine" class="script-input" type="number" min="1">
                    </label>
                  </div>
                  <label class="script-field">
                    <span>Characters</span>
                    <input v-model="sceneCharacters" class="script-input" type="text" autocomplete="off">
                  </label>
                  <label class="script-field">
                    <span>Props</span>
                    <input v-model="sceneProps" class="script-input" type="text" autocomplete="off">
                  </label>
                  <label class="script-field">
                    <span>Action</span>
                    <textarea v-model="sceneAction" class="script-textarea script-textarea--short" spellcheck="false" />
                  </label>
                  <label class="script-field">
                    <span>Emotional beat</span>
                    <input v-model="sceneEmotionalBeat" class="script-input" type="text" autocomplete="off">
                  </label>
                  <label class="check-row">
                    <input v-model="sceneAllowOverlap" type="checkbox">
                    <span>Allow overlap</span>
                  </label>
                  <div class="script-actions">
                    <Button size="small" type="primary" :loading="sceneSaving" :disabled="!scriptRawText.trim()" @click="confirmCurrentScriptScene">
                      Confirm scene
                    </Button>
                    <Button size="small" :loading="candidateAction === 'load'" @click="loadCandidateRows">
                      Load candidates
                    </Button>
                  </div>
                </div>

                <div class="scene-editor">
                  <label class="script-field">
                    <span>Candidate id</span>
                    <input v-model="candidateId" class="script-input" type="text" autocomplete="off">
                  </label>
                  <div class="line-range-grid">
                    <label class="script-field">
                      <span>Shot index</span>
                      <input v-model.number="candidateIndex" class="script-input" type="number" min="1">
                    </label>
                    <label class="script-field">
                      <span>Seconds</span>
                      <input v-model.number="candidateDuration" class="script-input" type="number" min="1">
                    </label>
                  </div>
                  <div class="line-range-grid">
                    <label class="script-field">
                      <span>Start line</span>
                      <input v-model.number="candidateStartLine" class="script-input" type="number" min="1">
                    </label>
                    <label class="script-field">
                      <span>End line</span>
                      <input v-model.number="candidateEndLine" class="script-input" type="number" min="1">
                    </label>
                  </div>
                  <label class="script-field">
                    <span>Character refs</span>
                    <input v-model="candidateCharacters" class="script-input" type="text" autocomplete="off">
                  </label>
                  <label class="script-field">
                    <span>Visual description</span>
                    <textarea v-model="candidateDescription" class="script-textarea script-textarea--short" spellcheck="false" />
                  </label>
                  <div class="script-actions">
                    <Button size="small" type="primary" :loading="candidateSaving" :disabled="!activeSceneId || !candidateDescription.trim()" @click="saveCurrentShotCandidate">
                      Save candidate
                    </Button>
                  </div>
                  <Alert
                    v-if="sceneResult"
                    class="service-alert"
                    :type="sceneResult.ok ? 'success' : 'error'"
                    show-icon
                    :message="sceneResult.error?.userMessage || candidateSummary"
                    :description="sceneResult.error ? `${sceneResult.error.code} · ${sceneResult.error.correlationId}` : latestEventLabel"
                  />
                  <List
                    v-if="sceneRecoveryActions.length"
                    class="recovery-list"
                    size="small"
                    :data-source="sceneRecoveryActions"
                  >
                    <template #renderItem="{ item }">
                      <ListItem>
                        <span>{{ item }}</span>
                      </ListItem>
                    </template>
                  </List>
                </div>
              </div>
            </TabPane>
            <TabPane key="blueprint" tab="Blueprint">
              <div class="placeholder-list">
                <strong>Production templates</strong>
                <span>Blueprint insertion is reserved for the Canvas slice.</span>
              </div>
            </TabPane>
            <TabPane key="search" tab="Search">
              <div class="placeholder-list">
                <strong>Project search</strong>
                <span>Search results will route to Canvas and Inspector later.</span>
              </div>
            </TabPane>
            <TabPane key="history" tab="History">
              <div class="placeholder-list">
                <strong>Run history</strong>
                <span>Run/Event/Audit records are introduced after the API slice.</span>
              </div>
            </TabPane>
          </Tabs>
        </LayoutSider>

        <LayoutContent class="canvas-shell" aria-label="Project Canvas">
          <section class="canvas-toolbar" aria-label="Canvas actions">
            <div>
              <p class="eyebrow">Project Canvas</p>
              <h2>{{ graphCanvas?.id || 'Creative graph' }}</h2>
            </div>
            <div class="canvas-actions">
              <Tooltip title="Fit visible nodes">
                <Button size="small" :disabled="!graphCanvas" @click="graphCanvasRef?.fitView()">
                  Fit
                </Button>
              </Tooltip>
              <Tooltip title="Toggle Canvas grid">
                <Button
                  size="small"
                  aria-label="Toggle Canvas grid"
                  :disabled="!graphCanvas"
                  @click="toggleCanvasGrid"
                >
                  <template #icon>
                    <BgColorsOutlined />
                  </template>
                </Button>
              </Tooltip>
              <Tooltip title="Move selected node">
                <Button
                  size="small"
                  aria-label="Move selected node"
                  :disabled="!graphCanvas"
                  @click="graphCanvasRef?.nudgeSelected()"
                >
                  <template #icon>
                    <ThunderboltOutlined />
                  </template>
                </Button>
              </Tooltip>
              <Tooltip title="Focus selected node">
                <Button
                  size="small"
                  aria-label="Focus selected node"
                  :disabled="selectedGraphNodes.length === 0"
                  @click="graphCanvasRef?.focusSelected()"
                >
                  <template #icon>
                    <ShareAltOutlined />
                  </template>
                </Button>
              </Tooltip>
            </div>
          </section>

          <section class="canvas-stage">
            <GraphCanvas
              ref="graphCanvasRef"
              :canvas="graphCanvas"
              :theme="graphTheme"
              :grid="canvasGrid"
              :loading="graphLoading"
              :saving="graphSaving"
              @save="saveCanvasLayout"
              @select="(node) => (selectedGraphNode = node)"
              @selection="(nodes) => (selectedGraphNodes = nodes)"
              @viewport="updateCanvasViewport"
            />
            <section v-if="shotCandidates.length" class="script-expansion-overlay" aria-label="Script expansion table">
              <div class="script-expansion-header">
                <div>
                  <p class="eyebrow">Script expansion</p>
                  <h2>{{ activeSceneId || 'Scene candidates' }}</h2>
                </div>
                <Tag>{{ candidateSummary }}</Tag>
              </div>
              <div class="script-expansion-scroll">
                <table class="script-expansion-table">
                  <thead>
                    <tr>
                      <th>Shot</th>
                      <th>Seconds</th>
                      <th>Visual</th>
                      <th>Characters</th>
                      <th>Range</th>
                      <th>Status</th>
                      <th>Action</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="candidate in shotCandidates" :key="candidate.id">
                      <td>{{ candidate.index }}</td>
                      <td>{{ candidate.durationSeconds }}</td>
                      <td :title="candidate.visualDescription">{{ candidate.visualDescription }}</td>
                      <td>{{ candidate.characterRefs.map((ref) => ref.name || ref.characterId).join(', ') || '-' }}</td>
                      <td>{{ candidate.sourceRange.startLine }}-{{ candidate.sourceRange.endLine }}</td>
                      <td><Tag>{{ candidate.status }}</Tag></td>
                      <td>
                        <div class="table-actions">
                          <Button size="small" :loading="candidateAction === candidate.id" :disabled="candidate.status === 'accepted'" @click="confirmCandidateRow(candidate)">
                            Confirm
                          </Button>
                          <Button size="small" :disabled="candidate.status === 'accepted'" @click="rejectCandidateRow(candidate)">
                            Reject
                          </Button>
                        </div>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </section>
          </section>
        </LayoutContent>

        <LayoutSider class="inspector" width="392" theme="light">
          <section class="inspector-header">
            <div class="inspector-title">
              <p class="eyebrow">Inspector</p>
              <h2 :title="selectedGraphNode?.title || 'Canvas selection'">
                {{ selectedGraphNode?.title || 'Canvas selection' }}
              </h2>
            </div>
            <div class="inspector-actions">
              <Tag :color="selectedGraphNode ? 'processing' : 'default'">
                {{ selectedGraphNode?.kind || 'canvas' }}
              </Tag>
              <Tooltip title="Focus selected node">
                <Button
                  size="small"
                  aria-label="Focus selected node"
                  :disabled="selectedGraphNodeIDs.length === 0"
                  @click="graphCanvasRef?.focusSelected()"
                >
                  <template #icon>
                    <ShareAltOutlined />
                  </template>
                </Button>
              </Tooltip>
              <Tooltip title="Copy selection summary">
                <Button size="small" aria-label="Copy selection summary" @click="copyInspectorSummary">
                  <template #icon>
                    <CopyOutlined />
                  </template>
                </Button>
              </Tooltip>
            </div>
          </section>

          <div class="status-badge-row" aria-label="Selection status badges">
            <Tag
              v-for="badge in selectedNodeBadges"
              :key="badge.key"
              :color="badge.tone"
              :title="badge.detail"
              class="status-badge"
            >
              {{ badge.label }}
            </Tag>
            <Tag v-if="!selectedNodeBadges.length" color="default">Canvas</Tag>
          </div>

          <Tabs v-model:activeKey="activeInspectorTab" size="small" class="inspector-tabs" :tab-bar-gutter="10">
            <TabPane key="properties" tab="Props">
              <div class="inspector-tab-body">
                <dl class="property-grid">
                  <div>
                    <dt>Type</dt>
                    <dd>{{ selectedGraphNode?.kind || 'Canvas' }}</dd>
                  </div>
                  <div>
                    <dt>Status</dt>
                    <dd>{{ selectedGraphNode?.status || healthStatus }}</dd>
                  </div>
                  <div>
                    <dt>Source</dt>
                    <dd>{{ selectedGraphNode?.source || graphCanvas?.projectId || '-' }}</dd>
                  </div>
                  <div>
                    <dt>Reference</dt>
                    <dd :title="selectedGraphNode?.refId || ''">
                      {{ selectedGraphNode?.refId || '-' }}
                    </dd>
                  </div>
                  <div v-if="selectedGraphNode?.sourceEventId">
                    <dt>Source event</dt>
                    <dd>{{ selectedGraphNode.sourceEventId }}</dd>
                  </div>
                  <div v-if="selectedGraphNode">
                    <dt>Canvas</dt>
                    <dd>{{ selectedGraphNode.position.x }}, {{ selectedGraphNode.position.y }} · {{ selectedGraphNode.size.width }}x{{ selectedGraphNode.size.height }}</dd>
                  </div>
                </dl>

                <List
                  v-if="selectedNodeDataRows.length"
                  class="detail-list"
                  size="small"
                  :data-source="selectedNodeDataRows"
                >
                  <template #renderItem="{ item }">
                    <ListItem>
                      <strong>{{ item.label }}</strong>
                      <span :title="item.value">{{ item.value }}</span>
                    </ListItem>
                  </template>
                </List>
              </div>
            </TabPane>

            <TabPane key="relations" tab="Links">
              <div class="inspector-tab-body">
                <List class="relation-list" size="small" :data-source="selectedRelations">
                  <template #renderItem="{ item }">
                    <ListItem>
                      <div class="relation-row">
                        <strong :title="item.label">{{ item.label }}</strong>
                        <span>{{ item.direction }} · {{ item.peerTitle }}</span>
                        <span v-if="item.validity !== 'valid'" class="relation-detail">
                          {{ item.detail }}
                        </span>
                      </div>
                      <Tag :color="item.tone">{{ item.relation }}</Tag>
                    </ListItem>
                  </template>
                </List>
                <div v-if="!selectedRelations.length" class="placeholder-list">
                  <strong>No selected relations</strong>
                  <span>{{ selectedGraphNode ? selectedGraphNode.id : graphCanvas?.projectId || 'Canvas' }}</span>
                </div>
              </div>
            </TabPane>

            <TabPane key="continuity" tab="Rules">
              <div class="inspector-tab-body">
                <List v-if="continuityRiskRows.length" class="risk-list" size="small" :data-source="continuityRiskRows">
                  <template #renderItem="{ item }">
                    <ListItem>
                      <div class="risk-row">
                        <strong>{{ item.label }}</strong>
                        <span>{{ item.detail }}</span>
                      </div>
                      <Tag :color="item.tone">{{ item.tone }}</Tag>
                    </ListItem>
                  </template>
                </List>
                <div v-else class="placeholder-list">
                  <strong>Continuity ready</strong>
                  <span>{{ selectedGraphNode?.title || graphCanvas?.id || 'Canvas' }}</span>
                </div>
              </div>
            </TabPane>

            <TabPane key="tasks" tab="Tasks">
              <div class="inspector-tab-body">
                <section class="task-preview">
                  <strong>{{ selectedGraphNode?.kind || 'project' }} context</strong>
                  <span>{{ inspectorSummary }}</span>
                  <div class="status-badge-row">
                    <Tag :color="saveBadge.tone" :title="saveBadge.detail">{{ saveBadge.label }}</Tag>
                    <Tag :color="healthStatusTone">Health {{ healthStatus }}</Tag>
                  </div>
                </section>

                <List class="queue-list" size="small" :data-source="queueItems">
                  <template #renderItem="{ item }">
                    <ListItem>
                      <span class="queue-label">{{ item.label }}</span>
                      <Tag>{{ item.value }}</Tag>
                    </ListItem>
                  </template>
                </List>

                <section class="service-panel" aria-label="Go service probe">
                  <div class="service-actions">
                    <Button
                      size="small"
                      type="primary"
                      :loading="probeLoading"
                      @click="runProbe('status')"
                    >
                      Probe Go service
                    </Button>
                    <Button
                      size="small"
                      danger
                      :loading="probeLoading"
                      @click="runProbe('structured_error')"
                    >
                      Show structured error
                    </Button>
                  </div>

                  <div class="service-actions">
                    <Button
                      size="small"
                      :loading="projectLoading === 'create'"
                      @click="runProjectAction('create')"
                    >
                      <template #icon>
                        <AppstoreOutlined />
                      </template>
                      Create project
                    </Button>
                    <Button
                      size="small"
                      :loading="projectLoading === 'open'"
                      @click="runProjectAction('open')"
                    >
                      <template #icon>
                        <FolderOpenOutlined />
                      </template>
                      Open project
                    </Button>
                    <Button
                      size="small"
                      :loading="projectLoading === 'save'"
                      @click="runProjectAction('save')"
                    >
                      <template #icon>
                        <SaveOutlined />
                      </template>
                      Save project
                    </Button>
                    <Button
                      size="small"
                      :loading="projectLoading === 'health'"
                      @click="runProjectAction('health')"
                    >
                      <template #icon>
                        <SearchOutlined />
                      </template>
                      Health check
                    </Button>
                    <Button
                      size="small"
                      :loading="graphLoading"
                      @click="loadProjectGraph()"
                    >
                      <template #icon>
                        <BranchesOutlined />
                      </template>
                      Load graph
                    </Button>
                  </div>

                  <Alert
                    v-if="probeSnapshot"
                    class="service-alert"
                    type="success"
                    show-icon
                    :message="probeSnapshot.summary"
                    :description="`${probeSnapshot.serviceName} · ${probeSnapshot.status} · ${probeSnapshot.checkedAt}`"
                  />
                  <div v-if="probeSnapshot" class="capability-row">
                    <Tag v-for="capability in probeSnapshot.capabilities" :key="capability">
                      {{ capability }}
                    </Tag>
                  </div>

                  <Alert
                    v-if="probeError"
                    class="service-alert"
                    :type="probeError.severity === 'blocking' ? 'error' : 'warning'"
                    show-icon
                    :message="probeError.userMessage"
                    :description="`${probeError.code} · ${probeError.severity} · ${probeError.correlationId}`"
                  />
                  <List
                    v-if="probeError"
                    class="recovery-list"
                    size="small"
                    :data-source="probeError.recoveryActions"
                  >
                    <template #renderItem="{ item }">
                      <ListItem>
                        <span>{{ item }}</span>
                      </ListItem>
                    </template>
                  </List>

                  <Alert
                    v-if="projectResult?.summary"
                    class="service-alert"
                    :type="projectResult.ok ? 'success' : 'warning'"
                    show-icon
                    :message="projectResult.summary.name"
                    :description="`${projectResult.summary.openMode} · ${projectResult.summary.lockState} · graph ${projectResult.summary.graphVersion}`"
                  />
                  <Alert
                    v-if="projectResult?.health"
                    class="service-alert"
                    :type="projectResult.health.status === 'blocking' ? 'error' : projectResult.health.status === 'warning' ? 'warning' : 'success'"
                    show-icon
                    :message="`Health ${projectResult.health.status}`"
                    :description="`${projectResult.health.items.length} items · ${projectResult.health.checkedAt}`"
                  />
                  <Alert
                    v-if="projectResult?.error"
                    class="service-alert"
                    :type="projectResult.error.severity === 'blocking' ? 'error' : 'warning'"
                    show-icon
                    :message="projectResult.error.userMessage"
                    :description="`${projectResult.error.code} · ${projectResult.error.correlationId}`"
                  />
                  <Alert
                    v-if="graphResult"
                    class="service-alert"
                    :type="canvasHealthType"
                    show-icon
                    :message="graphResult.ok ? 'Graph View loaded' : graphResult.error?.userMessage"
                    :description="graphResult.canvas ? `${graphResult.canvas.nodes.length} nodes · ${graphResult.canvas.edges.length} edges · graph ${graphResult.canvas.version}` : graphResult.error?.code"
                  />
                </section>
              </div>
            </TabPane>

            <TabPane key="runs" tab="Runs">
              <div class="inspector-tab-body">
                <List v-if="runtimeEvents.length" class="event-list" size="small" :data-source="runtimeEvents">
                  <template #renderItem="{ item }">
                    <ListItem>
                      <ListItemMeta>
                        <template #title>
                          <strong class="event-title">
                            {{ item.eventType }} · {{ item.state }} · {{ item.progress }}%
                          </strong>
                        </template>
                        <template #description>
                          <span>{{ item.summary }} · {{ item.createdAt }}</span>
                        </template>
                      </ListItemMeta>
                    </ListItem>
                  </template>
                </List>
                <List v-if="projectResult?.events.length" class="event-list" size="small" :data-source="projectResult.events">
                  <template #renderItem="{ item }">
                    <ListItem>
                      <ListItemMeta>
                        <template #title>
                          <strong class="event-title">
                            {{ item.eventType }} · {{ item.state }}
                          </strong>
                        </template>
                        <template #description>
                          <span>{{ item.summary }} · {{ item.createdAt }}</span>
                        </template>
                      </ListItemMeta>
                    </ListItem>
                  </template>
                </List>
                <List v-if="scriptResult?.events.length" class="event-list" size="small" :data-source="scriptResult.events">
                  <template #renderItem="{ item }">
                    <ListItem>
                      <ListItemMeta>
                        <template #title>
                          <strong class="event-title">
                            {{ item.eventType }} · {{ item.state }}
                          </strong>
                        </template>
                        <template #description>
                          <span>{{ item.summary }} · {{ item.createdAt }}</span>
                        </template>
                      </ListItemMeta>
                    </ListItem>
                  </template>
                </List>
                <List v-if="sceneResult?.events.length" class="event-list" size="small" :data-source="sceneResult.events">
                  <template #renderItem="{ item }">
                    <ListItem>
                      <ListItemMeta>
                        <template #title>
                          <strong class="event-title">
                            {{ item.eventType }} · {{ item.state }}
                          </strong>
                        </template>
                        <template #description>
                          <span>{{ item.summary }} · {{ item.createdAt }}</span>
                        </template>
                      </ListItemMeta>
                    </ListItem>
                  </template>
                </List>
                <div v-if="!runtimeEvents.length && !projectResult?.events.length && !scriptResult?.events.length && !sceneResult?.events.length" class="placeholder-list">
                  <strong>No run records</strong>
                  <span>{{ latestEventLabel }}</span>
                </div>
              </div>
            </TabPane>

            <TabPane key="audit" tab="Audit">
              <div class="inspector-tab-body">
                <dl class="property-grid">
                  <div>
                    <dt>Selection</dt>
                    <dd>{{ selectedSummary }}</dd>
                  </div>
                  <div>
                    <dt>Save</dt>
                    <dd>{{ saveBadge.label }}</dd>
                  </div>
                  <div>
                    <dt>Latest event</dt>
                    <dd>{{ latestEventLabel }}</dd>
                  </div>
                  <div v-if="copiedInspectorText">
                    <dt>Copied</dt>
                    <dd>{{ copiedInspectorText }}</dd>
                  </div>
                </dl>

                <Alert
                  v-if="errorSummary"
                  class="service-alert"
                  :type="errorSummary.severity === 'blocking' || errorSummary.severity === 'error' ? 'error' : 'warning'"
                  show-icon
                  :message="errorSummary.message"
                  :description="`${errorSummary.code} · retryable ${errorSummary.retryable ? 'yes' : 'no'}`"
                />
                <List
                  v-if="graphErrors.length"
                  class="recovery-list"
                  size="small"
                  :data-source="graphErrors"
                >
                  <template #renderItem="{ item }">
                    <ListItem>
                      <span>{{ item.code }} · {{ item.userMessage }}</span>
                    </ListItem>
                  </template>
                </List>
                <List
                  v-if="projectResult?.health?.items.length"
                  class="recovery-list"
                  size="small"
                  :data-source="projectResult.health.items"
                >
                  <template #renderItem="{ item }">
                    <ListItem>
                      <div class="health-item">
                        <strong>
                          {{ item.code }} · {{ item.severity }}
                        </strong>
                        <span>{{ item.userMessage }}</span>
                        <span v-if="item.path">Path {{ item.path }}</span>
                        <span v-if="item.affectedObjects.length">
                          Affected {{ item.affectedObjects.join(', ') }}
                        </span>
                        <span v-if="item.technicalDetail">
                          Detail {{ item.technicalDetail }}
                        </span>
                        <List
                          v-if="item.recoveryActions.length"
                          class="health-actions"
                          size="small"
                          :data-source="item.recoveryActions"
                        >
                          <template #renderItem="{ item: action }">
                            <ListItem>
                              <span>{{ action }}</span>
                            </ListItem>
                          </template>
                        </List>
                      </div>
                    </ListItem>
                  </template>
                </List>
              </div>
            </TabPane>
          </Tabs>
        </LayoutSider>
      </Layout>

      <footer class="bottom-bar" aria-label="Workbench status">
        <div class="bottom-group">
          <Badge status="processing" :text="selectedSummary" />
          <span>Zoom {{ graphViewportLabel }}</span>
          <span>Grid {{ canvasGrid.visible ? `${canvasGrid.size}px` : 'off' }}</span>
        </div>
        <div class="bottom-group">
          <Tag :color="saveBadge.tone" :title="saveBadge.detail">{{ saveBadge.label }}</Tag>
          <Tag :color="healthStatusTone">Health {{ healthStatus }}</Tag>
          <span v-if="errorSummary" :title="errorSummary.message">
            {{ errorSummary.code }}
          </span>
          <span>Graph events {{ graphResult?.events.length || 0 }}</span>
          <span :title="latestEventLabel">{{ latestEventLabel }}</span>
          <Progress class="queue-progress" :percent="runtimeEvents[0]?.progress ?? 0" size="small" />
        </div>
        <Button size="small" :disabled="graphLoading" @click="loadProjectGraph()">
          <template #icon>
            <ReloadOutlined />
          </template>
          Reload
        </Button>
      </footer>
    </Layout>
  </main>
</template>
