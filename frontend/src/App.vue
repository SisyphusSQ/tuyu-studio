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
  type AssetBindingDTO,
  type AssetBindingDuplicatePolicy,
  type AssetBindingTargetType,
  type AssetDTO,
  type AssetDuplicatePolicy,
  type AssetLibraryResultDTO,
  type CanvasGridDTO,
  type CanvasTheme,
  type ContinuityLibraryResultDTO,
  type ContinuityResultDTO,
  type ContinuityRuleSeverity,
  type GenerationPackageResultDTO,
  type GraphNodeDTO,
  type MockRunResultDTO,
  type ProfileDTO,
  type ProjectCanvasDTO,
  type ProjectGraphLayoutSaveCommandDTO,
  type ProjectGraphViewResultDTO,
  type ProjectOperationName,
  type ProjectOperationResultDTO,
  type ResultDuplicatePolicy,
  type ResultReviewResultDTO,
  type ResultReviewStatus,
  type RuntimeEventDTO,
  type ScriptSceneCandidateResultDTO,
  type ScriptDocumentDTO,
  type ScriptDocumentResultDTO,
  type ShotCandidateDTO,
  type ShotContextResultDTO,
  type WorkbenchProbeMode,
  type WorkbenchStatusDTO,
} from './api/dto'
import {
  DEFAULT_ASSET_IMPORT_ROLE,
  DEFAULT_ASSET_IMPORT_SOURCE,
  importAsset,
  listAssets,
} from './api/assets'
import {
  bindAsset,
  DEFAULT_BINDING_PURPOSE,
  DEFAULT_BINDING_TARGET_ID,
  DEFAULT_BINDING_TARGET_TYPE,
  listAssetBindings,
  setMainReference,
} from './api/assetBindings'
import {
  listContinuity,
  saveContinuityRule,
  unlockAssetBinding,
  unlockContinuityRule,
} from './api/continuity'
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
import {
  markShotContextDirty,
  promoteShotContext,
  validateShotContext,
} from './api/shotContext'
import { exportGenerationPackage } from './api/packageExport'
import { cancelMockRun, retryMockRun, startMockRun } from './api/mockRun'
import {
  DEFAULT_RESULT_REBIND_REASON,
  DEFAULT_RESULT_SHOT_ID,
  DEFAULT_RESULT_SOURCE_PATH,
  importResult,
  listResults,
  rebindResult,
  traceResult,
  updateResultReview,
} from './api/resultReview'
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
const assetRows = ref<AssetDTO[]>([])
const assetResult = ref<AssetLibraryResultDTO>()
const assetLoading = ref(false)
const assetImporting = ref(false)
const assetImportPath = ref(DEFAULT_ASSET_IMPORT_SOURCE)
const assetImportRole = ref(DEFAULT_ASSET_IMPORT_ROLE)
const assetDuplicatePolicy = ref<AssetDuplicatePolicy>('cancel')
const assetManagedReference = ref(false)
const bindingResult = ref<ContinuityLibraryResultDTO>()
const continuityResult = ref<ContinuityResultDTO>()
const bindingLoading = ref(false)
const bindingSubmitting = ref(false)
const mainReferenceSaving = ref(false)
const continuityAction = ref('')
const bindingAssetId = ref('')
const bindingTargetType = ref<AssetBindingTargetType>(DEFAULT_BINDING_TARGET_TYPE)
const bindingTargetId = ref(DEFAULT_BINDING_TARGET_ID)
const bindingPurpose = ref(DEFAULT_BINDING_PURPOSE)
const bindingDuplicatePolicy = ref<AssetBindingDuplicatePolicy>('cancel')
const mainReferenceAssetId = ref('')
const mainReferenceClear = ref(false)
const continuityRuleId = ref('rule_char_mina_raincoat')
const continuityRuleText = ref('Mina keeps the yellow raincoat visible in rainy exterior shots.')
const continuityRuleSeverity = ref<ContinuityRuleSeverity>('blocking')
const continuityRuleLocked = ref(true)
const continuityUnlockReason = ref('manual continuity override')
const bindingUnlockReason = ref('manual binding override')
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
const shotContextResult = ref<ShotContextResultDTO>()
const shotContextAction = ref('')
const shotContextId = ref('shot_001')
const shotDirtyReason = ref('manual revision')
const generationPackageResult = ref<GenerationPackageResultDTO>()
const packageAction = ref('')
const packageShotId = ref('shot_002')
const mockRunResult = ref<MockRunResultDTO>()
const mockRunAction = ref('')
const resultReviewResult = ref<ResultReviewResultDTO>()
const resultAction = ref('')
const resultSourcePath = ref(DEFAULT_RESULT_SOURCE_PATH)
const resultShotId = ref(DEFAULT_RESULT_SHOT_ID)
const resultPackageId = ref('')
const resultRunId = ref('')
const resultDuplicatePolicy = ref<ResultDuplicatePolicy>('cancel')
const resultReviewStatus = ref<ResultReviewStatus>('approved')
const resultReviewNotes = ref('Looks usable for alpha review.')
const resultReviewReason = ref('alpha review decision')
const resultRebindShotId = ref('shot_001')
const resultRebindPackageId = ref('')
const resultUnbindShot = ref(false)
const resultUnbindPackage = ref(false)
const resultRebindReason = ref(DEFAULT_RESULT_REBIND_REASON)

const railItems = [
  { key: 'assets', icon: () => h(DatabaseOutlined), label: 'Assets' },
  { key: 'script', icon: () => h(FileTextOutlined), label: 'Script' },
  { key: 'blueprint', icon: () => h(BranchesOutlined), label: 'Blueprint' },
  { key: 'search', icon: () => h(SearchOutlined), label: 'Search' },
  { key: 'history', icon: () => h(HistoryOutlined), label: 'History' },
]

const queueItems = computed(() => [
  {
    label: 'Package export',
    value: generationPackageResult.value?.package?.generationPackageStatus ||
      generationPackageResult.value?.error?.code ||
      'idle',
  },
  {
    label: 'Mock run',
    value: mockRunResult.value?.run?.status ||
      mockRunResult.value?.error?.code ||
      'idle',
  },
  {
    label: 'Review import',
    value: resultReviewResult.value?.result?.status ||
      resultReviewResult.value?.error?.code ||
      'idle',
  },
])

