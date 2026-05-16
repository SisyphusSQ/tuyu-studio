# 测试策略

## 目标与边界

定义单元、集成、手动验收、文档校验和回归命令。本文只定义测试层级和入口，不写具体实现代码。

## 输入 / 输出

| 输入 | 输出 |
| --- | --- |
| 设计模块和切片 | 测试范围 |
| 领域对象、错误语义 | 单元和集成断言 |
| 用户生产路径 | 手动验收脚本 |
| 文档约束 | 文档检查命令 |
| 回归入口 | 可重复执行的命令 |

## 核心对象或规则

测试分层：

### Go Domain / Service 单测

不启动 Wails，直接测试领域和应用 service：

| 模块 | 重点 |
| --- | --- |
| ProjectStore | 原子写、备份、schema 迁移、恢复 |
| PathGuard | 规范化、越界拒绝、符号链接、外部确认、授权范围 |
| GraphDomain | 节点边校验、删除策略、version 冲突 |
| AssetLibrary | 摘要、重复检测、绑定、缺失检查 |
| InstructionCompiler | 编译顺序、data 包装、冲突检测、上下文摘要 |
| RuntimeGateway | 状态转换、取消、超时、重试、错误映射 |
| PackageExporter | manifest、相对路径、引用完整性 |
| ProviderConfig | profile 解析、credentialRef、override 白名单、keyring 缺失和凭据泄漏检测 |

### DTO / Contract 测试

DTO 是前后端稳定契约，必须有 golden、schema 或等价结构检查：

- Command DTO 字段不漂移，不能接收内部领域对象。
- View DTO 不裸露内部领域对象，只暴露 UI 需要的投影、摘要和下一步动作。
- Error DTO 必含 `code`、`retryable`、`targetType`、`targetId`、`recoveryActions` 或等价字段。
- Event DTO 可按 `eventId` 幂等去重，并能重建运行队列。
- Go 侧 DTO 变更后，前端 `vue-tsc` 必须能发现类型问题。

### Provider Fake / Adapter 测试

Provider adapter 不打真实外部 API，先用 fake transport 或 fake adapter：

- capability 缺失时禁用入口或返回 `provider_capability_missing`。
- credential 缺失时返回 `provider_credential_missing`。
- submit、poll、download 覆盖成功、超时、限流、输出无效。
- adapter 不写项目文件，只返回结果或错误。
- RuntimeGateway 负责写 run、attempt、audit 和 result。

### 前端组件测试

前端测试只验证 UI 行为和 DTO 渲染：

- 错误 DTO 渲染 recovery action。
- RunQueue 根据事件更新状态，重复事件不重复显示。
- ProviderSettings 不展示真实 key，提交后清空输入框。
- GraphCanvas 交互产出 command，不直接改项目 truth。
- 长路径、长标题、长错误不撑破布局。

### Wails Bridge Smoke

少量关键链路覆盖 Wails v2 binding 和事件通道：

- 前端调用 `ProjectOpen` 或等价查询。
- Go 返回 `ProjectSummaryDTO`。
- Go 推送一个 runtime event。
- 前端 run queue 收到并展示。
- bridge 异常由 API wrapper 转成 `transport_error`。

集成测试：

- Wails bridge smoke：前端调用 Go 查询方法，Go 推送事件，前端运行队列可见。
- 项目重启恢复：新建、编辑、导入、关闭、重开、校验。
- 资产到镜头：导入角色图和场景图，绑定 Shot，确认 context_ready。
- 指令任务：创建 PromptRun，校验 compiled 摘要、contextDigest、schema 和输出。
- 导出包：从 Shot 导出包，校验 manifest 和引用文件。
- 结果回收：导入结果文件，绑定 Shot，切换 review 状态。
- 健康检查：删除资产副本，运行检查，确认影响对象列表。
- Provider 配置：项目引用全局 profile，缺失 profile 和缺失 credential 分别提示，API key 不进入项目文件。

