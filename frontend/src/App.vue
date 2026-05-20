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

const LayoutHeader = Layout.Header
const LayoutContent = Layout.Content
const LayoutSider = Layout.Sider
const ListItem = List.Item
const ListItemMeta = List.Item.Meta
const TabPane = Tabs.TabPane

const themeMode = ref<'dark' | 'warm'>('dark')
const activeRailTab = ref('assets')
const activeInspectorTab = ref('properties')

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
            </TabPane>
          </Tabs>

          <Divider />
          <Alert
            type="warning"
            show-icon
            message="Known UI limit"
            description="This card only creates the production shell. Structured Go service calls and Error/Event DTO display are handled by TOO-162."
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
          <span>Events 0 running</span>
          <Progress class="queue-progress" :percent="24" size="small" />
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