const graphCanvas = computed(() => activeGraphCanvas.value)
const graphErrors = computed(() => graphResult.value?.errors || [])
const assetSummary = computed(() => {
  const missing = assetRows.value.filter((asset) => asset.missing).length
  const managed = assetRows.value.filter((asset) => asset.source.kind === 'managed_reference').length
  return `${assetRows.value.length} assets · ${missing} missing · ${managed} managed`
})
const assetRecoveryActions = computed(() => assetResult.value?.error?.recoveryActions || [])
const continuityProfiles = computed(() => bindingResult.value?.profiles || [])
const continuityLineage = computed(() => bindingResult.value?.lineage || [])
const continuityRules = computed(() => continuityResult.value?.rules || continuityProfiles.value.flatMap((profile) => profile.continuityRules))
const selectedContinuityProfile = computed(() => continuityProfiles.value.find((profile) => (
  profile.type === bindingTargetType.value && profile.id === bindingTargetId.value
)))
const selectedMainReferenceBinding = computed<AssetBindingDTO | undefined>(() => {
  const profile = selectedContinuityProfile.value
  if (!profile?.mainReferenceAssetId) {
    return undefined
  }
  return profile.bindings.find((binding) => (
    binding.purpose === 'main_reference' &&
    binding.targetType === profile.type &&
    binding.targetId === profile.id &&
    binding.assetId === profile.mainReferenceAssetId
  ))
})
const selectedBindingForUnlock = computed<AssetBindingDTO | undefined>(() => {
  for (const lineage of continuityLineage.value) {
    for (const binding of lineage.bindings) {
      if (
        binding.assetId === bindingAssetId.value &&
        binding.targetType === bindingTargetType.value &&
        binding.targetId === bindingTargetId.value &&
        (binding.purpose || 'reference') === bindingPurpose.value
      ) {
        return binding
      }
    }
  }
  return undefined
})
const bindingRecoveryActions = computed(() => bindingResult.value?.error?.recoveryActions || [])
const continuityRecoveryActions = computed(() => continuityResult.value?.error?.recoveryActions || continuityResult.value?.impact?.recoveryActions || [])
const assetStatusType = computed(() => {
  if (assetResult.value?.error?.severity === 'blocking' || assetResult.value?.error?.severity === 'error') {
    return 'error'
  }
  if (assetResult.value?.error || assetRows.value.some((asset) => asset.missing || asset.thumbnailStatus === 'thumbnail_failed')) {
    return 'warning'
  }
  return assetResult.value?.ok ? 'success' : 'info'
})
const assetEventLabel = computed(() => latestEventSummary([], assetResult.value?.events || [], []))
const bindingSummary = computed(() => {
  const mainReferences = continuityProfiles.value.filter((profile) => profile.mainReferenceAssetId).length
  return `${continuityProfiles.value.length} profiles · ${continuityLineage.value.length} lineage rows · ${mainReferences} main refs`
})
const bindingStatusType = computed(() => {
  if (bindingResult.value?.error?.severity === 'blocking' || bindingResult.value?.error?.severity === 'error') {
    return 'error'
  }
  if (bindingResult.value?.error || continuityProfiles.value.some((profile) => profile.missingMainReference)) {
    return 'warning'
  }
  return bindingResult.value?.ok ? 'success' : 'info'
})
const bindingEventLabel = computed(() => latestEventSummary([], bindingResult.value?.events || [], []))
const continuityEventLabel = computed(() => latestEventSummary([], continuityResult.value?.events || [], []))
const continuitySummary = computed(() => {
  const locked = continuityRules.value.filter((rule) => rule.locked).length
  const affectedShots = continuityResult.value?.impact?.affectedShots.length || 0
  return `${continuityRules.value.length} rules · ${locked} locked · ${affectedShots} dirty shots`
})
const operationalEvents = computed(() => [
  ...(projectResult.value?.events || []),
  ...(assetResult.value?.events || []),
  ...(bindingResult.value?.events || []),
  ...(continuityResult.value?.events || []),
  ...(scriptResult.value?.events || []),
  ...(sceneResult.value?.events || []),
  ...(generationPackageResult.value?.events || []),
  ...(resultReviewResult.value?.events || []),
])
const auditHealthItems = computed(() => [
  ...(projectResult.value?.health?.items || []),
  ...(assetResult.value?.health?.items || []),
  ...(bindingResult.value?.health?.items || []),
  ...(continuityResult.value?.health?.items || []),
  ...(generationPackageResult.value?.health?.items || []),
  ...(mockRunResult.value?.health?.items || []),
  ...(resultReviewResult.value?.health?.items || []),
])
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
  assetResult.value?.error,
  bindingResult.value?.error,
  continuityResult.value?.error,
  scriptResult.value?.error,
  sceneResult.value?.error,
  shotContextResult.value?.error,
  generationPackageResult.value?.error,
  mockRunResult.value?.error,
  resultReviewResult.value?.error,
  probeError.value,
  ...graphErrors.value,
].filter((error): error is AppErrorDTO => Boolean(error)))
const healthStatus = computed(() => highestHealthStatus([
  graphResult.value?.health,
  projectResult.value?.health,
  assetResult.value?.health,
  bindingResult.value?.health,
  continuityResult.value?.health,
  generationPackageResult.value?.health,
  mockRunResult.value?.health,
  resultReviewResult.value?.health,
], visibleErrors.value))
const healthStatusTone = computed(() => healthTone(healthStatus.value))
const errorSummary = computed(() => firstErrorSummary([
  ...visibleErrors.value,
]))
const latestEventLabel = computed(() => latestEventSummary(
  graphResult.value?.events,
  [
    ...(projectResult.value?.events || []),
    ...(assetResult.value?.events || []),
    ...(bindingResult.value?.events || []),
    ...(continuityResult.value?.events || []),
    ...(scriptResult.value?.events || []),
    ...(sceneResult.value?.events || []),
    ...(shotContextResult.value?.events || []),
    ...(generationPackageResult.value?.events || []),
    ...(resultReviewResult.value?.events || []),
  ],
  runtimeEvents.value,
))
const runtimeProgress = computed(() => {
  const last = runtimeEvents.value[runtimeEvents.value.length - 1]
  return last?.progress ?? 0
})
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
const selectedShotID = computed(() => shotIDFromNode(selectedGraphNode.value))
const shotContextReport = computed(() => shotContextResult.value?.report)
const shotContextStatusType = computed(() => {
  if (!shotContextResult.value) {
    return 'info'
  }
  if (shotContextResult.value.error || !shotContextResult.value.report.canEnterContextReady) {
    return 'error'
  }
  return shotContextResult.value.shot?.status === 'context_ready' ? 'success' : 'warning'
})
const shotContextStatusLabel = computed(() => {
  if (!shotContextResult.value) {
    return 'idle'
  }
  if (shotContextResult.value.error || !shotContextResult.value.report.canEnterContextReady) {
    return 'blocked'
  }
  if (shotContextResult.value.report.status === 'context_ready') {
    return 'context_ready'
  }
  if (shotContextResult.value.report.status === 'context_dirty') {
    return 'context_dirty'
  }
  return 'eligible'
})
const shotContextIssueRows = computed(() => [
  ...(shotContextReport.value?.blocking || []),
  ...(shotContextReport.value?.warnings || []),
])
const shotContextReferenceRows = computed(() => shotContextReport.value?.references || [])
const activePackageShotID = computed(() => packageShotId.value.trim() || shotContextId.value.trim() || selectedShotID.value)
const activeMockShotID = computed(() => activePackageShotID.value || selectedShotID.value)
const activeMockPackageID = computed(() => generationPackageResult.value?.package?.packageId || '')
const mockRunStatusType = computed(() => {
  if (!mockRunResult.value) {
    return 'info'
  }
  if (mockRunResult.value.error?.severity === 'blocking' || mockRunResult.value.error?.severity === 'error') {
    return 'error'
  }
  if (mockRunResult.value.error || mockRunResult.value.run?.status === 'failed' || mockRunResult.value.run?.status === 'cancelled') {
    return 'warning'
  }
  return mockRunResult.value.run?.status === 'completed' ? 'success' : 'info'
})
const mockRunSummary = computed(() => {
  const run = mockRunResult.value?.run
  if (!run) {
    return mockRunResult.value?.error?.userMessage || 'No mock run started'
  }
  return `${run.runId} · ${run.providerMode} · attempt ${run.attempt}`
})
const mockRunOutputLabel = computed(() => {
  const output = mockRunResult.value?.run?.output
  if (!output) {
    return 'No placeholder output'
  }
  return `${output.relativePath} · ${output.digest.slice(0, 16)}`
})
const mockRunRecoveryActions = computed(() => mockRunResult.value?.error?.recoveryActions || [])
const resultRows = computed(() => resultReviewResult.value?.results || [])
const activeResult = computed(() => resultReviewResult.value?.result || resultRows.value[0])
const activeResultId = computed(() => activeResult.value?.id || '')
const activeResultTrace = computed(() => resultReviewResult.value?.trace)
const activeResultShotID = computed(() => resultShotId.value.trim() || selectedShotID.value || activeMockShotID.value)
const activeResultPackageID = computed(() => resultPackageId.value.trim() || activeMockPackageID.value)
const resultStatusType = computed(() => {
  if (!resultReviewResult.value) {
    return 'info'
  }
  if (resultReviewResult.value.error?.severity === 'blocking' || resultReviewResult.value.error?.severity === 'error') {
    return 'error'
  }
  const status = resultReviewResult.value.result?.status
  if (resultReviewResult.value.error || status === 'missing_file' || status === 'needs_revision' || status === 'rejected' || status === 'binding_pending') {
    return 'warning'
  }
  return status === 'approved' || resultReviewResult.value.ok ? 'success' : 'info'
})
const resultSummary = computed(() => {
  const result = activeResult.value
  if (!result) {
    return resultReviewResult.value?.error?.userMessage || 'No result imported'
  }
  return `${result.id} · ${result.status} · take ${result.takeNumber || 'pending'}`
})
const resultTraceLabel = computed(() => {
  const result = activeResult.value
  if (!result) {
    return 'No trace'
  }
  return [
    result.shotId || 'no-shot',
    result.packageId || 'no-package',
    result.source.runId || result.source.kind,
    result.assetId,
  ].filter(Boolean).join(' · ')
})
const resultRecoveryActions = computed(() => (
  resultReviewResult.value?.error?.recoveryActions ||
  resultReviewResult.value?.trace?.recoveryActions ||
  []
))
const packageStatusType = computed(() => {
  if (!generationPackageResult.value) {
    return 'info'
  }
  if (generationPackageResult.value.error?.severity === 'blocking' || generationPackageResult.value.error?.severity === 'error') {
    return 'error'
  }
  if (generationPackageResult.value.error) {
    return 'warning'
  }
  return generationPackageResult.value.package?.generationPackageStatus === 'ready' ? 'success' : 'warning'
})
const packageSummary = computed(() => {
  const pkg = generationPackageResult.value?.package
  if (!pkg) {
    return generationPackageResult.value?.error?.userMessage || 'No package exported'
  }
  return `${pkg.packageId} · v${pkg.packageVersion} · ${pkg.references.length} refs`
})
const packageRecoveryActions = computed(() => generationPackageResult.value?.error?.recoveryActions || [])
const packageReferenceRows = computed(() => generationPackageResult.value?.package?.references || [])