手动验收：

1. 新建项目。
2. 导入角色参考图、场景参考图和结果视频占位文件。
3. 创建角色、场景和 Shot。
4. 绑定资产并锁定一条连续性规则。
5. 预览指令栈。
6. 运行一次结构化文本任务或受控模拟运行。
7. 导出 Shot 交接包。
8. 回收结果并标记 review。
9. 重启后确认对象、状态、路径和审计仍存在。

文档验证：

- 技术栈文档与交付切片一致，Wails、Go、Vue、Ant Design Vue、AntV 的职责边界不互相覆盖。
- SDK 清单与 provider 配置文档一致，前端不得引入 AI SDK 或文件系统 SDK 绕过 Wails binding。
- 设计文档不包含禁用供应商名、降级口径、机器私有路径或凭据。
- 每篇子文档包含固定七个二级章节。
- README 是模块总览，子文档承接详细规则。
- 路径示例使用 `{studio_root}` 或项目相对路径。

跨平台桌面 smoke：

- macOS、Windows、Linux 分别启动 Wails App Shell。
- 验证窗口启动、路由切换、Ant Design Vue 表单/抽屉/通知、AntV 画布渲染、快捷键、拖拽、文件选择和事件推送。
- 验证长项目名、长节点标题、长错误详情和窄屏折叠不撑破布局。

Provider 配置验收：

- 多项目默认复用同一个全局 provider profile。
- 项目只保存 `providerProfileId` 和非敏感 override。
- API key 存在系统密钥存储，不存在项目、PromptRun、audit、manifest、log 或支持包。
- profile 缺失、credential 缺失、keyring 不可用和 override 非法有不同错误码。
- 项目迁移到另一台机器后，缺失 profile 或 credential 时提示配置，不静默运行失败。

切片验收门禁：

- `go test ./...` 通过。
- `vue-tsc` 通过。
- 前端 build 通过。
- Wails bridge smoke 通过，至少覆盖当前切片新增 binding。
- 新增 DTO 有测试、golden 或 schema 检查。
- 新增错误码进入 error registry。
- 涉及 provider、key、path、export 的变更必须有安全和脱敏测试。
- 涉及设计约定变化时，同步 `docs/design` 和 `tuyu-design` 页面。

回归命令：

```bash
make harness-check
make harness-review-gate PLAN=.agents/plans/2026-05-14-design-docs-production-refine.md
git diff --check -- docs/design
find docs/design -maxdepth 3 -type f -print | sort
```

## 状态推进

| 状态 | 推进规则 |
| --- | --- |
| test_planned | 测试范围和验收项已映射 |
| test_ready | 数据夹具、命令或手动脚本可执行 |
| test_running | 正在执行验证 |
| test_failed | 记录失败原因和影响能力 |
| test_passed | 记录命令、摘要和脱敏结果 |

## 错误语义

| 错误 | 语义 |
| --- | --- |
| test_fixture_missing | 缺少可复用测试数据 |
| command_unavailable | 回归命令不可执行 |
| manual_step_blocked | 手动验收路径被阻断 |
| docs_validation_failed | 文档结构或敏感词检查失败 |
| regression_failed | 已有能力回归 |

## 恢复与重试

- fixture 缺失时补可提交的脱敏样例或生成脚本。
- command_unavailable 需要更新文档中的验证入口或补脚本。
- manual_step_blocked 记录阻断步骤，修复后从该步骤重试。
- docs_validation_failed 先修正文档，再重新执行扫描。
- regression_failed 保留失败摘要，修复后跑相关层级和仓库级回归。

## 验收标准

- 每个切片至少有单元、集成、手动或文档检查中的一个稳定验证入口。
- 关键能力的错误和恢复路径进入测试范围。
- 回归命令可复制执行，输出可写入脱敏验证摘要。
- 文档检查能覆盖固定章节、禁用词、路径占位和 README/子文档边界。
