# autoWriteTestCase 开发计划

基于 [README.md](./README.md) 的技术选型：**Go 核心 + 流水线架构**。目标是先跑通「需求 → 结构化用例」闭环，再增强质量与集成能力。

---

## 目标与范围

### 产品目标

输入需求文档，自动产出可评审、可追溯的测试用例（功能 / 边界 / 异常），并支持导出到常见格式。

### 非目标（首期不做）

- 全自动写入生产测试平台（无人工确认）
- 端到端 UI 自动化脚本生成
- 复杂多 Agent 自主探索
- 完整 Web 审核台（放到产品化阶段）

### 成功标准（MVP）

- 给定一份 Markdown 需求，能一键生成 JSON + Excel 用例
- 每条用例含标准字段，并绑定需求/验收点 ID
- CLI 可本地运行；Prompt 与模型可配置
- 有一组样例需求 + 黄金用例，用于回归生成质量

---

## 阶段总览

| 阶段 | 名称 | 周期（建议） | 产出 |
|------|------|--------------|------|
| P0 | 工程骨架 | 3–5 天 | 可编译的 Go CLI、目录与接口约定 |
| P1 | MVP 闭环 | 2–3 周 | Markdown → 生成 → 校验 → JSON/Excel |
| P2 | 质量增强 | 2–3 周 | 追溯矩阵、去重、review、Prompt 回归 |
| P3 | 集成扩展 | 3–4 周 | 更多输入输出、HTTP API、可选 sidecar |
| P4 | 产品化 | 按需 | 平台对接、审核台、存储与权限 |

---

## P0：工程骨架（Week 0）

### 任务

1. 初始化 Go module、`cmd/awtc`、`internal/*` 目录
2. 定义核心领域模型（`RequirementDoc` / `AcceptanceCriteria` / `TestCase` / `TestScenario`）
3. 定义阶段接口：`Ingester` / `Parser` / `Analyzer` / `Generator` / `Validator` / `Exporter`
4. 实现空 pipeline：按顺序调用各阶段，统一错误与日志（`slog`）
5. 配置加载：YAML + 环境变量（模型、API Key、温度、输出目录）
6. 约定 Prompt 目录与版本命名（如 `prompts/v1/generate_cases.md`）

### 交付物

- `go build ./cmd/awtc` 通过
- `awtc version` / `awtc generate --help` 可用
- `internal/model` 与接口文档（代码注释即可）

### 验收

- 新成员按 README 目录能定位各层职责
- 无真实 LLM 时可用 mock Generator 跑通 pipeline

---

## P1：MVP 闭环（Week 1–3）

### 1.1 输入与解析

- [ ] Markdown Ingester（本地文件）
- [ ] 轻量 Parser：按标题层级切分模块 / 功能点 / 验收标准列表
- [ ] 为每条 AC 分配稳定 ID（如 `AC-001`）

### 1.2 场景分析（规则优先）

- [ ] 规则引擎雏形：从 AC 推导「正常 / 边界 / 异常」三类场景骨架
- [ ] 可选：一次 LLM 调用补全遗漏场景（失败则降级为纯规则）

### 1.3 LLM 生成

- [ ] OpenAI 兼容 HTTP Client（支持自定义 base URL）
- [ ] JSON Schema / 结构化输出约束 `TestCase[]`
- [ ] Prompt v1：输入需求片段 + 场景列表 → 输出用例
- [ ] 重试、超时、速率限制基础处理

### 1.4 校验与导出

- [ ] 必填字段与 Schema 校验
- [ ] 简单追溯：每条 AC 至少 1 条用例（不足则告警）
- [ ] Exporter：`json`、`excel`（xlsx）、`markdown`
- [ ] CLI：`awtc generate -i req.md -o ./out -f json,excel`

### 交付物

- 可演示的端到端命令
- `examples/` 下 1–2 份样例需求与期望输出说明
- Prompt v1 + 配置样例 `configs/default.yaml`

### 验收

- 样例需求生成用例可人工评审通过率 ≥ 70%（结构正确、步骤可读）
- 无 Key 时有清晰错误提示；有 Key 时稳定产出 JSON

---

## P2：质量增强（Week 4–6）

### 任务