onMounted(() => {
  void loadProjectGraph()
  void loadAssetRows()
})

async function loadAssetRows() {
  assetLoading.value = true
  assetResult.value = await listAssets()
  assetRows.value = assetResult.value.assets
  synchronizeBindingDefaults()
  if (!assetResult.value.ok) {
    activeInspectorTab.value = 'tasks'
  }
  assetLoading.value = false
  await loadCurrentAssetBindings(false)
}

async function importCurrentAsset() {
  assetImporting.value = true
  const previousRows = assetRows.value
  assetResult.value = await importAsset({
    sourcePath: assetImportPath.value,
    role: assetImportRole.value,
    duplicatePolicy: assetDuplicatePolicy.value,
    managedReference: assetManagedReference.value,
  })
  assetRows.value = assetResult.value.ok || assetResult.value.assets.length > 0
    ? assetResult.value.assets
    : previousRows
  synchronizeBindingDefaults()
  activeInspectorTab.value = 'tasks'
  if (assetResult.value.ok) {
    await loadCurrentAssetBindings(false)
  }
  assetImporting.value = false
}

async function loadCurrentAssetBindings(revealTasks = true) {
  bindingLoading.value = true
  bindingResult.value = await listAssetBindings()
  ingestContinuityResult(bindingResult.value)
  continuityResult.value = await listContinuity()
  synchronizeContinuityDefaults()
  if (revealTasks) {
    activeInspectorTab.value = 'tasks'
  }
  bindingLoading.value = false
}

async function bindCurrentAsset() {
  bindingSubmitting.value = true
  bindingResult.value = await bindAsset({
    assetId: bindingAssetId.value,
    targetType: bindingTargetType.value,
    targetId: bindingTargetId.value,
    purpose: bindingPurpose.value,
    duplicatePolicy: bindingDuplicatePolicy.value,
  })
  ingestContinuityResult(bindingResult.value)
  if (bindingResult.value.ok) {
    await loadProjectGraph()
  } else {
    activeInspectorTab.value = 'tasks'
  }
  bindingSubmitting.value = false
}

async function setCurrentMainReference() {
  mainReferenceSaving.value = true
  bindingResult.value = await setMainReference({
    assetId: mainReferenceAssetId.value || bindingAssetId.value,
    targetType: bindingTargetType.value,
    targetId: bindingTargetId.value,
    clear: mainReferenceClear.value,
  })
  ingestContinuityResult(bindingResult.value)
  if (bindingResult.value.ok) {
    mainReferenceClear.value = false
    await loadProjectGraph()
  } else {
    activeInspectorTab.value = 'tasks'
  }
  mainReferenceSaving.value = false
}

async function saveCurrentContinuityRule() {
  continuityAction.value = 'save-rule'
  continuityResult.value = await saveContinuityRule({
    id: continuityRuleId.value,
    targetType: bindingTargetType.value,
    targetId: bindingTargetId.value,
    rule: continuityRuleText.value,
    severity: continuityRuleSeverity.value,
    locked: continuityRuleLocked.value,
  })
  synchronizeContinuityDefaults()
  if (continuityResult.value.ok) {
    await loadCurrentAssetBindings(false)
    await loadProjectGraph()
  } else {
    activeInspectorTab.value = 'tasks'
  }
  continuityAction.value = ''
}

async function unlockCurrentContinuityRule() {
  continuityAction.value = 'unlock-rule'
  continuityResult.value = await unlockContinuityRule({
    ruleId: continuityRuleId.value,
    reason: continuityUnlockReason.value,
  })
  synchronizeContinuityDefaults()
  if (continuityResult.value.ok) {
    await loadCurrentAssetBindings(false)
    await loadProjectGraph()
  } else {
    activeInspectorTab.value = 'tasks'
  }
  continuityAction.value = ''
}

async function unlockCurrentAssetBinding() {
  const binding = selectedBindingForUnlock.value
  continuityAction.value = 'unlock-binding'
  continuityResult.value = await unlockAssetBinding({
    bindingId: binding?.id,
    assetId: bindingAssetId.value,
    targetType: bindingTargetType.value,
    targetId: bindingTargetId.value,
    purpose: bindingPurpose.value,
    reason: bindingUnlockReason.value,
  })
  if (continuityResult.value.ok) {
    await loadCurrentAssetBindings(false)
    await loadProjectGraph()
  } else {
    activeInspectorTab.value = 'tasks'
  }
  continuityAction.value = ''
}

async function unlockCurrentMainReferenceBinding() {
  const binding = selectedMainReferenceBinding.value
  if (!binding) {
    return
  }
  continuityAction.value = 'unlock-main-reference'
  continuityResult.value = await unlockAssetBinding({
    bindingId: binding.id,
    reason: bindingUnlockReason.value,
  })
  if (continuityResult.value.ok) {
    await loadCurrentAssetBindings(false)
    await loadProjectGraph()
  } else {
    activeInspectorTab.value = 'tasks'
  }
  continuityAction.value = ''
}

function ingestContinuityResult(result: ContinuityLibraryResultDTO) {
  if (result.assets.length) {
    assetRows.value = result.assets
  }
  synchronizeBindingDefaults(result)
}

