<script setup lang="ts">
import {
  Alert,
  Badge,
  Button,
  Divider,
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
  DatabaseOutlined,
  FileTextOutlined,
  FolderOpenOutlined,
  HistoryOutlined,
  PlayCircleOutlined,
  SaveOutlined,
  SearchOutlined,
  SettingOutlined,
  ShareAltOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons-vue'
import { h, ref } from 'vue'

import {
  type AppErrorDTO,
  type ProjectOperationName,
  type ProjectOperationResultDTO,
  type RuntimeEventDTO,
  type WorkbenchProbeMode,
  type WorkbenchStatusDTO,
} from './api/dto'
import { runProjectOperation, runWorkbenchProbe } from './api/workbench'

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

const canvasNodes = [
  { title: 'Script source', meta: '7 scenes detected', state: 'draft' },
  { title: 'Character reference group', meta: '2 locked rules', state: 'ready' },
  { title: 'Shot candidate table', meta: 'awaiting confirmation', state: 'blocked' },
]

const queueItems = [
  { label: 'Package export', value: 'idle' },
  { label: 'Mock run', value: 'not configured' },
  { label: 'Review import', value: 'waiting for TOO-180' },
]

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
          <Badge status="success" text="Saved 11:37" />
          <Badge status="warning" text="2 queue slots" />
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
          <Tooltip title="Save workspace snapshot">
            <Button size="small" aria-label="Save workspace snapshot">
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
              <div class="placeholder-list">
                <strong>Scene split queue</strong>
                <span>Manual scene and shot candidates start in TOO-172.</span>
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

        <LayoutContent class="canvas-shell" aria-label="Project Canvas placeholder">
          <section class="canvas-toolbar" aria-label="Canvas actions">
            <div>
              <p class="eyebrow">Project Canvas</p>
              <h2>Creative graph mount point</h2>
            </div>
            <div class="canvas-actions">
              <Tooltip title="Fit visible nodes">
                <Button size="small">Fit</Button>
              </Tooltip>
              <Tooltip title="Toggle graph links">
                <Button size="small">
                  <template #icon>
                    <ShareAltOutlined />
                  </template>
                </Button>
              </Tooltip>
              <Tooltip title="Run selected placeholder">
                <Button size="small">
                  <template #icon>
                    <PlayCircleOutlined />
                  </template>
                </Button>
              </Tooltip>
            </div>
          </section>

          <section class="canvas-stage">
            <div class="production-frame">
              <div class="frame-header">
                <Tag color="blue">ProductionFrame</Tag>
                <span>Shot context shell</span>
              </div>
              <div class="node-lane">
                <article
                  v-for="node in canvasNodes"
                  :key="node.title"
                  class="canvas-node"
                  :class="`state-${node.state.replace(' ', '-')}`"
                >
                  <div class="node-icon">
                    <ThunderboltOutlined />
                  </div>
                  <div>
                    <strong :title="node.title">{{ node.title }}</strong>
                    <span>{{ node.meta }}</span>
                  </div>
                </article>
              </div>
              <Alert
                class="canvas-alert"
                type="info"
                show-icon
                message="Canvas renderer intentionally reserved"
                description="TOO-161 proves the stable workbench zones. Graph rendering, drag, links, and persisted viewport belong to later Canvas issues."
              />
            </div>
          </section>
        </LayoutContent>

        <LayoutSider class="inspector" width="344" theme="light">
          <section class="inspector-header">
            <div>
              <p class="eyebrow">Inspector</p>
              <h2>Shot candidate table</h2>
            </div>
            <Tag color="warning">context draft</Tag>
          </section>

          <Tabs v-model:activeKey="activeInspectorTab" size="small">
            <TabPane key="properties" tab="Properties">
              <dl class="property-grid">
                <div>
                  <dt>Type</dt>
                  <dd>Canvas placeholder</dd>
                </div>
                <div>
                  <dt>Source</dt>
                  <dd>Workbench first screen</dd>
                </div>
                <div>
                  <dt>Binding</dt>
                  <dd>No Wails generated import in page components</dd>
                </div>
              </dl>
            </TabPane>
            <TabPane key="relations" tab="Relations">
              <div class="placeholder-list">
                <strong>Incoming references</strong>
                <span>Script, Character, Scene and Prop edges arrive after Graph DTO.</span>
              </div>
            </TabPane>
            <TabPane key="tasks" tab="Tasks">
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
                        <span v-if="item.path">
                          Path {{ item.path }}
                        </span>
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
                <List
                  v-if="projectResult?.events.length"
                  class="event-list"
                  size="small"
                  :data-source="projectResult.events"
                >
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
              </section>

              <List class="event-list" size="small" :data-source="runtimeEvents">
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
            </TabPane>
          </Tabs>

          <Divider />
          <Alert
            type="info"
            show-icon
            message="Go service bridge boundary"
            description="Page components call frontend/src/api/workbench.ts; generated Wails binding imports stay outside Workbench components."
          />
        </LayoutSider>
      </Layout>

      <footer class="bottom-bar" aria-label="Workbench status">
        <div class="bottom-group">
          <Badge status="processing" text="Selection: Shot candidate table" />
          <span>Zoom 100%</span>
          <span>Grid medium</span>
        </div>
        <div class="bottom-group">
          <Tag color="success">Health clean</Tag>
          <span>Events {{ runtimeEvents.length }} captured</span>
          <Progress class="queue-progress" :percent="runtimeEvents[0]?.progress ?? 24" size="small" />
        </div>
        <Button size="small">
          <template #icon>
            <BgColorsOutlined />
          </template>
          Restore layout
        </Button>
      </footer>
    </Layout>
  </main>
</template>