- [ ] 验收标准追溯矩阵（AC × Case）导出
- [ ] 用例去重（标题/步骤相似度或 embedding，先做启发式）
- [ ] 冲突检测：同一 AC 下期望结果矛盾时标记
- [ ] `--review` 模式：输出待确认清单，支持人工改稿后 `awtc export`
- [ ] Prompt 版本对比：固定样例集，记录通过率/缺失覆盖率
- [ ] 用例类型标签完善：功能 / 边界 / 异常 / 权限 / 状态迁移
- [ ] 单元测试：parser、validator、export；pipeline 集成测试（mock LLM）

### 交付物

- `awtc trace` / 矩阵文件
- Prompt 回归脚本或 `make eval`
- 测试覆盖核心包

### 验收

- AC 覆盖率可量化；回归样例集可重复跑
- 重复用例明显减少；导出格式稳定

---

## P3：集成扩展（Week 7–10）

### 输入扩展

- [ ] 纯文本 / JSON 需求格式
- [ ] Excel 需求表（固定列模板）
- [ ] （可选）Python sidecar：Word/PDF → Markdown，Go 通过子进程或本地 HTTP 调用

### 输出扩展

- [ ] pytest 测试骨架（或接口用例 JSON / Postman collection）
- [ ] 对接一种用例平台 API（如 TestRail / 自建）——先做 dry-run

### 服务化

- [ ] HTTP API：`POST /generate`（上传或传路径）、异步任务可选
- [ ] 鉴权（至少 API Token）、请求日志、生成任务 ID

### 交付物

- API OpenAPI 文档
- sidecar 接口约定（若启用）
- 至少一种「代码/集合」导出

### 验收

- CLI 与 API 共用同一 pipeline
- 新 Adapter 不改 pipeline 主流程即可接入

---

## P4：产品化（按需）

- Web 审核台：展示用例、编辑、确认、批量导出
- Jira / 飞书需求同步
- 历史用例库（SQLite/Postgres）、版本对比
- 权限、租户、审计日志
- 观测：Prompt 版本、模型、token 消耗、生成耗时看板

---

## 里程碑与检查点

| 里程碑 | 检查点问题 |
|--------|------------|
| M1 骨架完成 | 能否用 mock 跑通全链路？ |
| M2 MVP 演示 | 真实需求能否生成可评审 Excel？ |
| M3 质量达标 | AC 覆盖与样例回归是否可量化？ |
| M4 可集成 | 是否有稳定 API + 至少两种新 IO？ |
| M5 可运营 | 是否支持审核、存储与平台同步？ |

---

## 风险与应对

| 风险 | 影响 | 应对 |
|------|------|------|
| LLM 输出不稳定 | 用例字段缺失/胡编 | Schema 强制 + 校验重试 + 降级规则场景 |
| 需求文档格式混乱 | 解析失败 | MVP 只保证 Markdown 模板；提供写法规范 |
| Word/PDF 难解析 | 阻塞输入 | P3 再上 sidecar，不阻塞 Go 核心 |
| Prompt 难评估 | 质量主观 | 固定黄金集 + 覆盖率指标 |
| API 成本/延迟 | 体验差 | 分块生成、缓存解析结果、可配模型 |

---

## 人力与分工建议（参考）

| 角色 | 职责 |
|------|------|
| 后端（Go） | pipeline、模型、CLI/API、导出 |
| 测试 / QA | 样例需求、黄金用例、评审标准、回归集 |
| （可选）前端 | P4 审核台 |
| （可选）脚本 | Word/PDF sidecar |

单人推进时：严格按 P0 → P1 → P2 顺序，P3 只做「当前真需要」的 Adapter。

---

## 近期两周执行清单（可直接开工）

**Week 1**

1. 建仓库结构与领域模型  
2. Markdown ingest + parse + AC ID  
3. mock generator + JSON export  
4. CLI `generate` 跑通  

**Week 2**

1. 接入真实 LLM + Schema  
2. Prompt v1 + 规则场景三类  
3. Excel/Markdown 导出 + 基础校验  
4. 准备 2 个样例需求并人工评一轮  

---

## 文档约定

- 架构与选型：`README.md`
- 本计划：`PLAN.md`（阶段完成后在对应 checkbox 勾选或追加变更记录）
- Prompt 变更：在 `prompts/` 下升版本，并在本文件「变更记录」注明

### 变更记录

| 日期 | 变更 |
|------|------|
| 2026-08-27 | 初版计划：Go 核心，P0–P4 分阶段落地 |
