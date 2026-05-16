# 发布就绪

## 目标与边界

定义发布门禁、迁移准备、脱敏、包可移植、恢复证明和 provider 可替换性。本文用于发布前检查，不替代日常切片验收。

## 输入 / 输出

| 输入 | 输出 |
| --- | --- |
| 切片验收结果 | 发布门禁清单 |
| schema 和迁移脚本 | 迁移准备结论 |
| 文档和测试摘要 | 脱敏检查结果 |
| GenerationPackage 样例 | 可移植性证明 |
| 恢复演练记录 | 恢复证明 |
| provider profile 与 adapter | 可替换性结论 |

## 核心对象或规则

发布门禁：

| 门禁 | 标准 |
| --- | --- |
| 功能闭环 | 核心生产路径从项目创建到结果 review 可完成 |
| 数据持久化 | 重启后对象、路径、状态、审计仍一致 |
| 错误语义 | 关键失败有错误码、用户提示和恢复入口 |
| 安全 | 路径守卫、外部确认、凭据排除通过 |
| 可观测 | PromptRun、包校验、健康报告、审计可读 |
| 文档 | README 和子文档同步，提交版测试摘要脱敏 |

迁移准备：

- schema version 有升级路径和迁移前备份。
- 迁移失败保留原项目可打开或可回滚。
- 跨目录打开后项目相对路径仍有效。
- 迁移报告记录对象数量、失败项和修复建议。

脱敏：

- release notes、docs/test、问题报告不包含私有素材、凭据、机器私有路径或完整外部地址。
- 示例路径使用 `{studio_root}`。
- 错误样例使用对象 ID、错误码和摘要。

包可移植：

- GenerationPackage 内 manifest 使用相对路径。
- 引用文件清单完整，摘要校验通过。
- 包复制到其他目录后仍可校验。
- 包内不包含逃逸路径或项目外符号链接。

恢复证明：

- 保存失败、任务失败、包导出失败、资产缺失、迁移失败都至少有一条演练记录。
- 恢复动作包含确认、执行结果和审计事件。

provider 可替换性：

- 核心领域对象不绑定单一外部能力。
- profile 和 adapter 可替换，不改变 Shot、GenerationPackage、PromptRun 的核心语义。
- 输出契约由项目 schema 和任务模式定义。

## 状态推进

| 状态 | 推进规则 |
| --- | --- |
| readiness_planned | 门禁清单确定 |
| evidence_collecting | 收集测试、审计、迁移和包校验证据 |
| blocked | 任一门禁失败 |
| ready | 所有门禁通过且证据脱敏 |
| released | 发布完成并归档摘要 |

## 错误语义

| 错误 | 语义 |
| --- | --- |
| release_gate_failed | 发布门禁未通过 |
| migration_proof_missing | 缺少迁移或回滚证据 |
| redaction_gate_failed | 发布材料含敏感内容 |
| portability_failed | 交接包复制后校验失败 |
| recovery_proof_missing | 缺少恢复演练证据 |
| provider_coupling_detected | 核心对象绑定单一外部能力 |

## 恢复与重试

- release_gate_failed 需回到对应切片补验收。
- migration_proof_missing 补迁移演练和回滚记录。
- redaction_gate_failed 先修正发布材料，再重新扫描。
- portability_failed 重新导出包并验证 manifest 和引用。
- recovery_proof_missing 补保存、任务、导出、资产、迁移失败演练。
- provider_coupling_detected 需要把差异收敛到 profile 或 adapter。

## 验收标准

- 发布前所有门禁都有通过证据和脱敏摘要。
- 迁移失败有可验证回滚路径。
- 交接包复制到其他目录后可校验并保持相对路径。
- 恢复证明覆盖保存、运行、导出、资产缺失和迁移失败。
- provider 替换不改变核心领域对象和状态语义。