function synchronizeContinuityDefaults() {
  const rule = continuityRules.value.find((item) => item.id === continuityRuleId.value) || continuityRules.value[0]
  if (!rule) {
    return
  }
  continuityRuleId.value = rule.id
  continuityRuleText.value = rule.rule
  continuityRuleSeverity.value = rule.severity === 'warning' || rule.severity === 'suggestion' ? rule.severity : 'blocking'
  continuityRuleLocked.value = rule.locked
  bindingTargetType.value = normalizeBindingTargetType(rule.targetType)
  bindingTargetId.value = rule.targetId
}

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
  const previousSelection = selectedGraphNode.value?.id
  graphResult.value = await runProjectGraphView(expectedGraphVersion)
  const canvas = graphResult.value.canvas
  if (canvas) {
    activeGraphCanvas.value = canvas
    themeMode.value = canvas.theme === 'warm_light' ? 'warm' : 'dark'
    canvasGrid.value = canvas.grid
    graphViewportLabel.value = `${Math.round(canvas.viewport.zoom * 100)}%`
    if (previousSelection) {
      selectedGraphNode.value = canvas.nodes.find((node) => node.id === previousSelection)
      selectedGraphNodes.value = selectedGraphNode.value ? [selectedGraphNode.value] : []
    }
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

function selectGraphNode(node: GraphNodeDTO | undefined) {
  selectedGraphNode.value = node
  if (node?.kind === 'shot') {
    shotContextId.value = shotIDFromNode(node) || shotContextId.value
    packageShotId.value = shotContextId.value
  }
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

async function checkShotContext() {
  const shotId = shotContextId.value.trim() || selectedShotID.value
  if (!shotId) {
    return
  }
  shotContextAction.value = 'check'
  shotContextResult.value = await validateShotContext({ shotId })
  activeInspectorTab.value = 'tasks'
  shotContextAction.value = ''
}

async function promoteCurrentShotContext() {
  const shotId = shotContextId.value.trim() || selectedShotID.value
  if (!shotId) {
    return
  }
  shotContextAction.value = 'ready'
  shotContextResult.value = await promoteShotContext({ shotId })
  if (shotContextResult.value.ok) {
    await loadProjectGraph()
  }
  activeInspectorTab.value = 'tasks'
  shotContextAction.value = ''
}

async function markCurrentShotContextDirty() {
  const shotId = shotContextId.value.trim() || selectedShotID.value
  if (!shotId) {
    return
  }
  shotContextAction.value = 'dirty'
  shotContextResult.value = await markShotContextDirty({
    shotId,
    reason: shotDirtyReason.value,
  })
  if (shotContextResult.value.ok) {
    await loadProjectGraph()
  }
  activeInspectorTab.value = 'tasks'
  shotContextAction.value = ''
}

async function exportCurrentGenerationPackage() {
  const shotId = activePackageShotID.value
  if (!shotId) {
    return
  }
  packageShotId.value = shotId
  packageAction.value = 'export'
  generationPackageResult.value = await exportGenerationPackage({ shotId })
  if (generationPackageResult.value.ok) {
    await loadProjectGraph()
  }
  activeInspectorTab.value = 'tasks'
  packageAction.value = ''
}

async function startCurrentMockRun() {
  const shotId = activeMockShotID.value
  const packageId = activeMockPackageID.value
  if (!shotId && !packageId && selectedGraphNodeIDs.value.length === 0) {
    return
  }
  mockRunAction.value = 'start'
  mockRunResult.value = await startMockRun({
    shotId,
    packageId,
    selectionIds: selectedGraphNodeIDs.value,
  })
  runtimeEvents.value = mockRunResult.value.events
  if (mockRunResult.value.ok) {
    resultRunId.value = mockRunResult.value.run?.runId || resultRunId.value
    resultShotId.value = mockRunResult.value.run?.shotId || resultShotId.value
    resultPackageId.value = mockRunResult.value.run?.packageId || resultPackageId.value
    resultSourcePath.value = ''
    await loadProjectGraph()
  }
  activeInspectorTab.value = 'runs'
  mockRunAction.value = ''
}

async function cancelCurrentMockRun() {
  const run = mockRunResult.value?.run
  const shotId = activeMockShotID.value || run?.shotId
  if (!run?.runId && !shotId) {
    return
  }
  mockRunAction.value = 'cancel'
  mockRunResult.value = await cancelMockRun({
    runId: run?.runId,
    shotId,
    packageId: run?.packageId || activeMockPackageID.value,
    selectionIds: selectedGraphNodeIDs.value.length ? selectedGraphNodeIDs.value : run?.selectionIds,
    cancelReason: 'user_cancelled',
  })
  runtimeEvents.value = mockRunResult.value.events
  if (mockRunResult.value.ok) {
    await loadProjectGraph()
  }
  activeInspectorTab.value = 'runs'
  mockRunAction.value = ''
}

async function retryCurrentMockRun() {
  const run = mockRunResult.value?.run
  if (!run?.runId) {
    return
  }
  mockRunAction.value = 'retry'
  mockRunResult.value = await retryMockRun({
    runId: run.runId,
    shotId: run.shotId || activeMockShotID.value,
    packageId: run.packageId || activeMockPackageID.value,
    selectionIds: run.selectionIds,
  })
  runtimeEvents.value = mockRunResult.value.events
  if (mockRunResult.value.ok) {
    await loadProjectGraph()
  }
  activeInspectorTab.value = 'runs'
  mockRunAction.value = ''
}

async function importCurrentResult() {
  resultAction.value = 'import'
  resultReviewResult.value = await importResult({
    sourcePath: resultSourcePath.value,
    shotId: activeResultShotID.value,
    packageId: activeResultPackageID.value,
    runId: resultRunId.value,
    duplicatePolicy: resultDuplicatePolicy.value,
  })
  if (resultReviewResult.value.ok) {
    resultReviewNotes.value = 'Looks usable for alpha review.'
    await loadProjectGraph()
  }
  activeInspectorTab.value = 'tasks'
  resultAction.value = ''
}

async function listCurrentResults() {
  resultAction.value = 'list'
  resultReviewResult.value = await listResults({
    shotId: resultShotId.value.trim() || undefined,
    packageId: resultPackageId.value.trim() || undefined,
  })
  activeInspectorTab.value = 'tasks'
  resultAction.value = ''
}

async function traceCurrentResult() {
  const resultId = activeResultId.value
  if (!resultId) {
    return
  }
  resultAction.value = 'trace'
  resultReviewResult.value = await traceResult({ resultId })
  if (resultReviewResult.value.result) {
    resultShotId.value = resultReviewResult.value.result.shotId || resultShotId.value
    resultPackageId.value = resultReviewResult.value.result.packageId || resultPackageId.value
  }
  activeInspectorTab.value = 'tasks'
  resultAction.value = ''
}

async function updateCurrentResultReview() {
  const resultId = activeResultId.value
  if (!resultId) {
    return
  }
  resultAction.value = 'review'
  resultReviewResult.value = await updateResultReview({
    resultId,
    reviewStatus: resultReviewStatus.value,
    reviewNotes: resultReviewNotes.value,
    reason: resultReviewReason.value,
  })
  if (resultReviewResult.value.ok) {
    await loadProjectGraph()
  }
  activeInspectorTab.value = 'tasks'
  resultAction.value = ''
}

async function rebindCurrentResult() {
  const resultId = activeResultId.value
  if (!resultId) {
    return
  }
  resultAction.value = 'rebind'
  resultReviewResult.value = await rebindResult({
    resultId,
    shotId: resultUnbindShot.value ? undefined : resultRebindShotId.value,
    packageId: resultUnbindPackage.value ? undefined : resultRebindPackageId.value,
    unbindShot: resultUnbindShot.value,
    unbindPackage: resultUnbindPackage.value,
    reason: resultRebindReason.value,
  })
  if (resultReviewResult.value.ok) {
    const updatedResult = resultReviewResult.value.result
    if (updatedResult) {
      resultShotId.value = updatedResult.shotId || ''
      resultPackageId.value = updatedResult.packageId || ''
      resultRebindShotId.value = updatedResult.shotId || ''
      resultRebindPackageId.value = updatedResult.packageId || ''
    }
    resultUnbindShot.value = false
    resultUnbindPackage.value = false
    await loadProjectGraph()
  }
  activeInspectorTab.value = 'tasks'
  resultAction.value = ''
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

function assetTone(asset: AssetDTO): string {
  if (asset.missing) {
    return 'error'
  }
  if (asset.thumbnailStatus === 'thumbnail_failed' || asset.source.kind === 'managed_reference') {
    return 'warning'
  }
  return 'success'
}

function assetStatusText(asset: AssetDTO): string {
  const state = asset.missing ? 'missing' : asset.thumbnailStatus
  return `${asset.type} · ${asset.role} · ${state}`
}

function assetDetailText(asset: AssetDTO): string {
  return [
    `digest ${asset.digestSummary || '-'}`,
    asset.source.kind,
    `${asset.bindingCount} bindings`,
  ].join(' · ')
}

function synchronizeBindingDefaults(result?: ContinuityLibraryResultDTO) {
  const assets = result?.assets.length ? result.assets : assetRows.value
  const profiles = result?.profiles || continuityProfiles.value
  const selectedAssetExists = assets.some((asset) => asset.id === bindingAssetId.value)
  const selectedMainAssetExists = assets.some((asset) => asset.id === mainReferenceAssetId.value)
  const firstAssetId = assets[0]?.id

  if (firstAssetId && (!bindingAssetId.value || !selectedAssetExists)) {
    bindingAssetId.value = firstAssetId
  }
  if (firstAssetId && (!mainReferenceAssetId.value || !selectedMainAssetExists)) {
    mainReferenceAssetId.value = firstAssetId
  }

  const selectedProfileExists = profiles.some((profile) => (
    profile.type === bindingTargetType.value && profile.id === bindingTargetId.value
  ))
  const firstProfile = profiles[0]
  if (firstProfile && (!bindingTargetId.value || !selectedProfileExists)) {
    bindingTargetType.value = normalizeBindingTargetType(firstProfile.type)
    bindingTargetId.value = firstProfile.id
  }
}

function normalizeBindingTargetType(value: string): AssetBindingTargetType {
  return value === 'scene' || value === 'prop' ? value : 'character'
}

function profileTone(profile: ProfileDTO): string {
  if (profile.missingMainReference) {
    return 'warning'
  }
  if (profile.mainReferenceAssetId) {
    return 'success'
  }
  return 'default'
}

function profileDetailText(profile: ProfileDTO): string {
  return [
    profile.relativePath || profile.type,
    `${profile.referenceAssetIds.length} refs`,
    `${profile.bindingCount} bindings`,
  ].join(' · ')
}

function profileMainReferenceText(profile: ProfileDTO): string {
  return profile.mainReferenceAssetId
    ? `${profile.mainReferenceAssetId} · ${profile.mainReferencePath || 'asset index'}`
    : 'No main reference'
}

function lineageDetailText(assetId: string): string {
  const lineage = continuityLineage.value.find((item) => item.assetId === assetId)
  if (!lineage) {
    return 'No lineage row loaded'
  }
  const targets = lineage.targetSummaries.map((target) => `${target.targetType}:${target.targetId}`)
  return targets.length ? targets.join(', ') : 'No targets'
}

function shotIDFromNode(node: GraphNodeDTO | undefined): string {
  if (node?.kind === 'shot' && node.data?.shotId) {
    return node.data.shotId
  }
  const refId = node?.kind === 'shot' ? node.refId || '' : ''
  const filename = refId.split('/').pop() || ''
  return filename.replace(/\.json$/i, '')
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
          <Badge status="processing" :text="`Queue ${runtimeEvents.length + (projectResult?.events.length || 0) + (assetResult?.events.length || 0) + (bindingResult?.events.length || 0) + (scriptResult?.events.length || 0) + (sceneResult?.events.length || 0) + (generationPackageResult?.events.length || 0) + (resultReviewResult?.events.length || 0)}`" />
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
            <Button
              size="small"
              type="primary"
              aria-label="Export handoff package"
              :loading="packageAction === 'export'"
              :disabled="!activePackageShotID"
              @click="exportCurrentGenerationPackage"
            >
              <template #icon>
                <CloudUploadOutlined />
              </template>
              Export
            </Button>
          </Tooltip>
          <Tooltip title="Run deterministic mock task">
            <Button
              size="small"
              aria-label="Run deterministic mock task"
              :loading="mockRunAction === 'start'"
              :disabled="!activeMockShotID && !activeMockPackageID && selectedGraphNodeIDs.length === 0"
              @click="startCurrentMockRun"
            >
              <template #icon>
                <ThunderboltOutlined />
              </template>
              Mock
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
              <div class="asset-import-panel">
                <label class="script-field">
                  <span>Source path</span>
                  <input
                    v-model="assetImportPath"
                    class="script-input"
                    type="text"
                    autocomplete="off"
                  >
                </label>
                <div class="asset-import-grid">
                  <label class="script-field">
                    <span>Role</span>
                    <input
                      v-model="assetImportRole"
                      class="script-input"
                      type="text"
                      autocomplete="off"
                    >
                  </label>
                  <label class="script-field">
                    <span>Duplicate</span>
                    <select v-model="assetDuplicatePolicy" class="script-input">
                      <option value="cancel">Cancel</option>
                      <option value="reuse">Reuse</option>
                      <option value="copy">Copy</option>
                    </select>
                  </label>
                </div>
                <label class="check-row">
                  <input v-model="assetManagedReference" type="checkbox">
                  <span>Managed reference</span>
                </label>
                <div class="rail-toolbar">
                  <Button size="small" type="primary" :loading="assetImporting" :disabled="!assetImportPath.trim()" @click="importCurrentAsset">
                    <template #icon>
                      <FolderOpenOutlined />
                    </template>
                    Import
                  </Button>
                  <Button size="small" :loading="assetLoading" @click="loadAssetRows">
                    <template #icon>
                      <ReloadOutlined />
                    </template>
                    List
                  </Button>
                </div>
              </div>
              <Alert
                v-if="assetResult"
                class="service-alert"
                :type="assetStatusType"
                show-icon
                :message="assetResult.error?.userMessage || assetSummary"
                :description="assetResult.duplicate ? `duplicate ${assetResult.duplicate.existingAssetId}` : assetEventLabel"
              />
              <List
                v-if="assetRecoveryActions.length"
                class="recovery-list"
                size="small"
                :data-source="assetRecoveryActions"
              >
                <template #renderItem="{ item }">
                  <ListItem>
                    <span>{{ item }}</span>
                  </ListItem>
                </template>
              </List>
              <div class="binding-panel">
                <div class="script-editor-bar">
                  <Tag :color="bindingStatusType">Continuity</Tag>
                  <Tag>{{ bindingSummary }}</Tag>
                </div>
                <div class="asset-import-grid">
                  <label class="script-field">
                    <span>Asset id</span>
                    <input
                      v-model="bindingAssetId"
                      class="script-input"
                      type="text"
                      autocomplete="off"
                    >
                  </label>
                  <label class="script-field">
                    <span>Target</span>
                    <select v-model="bindingTargetType" class="script-input">
                      <option value="character">Character</option>
                      <option value="scene">Scene</option>
                      <option value="prop">Prop</option>
                    </select>
                  </label>
                </div>
                <label class="script-field">
                  <span>Target id</span>
                  <input
                    v-model="bindingTargetId"
                    class="script-input"
                    type="text"
                    autocomplete="off"
                  >
                </label>
                <div class="asset-import-grid">
                  <label class="script-field">
                    <span>Purpose</span>
                    <select
                      v-model="bindingPurpose"
                      class="script-input"
                    >
                      <option value="reference">Reference</option>
                      <option value="identity">Identity</option>
                      <option value="style">Style</option>
                      <option value="detail">Detail</option>
                    </select>
                  </label>
                  <label class="script-field">
                    <span>Duplicate</span>
                    <select v-model="bindingDuplicatePolicy" class="script-input">
                      <option value="cancel">Cancel</option>
                      <option value="reuse">Reuse</option>
                    </select>
                  </label>
                </div>
                <label class="script-field">
                  <span>Main ref asset</span>
                  <input
                    v-model="mainReferenceAssetId"
                    class="script-input"
                    type="text"
                    autocomplete="off"
                  >
                </label>
                <label class="check-row">
                  <input v-model="mainReferenceClear" type="checkbox">
                  <span>Clear main reference</span>
                </label>
                <div class="rail-toolbar">
                  <Button size="small" :loading="bindingLoading" @click="loadCurrentAssetBindings()">
                    <template #icon>
                      <ReloadOutlined />
                    </template>
                    Bindings
                  </Button>
                  <Button size="small" type="primary" :loading="bindingSubmitting" :disabled="!bindingAssetId.trim() || !bindingTargetId.trim()" @click="bindCurrentAsset">
                    <template #icon>
                      <BranchesOutlined />
                    </template>
                    Bind
                  </Button>
                  <Button size="small" :loading="mainReferenceSaving" :disabled="!bindingTargetId.trim() || (!mainReferenceClear && !mainReferenceAssetId.trim())" @click="setCurrentMainReference">
                    <template #icon>
                      <ShareAltOutlined />
                    </template>
                    Main ref
                  </Button>
                </div>
              </div>
              <div class="binding-panel">
                <div class="script-editor-bar">
                  <Tag color="purple">Rules</Tag>
                  <Tag>{{ continuitySummary }}</Tag>
                </div>
                <label class="script-field">
                  <span>Rule id</span>
                  <input
                    v-model="continuityRuleId"
                    class="script-input"
                    type="text"
                    autocomplete="off"
                  >
                </label>
                <label class="script-field">
                  <span>Rule</span>
                  <textarea
                    v-model="continuityRuleText"
                    class="script-textarea"
                    rows="3"
                  />
                </label>
                <div class="asset-import-grid">
                  <label class="script-field">
                    <span>Severity</span>
                    <select v-model="continuityRuleSeverity" class="script-input">
                      <option value="blocking">Blocking</option>
                      <option value="warning">Warning</option>
                      <option value="suggestion">Suggestion</option>
                    </select>
                  </label>
                  <label class="check-row check-row--stacked">
                    <input v-model="continuityRuleLocked" type="checkbox">
                    <span>Locked</span>
                  </label>
                </div>
                <label class="script-field">
                  <span>Unlock reason</span>
                  <input
                    v-model="continuityUnlockReason"
                    class="script-input"
                    type="text"
                    autocomplete="off"
                  >
                </label>
                <label class="script-field">
                  <span>Binding unlock reason</span>
                  <input
                    v-model="bindingUnlockReason"
                    class="script-input"
                    type="text"
                    autocomplete="off"
                  >
                </label>
                <div class="rail-toolbar">
                  <Button size="small" :loading="continuityAction === 'save-rule'" :disabled="!bindingTargetId.trim() || !continuityRuleText.trim()" @click="saveCurrentContinuityRule">
                    <template #icon>
                      <SaveOutlined />
                    </template>
                    Save rule
                  </Button>
                  <Button size="small" :loading="continuityAction === 'unlock-rule'" :disabled="!continuityRuleId.trim() || !continuityUnlockReason.trim()" @click="unlockCurrentContinuityRule">
                    Unlock rule
                  </Button>
                  <Button size="small" :loading="continuityAction === 'unlock-binding'" :disabled="!bindingAssetId.trim() || !bindingTargetId.trim() || !bindingUnlockReason.trim()" @click="unlockCurrentAssetBinding">
                    Unlock binding
                  </Button>
                  <Button size="small" :loading="continuityAction === 'unlock-main-reference'" :disabled="!selectedMainReferenceBinding || !bindingUnlockReason.trim()" @click="unlockCurrentMainReferenceBinding">
                    Unlock main ref
                  </Button>
                </div>
              </div>
              <Alert
                v-if="continuityResult"
                class="service-alert"
                :type="continuityResult.error ? 'warning' : 'success'"
                show-icon
                :message="continuityResult.error?.userMessage || continuitySummary"
                :description="continuityResult.error ? `${continuityResult.error.code} · ${continuityResult.error.correlationId}` : continuityEventLabel"
              />
              <List
                v-if="continuityRecoveryActions.length"
                class="recovery-list"
                size="small"
                :data-source="continuityRecoveryActions"
              >
                <template #renderItem="{ item }">
                  <ListItem>
                    <span>{{ item }}</span>
                  </ListItem>
                </template>
              </List>
              <Alert
                v-if="bindingResult"
                class="service-alert"
                :type="bindingStatusType"
                show-icon
                :message="bindingResult.error?.userMessage || bindingSummary"
                :description="bindingResult.duplicate ? `duplicate ${bindingResult.duplicate.id || bindingResult.duplicate.assetId}` : bindingEventLabel"
              />
              <List
                v-if="bindingRecoveryActions.length"
                class="recovery-list"
                size="small"
                :data-source="bindingRecoveryActions"
              >
                <template #renderItem="{ item }">
                  <ListItem>
                    <span>{{ item }}</span>
                  </ListItem>
                </template>
              </List>
              <List v-if="continuityProfiles.length" class="profile-list" item-layout="horizontal" size="small" :data-source="continuityProfiles">
                <template #renderItem="{ item }">
                  <ListItem>
                    <ListItemMeta>
                      <template #title>
                        <strong class="asset-title" :title="item.relativePath">{{ item.name || item.id }}</strong>
                      </template>
                      <template #description>
                        <span class="asset-status">{{ item.type }} · {{ profileDetailText(item) }}</span>
                        <span class="asset-status">{{ profileMainReferenceText(item) }}</span>
                      </template>
                    </ListItemMeta>
                    <Tag :color="profileTone(item)">{{ item.mainReferenceAssetId ? 'main' : 'empty' }}</Tag>
                  </ListItem>
                </template>
              </List>
              <List class="asset-list" item-layout="horizontal" size="small" :data-source="assetRows">
                <template #renderItem="{ item }">
                  <ListItem>
                    <ListItemMeta>
                      <template #avatar>
                        <div class="asset-thumb" :class="`asset-thumb--${assetTone(item)}`">{{ item.type }}</div>
                      </template>
                      <template #title>
                        <strong class="asset-title" :title="item.relativePath">{{ item.originalName || item.id }}</strong>
                      </template>
                      <template #description>
                        <span class="asset-status">{{ assetStatusText(item) }}</span>
                        <span class="asset-status">{{ assetDetailText(item) }}</span>
                        <span class="asset-status">{{ lineageDetailText(item.id) }}</span>
                      </template>
                    </ListItemMeta>
                    <Tag :color="assetTone(item)">{{ item.missing ? 'missing' : item.thumbnailStatus }}</Tag>
                  </ListItem>
                </template>
              </List>
              <div v-if="!assetRows.length && !assetLoading" class="placeholder-list">
                <strong>No indexed assets</strong>
                <span>{{ assetSummary }}</span>
              </div>
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
              @select="selectGraphNode"
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

                <section v-if="selectedGraphNode?.kind === 'shot' || shotContextResult || shotContextId" class="service-panel" aria-label="Shot context">
                  <div class="shot-context-header">
                    <strong>Shot context</strong>
                    <Tag v-if="shotContextReport" :color="shotContextStatusType">
                      {{ shotContextStatusLabel }}
                    </Tag>
                  </div>
                  <label class="script-field">
                    <span>Shot id</span>
                    <input v-model="shotContextId" class="script-input" type="text" autocomplete="off">
                  </label>
                  <label class="script-field">
                    <span>Dirty reason</span>
                    <input v-model="shotDirtyReason" class="script-input" type="text" autocomplete="off">
                  </label>
                  <div class="service-actions">
                    <Button size="small" :loading="shotContextAction === 'check'" @click="checkShotContext">
                      Check context
                    </Button>
                    <Button size="small" type="primary" :loading="shotContextAction === 'ready'" @click="promoteCurrentShotContext">
                      Context ready
                    </Button>
                    <Button size="small" :loading="shotContextAction === 'dirty'" @click="markCurrentShotContextDirty">
                      Mark dirty
                    </Button>
                  </div>
                  <Alert
                    v-if="shotContextResult"
                    class="service-alert"
                    :type="shotContextStatusType"
                    show-icon
                    :message="shotContextResult.error?.userMessage || `${shotContextReport?.shotId || shotContextId} · ${shotContextReport?.status || 'checked'}`"
                    :description="shotContextResult.error ? `${shotContextResult.error.code} · ${shotContextResult.error.correlationId}` : `${shotContextReport?.missingFields.length || 0} missing · ${shotContextReport?.blocking.length || 0} blocking`"
                  />
                  <div v-if="shotContextReport?.missingFields.length" class="capability-row">
                    <Tag v-for="field in shotContextReport.missingFields" :key="field" color="error">
                      {{ field }}
                    </Tag>
                  </div>
                  <List
                    v-if="shotContextIssueRows.length"
                    class="recovery-list"
                    size="small"
                    :data-source="shotContextIssueRows"
                  >
                    <template #renderItem="{ item }">
                      <ListItem>
                        <span>{{ item.field || item.code }} · {{ item.userMessage }}</span>
                      </ListItem>
                    </template>
                  </List>
                  <List
                    v-if="shotContextReferenceRows.length"
                    class="recovery-list"
                    size="small"
                    :data-source="shotContextReferenceRows"
                  >
                    <template #renderItem="{ item }">
                      <ListItem>
                        <span>{{ item.kind }} · {{ item.referenceId }}</span>
                        <Tag :color="item.status === 'resolved' ? 'success' : 'error'">{{ item.status }}</Tag>
                      </ListItem>
                    </template>
                  </List>
                </section>

                <section class="service-panel" aria-label="Generation package">
                  <div class="shot-context-header">
                    <strong>Generation package</strong>
                    <Tag v-if="generationPackageResult" :color="packageStatusType">
                      {{ generationPackageResult.package?.generationPackageStatus || generationPackageResult.error?.code }}
                    </Tag>
                  </div>
                  <label class="script-field">
                    <span>Shot id</span>
                    <input v-model="packageShotId" class="script-input" type="text" autocomplete="off">
                  </label>
                  <div class="service-actions">
                    <Button
                      size="small"
                      type="primary"
                      :loading="packageAction === 'export'"
                      :disabled="!activePackageShotID"
                      @click="exportCurrentGenerationPackage"
                    >
                      <template #icon>
                        <CloudUploadOutlined />
                      </template>
                      Export package
                    </Button>
                  </div>
                  <Alert
                    v-if="generationPackageResult"
                    class="service-alert"
                    :type="packageStatusType"
                    show-icon
                    :message="packageSummary"
                    :description="generationPackageResult.package ? generationPackageResult.package.relativePath : `${generationPackageResult.error?.code} · ${generationPackageResult.error?.correlationId}`"
                  />
                  <dl v-if="generationPackageResult?.package" class="property-grid">
                    <div>
                      <dt>Manifest</dt>
                      <dd :title="generationPackageResult.package.manifestPath">
                        {{ generationPackageResult.package.manifestPath }}
                      </dd>
                    </div>
                    <div>
                      <dt>Prompt</dt>
                      <dd :title="generationPackageResult.package.promptPath">
                        {{ generationPackageResult.package.promptPath }}
                      </dd>
                    </div>
                    <div>
                      <dt>Checklist</dt>
                      <dd :title="generationPackageResult.package.uploadChecklistPath">
                        {{ generationPackageResult.package.uploadChecklistPath }}
                      </dd>
                    </div>
                    <div>
                      <dt>Digest</dt>
                      <dd :title="generationPackageResult.package.contextDigest">
                        {{ generationPackageResult.package.contextDigest.slice(0, 16) }}
                      </dd>
                    </div>
                  </dl>
                  <List
                    v-if="packageReferenceRows.length"
                    class="recovery-list"
                    size="small"
                    :data-source="packageReferenceRows"
                  >
                    <template #renderItem="{ item }">
                      <ListItem>
                        <span>{{ item.assetId }} · {{ item.packagePath }}</span>
                      </ListItem>
                    </template>
                  </List>
                  <List
                    v-if="packageRecoveryActions.length"
                    class="recovery-list"
                    size="small"
                    :data-source="packageRecoveryActions"
                  >
                    <template #renderItem="{ item }">
                      <ListItem>
                        <span>{{ item }}</span>
                      </ListItem>
                    </template>
                  </List>
                </section>

                <section class="service-panel" aria-label="Mock run">
                  <div class="shot-context-header">
                    <strong>Mock run</strong>
                    <Tag v-if="mockRunResult" :color="mockRunStatusType">
                      {{ mockRunResult.run?.status || mockRunResult.error?.code }}
                    </Tag>
                  </div>
                  <div class="service-actions">
                    <Button
                      size="small"
                      type="primary"
                      :loading="mockRunAction === 'start'"
                      :disabled="!activeMockShotID && !activeMockPackageID && selectedGraphNodeIDs.length === 0"
                      @click="startCurrentMockRun"
                    >
                      <template #icon>
                        <ThunderboltOutlined />
                      </template>
                      Start mock
                    </Button>
                    <Button
                      size="small"
                      :loading="mockRunAction === 'cancel'"
                      :disabled="!mockRunResult?.run && !activeMockShotID"
                      @click="cancelCurrentMockRun"
                    >
                      Cancel
                    </Button>
                    <Button
                      size="small"
                      :loading="mockRunAction === 'retry'"
                      :disabled="!mockRunResult?.run?.runId"
                      @click="retryCurrentMockRun"
                    >
                      <template #icon>
                        <ReloadOutlined />
                      </template>
                      Retry
                    </Button>
                  </div>
                  <Alert
                    v-if="mockRunResult"
                    class="service-alert"
                    :type="mockRunStatusType"
                    show-icon
                    :message="mockRunSummary"
                    :description="mockRunResult.run ? mockRunOutputLabel : `${mockRunResult.error?.code} · ${mockRunResult.error?.correlationId}`"
                  />
                  <dl v-if="mockRunResult?.run" class="property-grid">
                    <div>
                      <dt>Mode</dt>
                      <dd>{{ mockRunResult.run.providerMode }}</dd>
                    </div>
                    <div>
                      <dt>Target</dt>
                      <dd>{{ mockRunResult.run.packageId || mockRunResult.run.shotId || mockRunResult.run.selectionIds.join(', ') }}</dd>
                    </div>
                    <div>
                      <dt>Run</dt>
                      <dd :title="mockRunResult.run.runPath">{{ mockRunResult.run.runPath }}</dd>
                    </div>
                    <div>
                      <dt>Events</dt>
                      <dd :title="mockRunResult.run.eventsPath">{{ mockRunResult.run.eventsPath }}</dd>
                    </div>
                    <div>
                      <dt>Digest</dt>
                      <dd :title="mockRunResult.run.contextDigest">{{ mockRunResult.run.contextDigest.slice(0, 16) }}</dd>
                    </div>
                    <div v-if="mockRunResult.run.output">
                      <dt>Output</dt>
                      <dd :title="mockRunResult.run.output.relativePath">{{ mockRunResult.run.output.relativePath }}</dd>
                    </div>
                  </dl>
                  <List
                    v-if="mockRunRecoveryActions.length"
                    class="recovery-list"
                    size="small"
                    :data-source="mockRunRecoveryActions"
                  >
                    <template #renderItem="{ item }">
                      <ListItem>
                        <span>{{ item }}</span>
                      </ListItem>
                    </template>
                  </List>
                </section>

                <section class="service-panel" aria-label="Result review">
                  <div class="shot-context-header">
                    <strong>Result review</strong>
                    <Tag v-if="resultReviewResult" :color="resultStatusType">
                      {{ resultReviewResult.result?.status || resultReviewResult.error?.code }}
                    </Tag>
                  </div>
                  <label class="script-field">
                    <span>Source path</span>
                    <input v-model="resultSourcePath" class="script-input" type="text" autocomplete="off">
                  </label>
                  <div class="script-field-grid">
                    <label class="script-field">
                      <span>Shot id</span>
                      <input v-model="resultShotId" class="script-input" type="text" autocomplete="off">
                    </label>
                    <label class="script-field">
                      <span>Package id</span>
                      <input v-model="resultPackageId" class="script-input" type="text" autocomplete="off">
                    </label>
                  </div>
                  <div class="script-field-grid">
                    <label class="script-field">
                      <span>Run id</span>
                      <input v-model="resultRunId" class="script-input" type="text" autocomplete="off">
                    </label>
                    <label class="script-field">
                      <span>Duplicate</span>
                      <select v-model="resultDuplicatePolicy" class="script-input">
                        <option value="cancel">Cancel</option>
                        <option value="reuse">Reuse</option>
                        <option value="new_take">New take</option>
                      </select>
                    </label>
                  </div>
                  <div class="service-actions">
                    <Button size="small" type="primary" :loading="resultAction === 'import'" @click="importCurrentResult">
                      Import result
                    </Button>
                    <Button size="small" :loading="resultAction === 'list'" @click="listCurrentResults">
                      List results
                    </Button>
                    <Button
                      size="small"
                      :loading="resultAction === 'trace'"
                      :disabled="!activeResultId"
                      @click="traceCurrentResult"
                    >
                      Trace
                    </Button>
                  </div>
                  <Alert
                    v-if="resultReviewResult"
                    class="service-alert"
                    :type="resultStatusType"
                    show-icon
                    :message="resultSummary"
                    :description="resultReviewResult.result ? resultTraceLabel : `${resultReviewResult.error?.code} · ${resultReviewResult.error?.correlationId}`"
                  />
                  <dl v-if="activeResult" class="property-grid">
                    <div>
                      <dt>Asset</dt>
                      <dd :title="activeResult.assetId">{{ activeResult.assetId }}</dd>
                    </div>
                    <div>
                      <dt>Result file</dt>
                      <dd :title="activeResult.relativePath">{{ activeResult.relativePath }}</dd>
                    </div>
                    <div>
                      <dt>Record</dt>
                      <dd :title="activeResult.recordPath">{{ activeResult.recordPath }}</dd>
                    </div>
                    <div>
                      <dt>Review</dt>
                      <dd>{{ activeResult.reviewStatus }}</dd>
                    </div>
                  </dl>
                  <dl v-if="activeResultTrace" class="property-grid">
                    <div v-if="activeResultTrace.asset">
                      <dt>Trace asset</dt>
                      <dd :title="activeResultTrace.asset.relativePath">
                        {{ activeResultTrace.asset.id }} · {{ activeResultTrace.asset.role }}
                      </dd>
                    </div>
                    <div v-if="activeResultTrace.shot">
                      <dt>Trace shot</dt>
                      <dd :title="activeResultTrace.shot.description">
                        {{ activeResultTrace.shot.id }} · {{ activeResultTrace.shot.status }}
                      </dd>
                    </div>
                    <div v-if="activeResultTrace.package">
                      <dt>Trace package</dt>
                      <dd :title="activeResultTrace.package.manifestPath">
                        {{ activeResultTrace.package.packageId }} · {{ activeResultTrace.package.generationPackageStatus }}
                      </dd>
                    </div>
                    <div v-if="activeResultTrace.run">
                      <dt>Trace run</dt>
                      <dd :title="activeResultTrace.run.runPath">
                        {{ activeResultTrace.run.runId }} · {{ activeResultTrace.run.status }}
                      </dd>
                    </div>
                  </dl>
                  <div class="script-field-grid">
                    <label class="script-field">
                      <span>Review status</span>
                      <select v-model="resultReviewStatus" class="script-input">
                        <option value="pending">Pending</option>
                        <option value="approved">Approved</option>
                        <option value="needs_revision">Needs revision</option>
                        <option value="rejected">Rejected</option>
                      </select>
                    </label>
                    <label class="script-field">
                      <span>Reason</span>
                      <input v-model="resultReviewReason" class="script-input" type="text" autocomplete="off">
                    </label>
                  </div>
                  <label class="script-field">
                    <span>Review notes</span>
                    <textarea v-model="resultReviewNotes" class="script-textarea script-textarea--short" spellcheck="false" />
                  </label>
                  <div class="service-actions">
                    <Button
                      size="small"
                      :loading="resultAction === 'review'"
                      :disabled="!activeResultId"
                      @click="updateCurrentResultReview"
                    >
                      Save review
                    </Button>
                  </div>
                  <div class="script-field-grid">
                    <label class="script-field">
                      <span>Rebind shot</span>
                      <input v-model="resultRebindShotId" class="script-input" type="text" autocomplete="off">
                    </label>
                    <label class="script-field">
                      <span>Rebind package</span>
                      <input v-model="resultRebindPackageId" class="script-input" type="text" autocomplete="off">
                    </label>
                  </div>
                  <label class="script-field">
                    <span>Rebind reason</span>
                    <input v-model="resultRebindReason" class="script-input" type="text" autocomplete="off">
                  </label>
                  <div class="script-field-grid">
                    <label class="script-field">
                      <span>Unbind shot</span>
                      <input v-model="resultUnbindShot" type="checkbox">
                    </label>
                    <label class="script-field">
                      <span>Unbind package</span>
                      <input v-model="resultUnbindPackage" type="checkbox">
                    </label>
                  </div>
                  <div class="service-actions">
                    <Button
                      size="small"
                      :loading="resultAction === 'rebind'"
                      :disabled="!activeResultId || !resultRebindReason.trim()"
                      @click="rebindCurrentResult"
                    >
                      Rebind
                    </Button>
                  </div>
                  <List
                    v-if="resultRows.length"
                    class="recovery-list"
                    size="small"
                    :data-source="resultRows"
                  >
                    <template #renderItem="{ item }">
                      <ListItem>
                        <span>{{ item.id }} · {{ item.status }} · take {{ item.takeNumber || 'pending' }}</span>
                        <Tag :color="item.missing ? 'error' : 'processing'">{{ item.reviewStatus }}</Tag>
                      </ListItem>
                    </template>
                  </List>
                  <List
                    v-if="resultRecoveryActions.length"
                    class="recovery-list"
                    size="small"
                    :data-source="resultRecoveryActions"
                  >
                    <template #renderItem="{ item }">
                      <ListItem>
                        <span>{{ item }}</span>
                      </ListItem>
                    </template>
                  </List>
                </section>

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
                <Alert
                  v-if="mockRunResult?.run"
                  class="service-alert"
                  :type="mockRunStatusType"
                  show-icon
                  :message="mockRunSummary"
                  :description="mockRunOutputLabel"
                />
                <Alert
                  v-if="activeResult"
                  class="service-alert"
                  :type="resultStatusType"
                  show-icon
                  :message="resultSummary"
                  :description="resultTraceLabel"
                />
                <dl v-if="mockRunResult?.run" class="property-grid">
                  <div>
                    <dt>Status</dt>
                    <dd>{{ mockRunResult.run.status }}</dd>
                  </div>
                  <div>
                    <dt>Attempt</dt>
                    <dd>{{ mockRunResult.run.attempt }}</dd>
                  </div>
                  <div>
                    <dt>Run record</dt>
                    <dd :title="mockRunResult.run.runPath">{{ mockRunResult.run.runPath }}</dd>
                  </div>
                  <div v-if="mockRunResult.run.output">
                    <dt>Output digest</dt>
                    <dd :title="mockRunResult.run.output.digest">{{ mockRunResult.run.output.digest.slice(0, 16) }}</dd>
                  </div>
                </dl>
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
                <List v-if="operationalEvents.length" class="event-list" size="small" :data-source="operationalEvents">
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
                <div v-if="!runtimeEvents.length && !operationalEvents.length" class="placeholder-list">
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
                  v-if="auditHealthItems.length"
                  class="recovery-list"
                  size="small"
                  :data-source="auditHealthItems"
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
          <Progress class="queue-progress" :percent="runtimeProgress" size="small" />
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
