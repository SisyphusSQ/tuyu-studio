# Provider 配置与凭据

## 目标与边界

定义 provider profile、项目引用、项目级 override、credentialRef 和 API key 存储策略。本文只约束配置和凭据，不定义具体 provider 的私有接口字段。

## 输入 / 输出

| 输入 | 输出 |
| --- | --- |
| 全局 provider profile | 可复用 provider 能力和交接规则 |
| 项目默认 profile 引用 | 当前项目使用哪个 provider profile |
| 系统密钥存储 | 运行时可读取的 API key 或 token |
| 项目级 override | 非敏感的模型偏好、能力开关、包规则或预算标签 |

## 核心对象或规则

Provider 默认是多项目复用配置，不按项目复制一份完整配置。项目只保存 `providerProfileId`，全局 profile 保存非敏感配置和 `credentialRef`，真实 API key 只进入系统密钥存储。

```text
~/.tuyu-studio/
└─ config/
   ├─ app_settings.json
   └─ provider_profiles/
      ├─ openai-default.json
      ├─ deepseek-text.json
      └─ video-handoff.json

projects/{project_id}/
└─ project.tuyu.json
```

全局 profile 示例：

```json
{
  "id": "openai-default",
  "name": "OpenAI Default",
  "provider": "openai",
  "credentialRef": "tuyu-studio/provider/openai-default",
  "mode": "api_submit",
  "capabilityFlags": {
    "textUnderstanding": true,
    "imageUnderstanding": true,
    "videoUnderstanding": false,
    "imageGeneration": true,
    "videoGeneration": false,
    "apiSubmit": true,
    "pollResult": false
  },
  "packageRules": {
    "includeManifest": true,
    "includeScriptExcerpt": true,
    "includeContinuity": true,
    "includeUploadChecklist": true
  }
}
```

项目引用示例：

```json
{
  "defaults": {
    "providerProfileId": "openai-default"
  }
}
```

项目级 override 只允许非敏感字段：

```json
{
  "providerOverrides": {
    "openai-default": {
      "modelPreference": "image-generation-default",
      "budgetLabel": "client-a",
      "packageRules": {
        "includeManifest": true,
        "includeUploadChecklist": true
      },
      "capabilityFlags": {
        "videoGeneration": false,
        "apiSubmit": true
      }
    }
  }
}
```

禁止项目 override 包含 `apiKey`、`token`、`secret`、`password`、`credentialValue`、`refreshToken` 或任何可直接认证的字段。

配置解析顺序：

```text
1. 读取 project.tuyu.json.defaults.providerProfileId
2. 从 ~/.tuyu-studio/config/provider_profiles/{id}.json 读取全局 profile
3. 合并项目级非敏感 override
4. 用 profile.credentialRef 从系统密钥存储读取凭据
5. 创建运行时 adapter
```

API key 不写入：

- `project.tuyu.json`
- 领域对象文件
- `prompts/runs/`
- `audit/events.jsonl`
- `packages/*/manifest.json`
- app log
- 支持包或脱敏报告

## Provider Settings 生命周期

Provider Settings 是应用全局设置，不是项目设置。它管理：

- provider profile 列表。
- profile 名称、类型、base URL、模型偏好。
- capability flags。
- package rules。
- timeout、retry、rate limit。
- credentialRef。
- 凭据状态，例如 `configured`、`missing`、`store_unavailable`。

设置页可以显示“已配置 / 缺失 / 需更新”和脱敏摘要，例如 `last4`，但不显示真实 API key、token 或 refresh token。项目创建和项目设置页只选择默认 profile，并可保存非敏感 override；不在项目设置里录入 API key。

### 凭据写入规则

Credential 写入只走 Go 后端：

1. 前端表单收集 secret。
2. 前端调用 provider config API。
3. Go 后端写系统密钥存储。
4. 后端返回脱敏状态，例如 `configured: true` 和可选 `last4`。
5. 前端立即清空输入框。
6. audit 只记录 credential configured / rotated，不记录值。

`credentialRef` 命名保持稳定，建议使用：

```text
tuyu-studio/provider/{profileId}/{credentialKind}
```

示例：

- `tuyu-studio/provider/openai-default/api-key`
- `tuyu-studio/provider/video-prod/token`

profile 的 `id` 不应随意修改。重命名只改 display name，避免 credentialRef 漂移。

### Key rotation

Key rotation 是更新系统密钥存储里的 secret value，不改项目文件：

- `credentialRef` 不变。
- profile 非敏感配置不变。
- 后续任务读取新值。
- 旧 run 只记录使用过的 profile、adapter 和配置版本摘要，不记录 key。

### Profile 导入导出

如果未来支持 profile export/import：

- 可以导出 provider、capability、模型、base URL、package rules、timeout、retry。
- 可以导出 `credentialRef` 名称，或导出为空等待本机重新配置。
- 不导出真实 API key、token、refresh token。
- 导入后如果 credential 缺失，进入 `provider_credential_missing`，不静默运行失败。

### 多项目共享影响

多个项目默认复用同一个全局 profile。修改 profile capability、base URL、模型偏好或 package rules 可能影响多个项目的任务入口和交接格式。项目级 override 只能覆盖非敏感字段，避免为了单个项目复制完整 profile 或复制 key。

## 状态推进

| 状态 | 进入条件 | 输出 |
| --- | --- | --- |
| profile_missing | 项目引用的 profile 不存在 | 提示用户创建或选择 profile |
| credential_missing | profile 存在但 keyring 中无凭据 | 提示用户配置凭据 |
| profile_ready | profile 与 credentialRef 都可解析 | 允许按 capability 显示任务入口 |
| override_invalid | 项目 override 含敏感字段或非法 capability | 阻断项目打开或禁用对应 provider |
| credential_rotated | 用户更新系统密钥 | 后续任务使用新凭据，不改项目文件 |

## 错误语义

| 错误 | 用户提示 | 系统行为 |
| --- | --- | --- |
| provider_profile_missing | 当前项目引用的 provider profile 不存在 | 阻断自动任务，允许手动交接 |
| provider_credential_missing | 当前 provider 缺少 API key | 打开设置页配置凭据 |
| provider_override_rejected | 项目配置含敏感或非法字段 | 拒绝应用 override 并显示字段 |
| credential_store_unavailable | 系统密钥存储不可用 | 禁用需要凭据的任务 |
| credential_leak_detected | 项目、日志或包中检测到疑似凭据 | 阻断导出或支持包生成 |

## 恢复与重试

- profile 缺失时，用户可以创建同名全局 profile、切换到已有 profile，或把任务降级为手动交接。
- credential 缺失时，用户在应用设置中写入系统密钥存储；项目文件不变。
- override 被拒绝时，保留原始项目文件并提示删除敏感字段或非法字段。
- 系统密钥更新后不需要迁移项目；后续运行重新从 keyring 读取。
- 导出支持包前运行凭据扫描，发现疑似凭据时先中止并给出脱敏建议。

## 验收标准

- 新项目只保存 `providerProfileId`，不保存 API key。
- `~/.tuyu-studio/config/provider_profiles/` 只保存非敏感 profile 和 `credentialRef`。
- API key 存入 macOS Keychain、Windows Credential Manager 或 Linux Secret Service 等系统密钥存储。
- 多项目默认可复用同一个 provider profile。
- 项目级 override 只能覆盖非敏感能力、包规则、模型偏好和预算标签。
- 缺少 profile、缺少 credential、keyring 不可用和 override 非法有不同错误码。
- PromptRun、audit、manifest、log、支持包和脱敏报告不含 API key 或 token。
