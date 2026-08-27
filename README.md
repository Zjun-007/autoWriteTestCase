# autoWriteTestCase

根据需求文档，自动生成测试用例（功能 / 边界 / 异常 / 回归）的工具。

## 快速开始

```bash
# 依赖（如遇代理问题可：export GOPROXY=https://goproxy.cn,direct）
go mod tidy
go build -o bin/awtc ./cmd/awtc

# 默认 mock 生成器（无需 API Key）
./bin/awtc generate -i examples/login_requirement.md -o ./out -f json,excel,markdown
```

输出目录含 `cases.json` / `cases.xlsx` / `cases.md` 等。

接入真实 LLM：

```bash
export OPENAI_API_KEY=sk-...
# 或编辑 configs/default.yaml：llm.provider: openai
./bin/awtc generate -i examples/login_requirement.md -o ./out
```

开发计划见 [PLAN.md](./PLAN.md)。

## 语言选择

### 结论：核心逻辑可以用 Go

**可以。** 流水线编排、领域模型、校验、追溯、导出适配器都适合用 Go 写；LLM 调用本质是 HTTP + JSON Schema，Go 完全够用。

| 维度 | Go 做核心 | Python 做核心 |
|------|-----------|---------------|
| 适合场景 | 团队主栈是 Go、要单二进制分发、CLI/服务长期维护 | 要最快验证 Prompt、重度文档解析、重度 Agent 编排 |
| LLM | 官方 HTTP SDK / 社区 client + JSON Schema 约束输出 | LangChain / instructor 等更现成 |
| 文档解析 | Markdown/纯文本/JSON 好做；Word/PDF/Excel 库偏少，复杂格式可外挂脚本 | `docx`/`pypdf`/`openpyxl` 更省事 |
| 工程形态 | 单文件部署、并发强、类型与接口清晰 | 原型快，依赖与运行环境更重 |
| 风险点 | 复杂 Agent / 多轮工具调用要自研更多 | 部署与类型约束弱于 Go |

**建议怎么选：**

- 团队熟悉 Go、产品要做成 **CLI / 内部服务** → **核心用 Go**（推荐）
- 先快速试 Prompt 质量、大量非结构化文档 → 可先用 Python 做 MVP，稳定后再迁 Go；或 **Go 核心 + 少量 Python sidecar** 只处理 Word/PDF

**前端 / 插件（可选）：** TypeScript（Web UI / VS Code 插件），通过 API 调 Go 服务。

---

## 推荐架构：流水线 + 可插拔适配器

```
需求输入 → 解析归一化 → 场景分析 → LLM 生成 → 校验/去重 → 导出
                ↑              ↑
           Adapter 层      Prompt / Schema
```

核心逻辑（Go）负责除「超复杂文档解析」以外的全部阶段；各阶段用 `interface` 解耦，便于替换 LLM Provider 与导入导出格式。

### 分层职责

1. **Ingest（输入层）**  
   适配器读取：Markdown / Word / PDF / Excel / Jira / 飞书 / 纯文本。  
   统一产出 `RequirementDoc`（标题、描述、验收标准、优先级、关联接口等）。

2. **Parse & Normalize（解析层）**  
   切分模块、用户故事、验收条件；抽取实体、规则、前置条件。  
   输出结构化 `RequirementGraph`（模块 → 功能点 → 约束）。

3. **Analyze（分析层）**  
   规则 + LLM 混合：等价类、边界值、异常路径、权限矩阵、状态迁移。  
   产出 `TestScenario[]`（场景意图，尚未写成具体步骤）。

4. **Generate（生成层）**  
   按 JSON Schema 调用 LLM，生成标准用例字段：  
   `id / title / preconditions / steps / expected / priority / type / tags`。  
   支持按目标框架二次渲染（pytest、Postman、Allure、Excel）。

5. **Validate（校验层）**  
   - Schema 校验、必填字段  
   - 与需求追溯（每条 AC 至少一条用例）  
   - 去重、冲突检测  
   - 可选：人工审核 / 二次 LLM review

6. **Export（导出层）**  
   适配器写出：Excel、Markdown、JSON、pytest 骨架、TestRail / Zephyr API 等。

### Go 技术选型

| 组件 | 选型 |
|------|------|
| CLI | `cobra` 或 `urfave/cli` |
| API（可选） | 标准库 `net/http` 或 `chi` / `gin` / `echo` |
| 配置 | `viper` 或 YAML + env |
| 领域模型 / 校验 | `struct` + `go-jsonschema` / `validator` |
| LLM | OpenAI/Anthropic 兼容 HTTP client；强制 JSON Schema / tool call |
| 编排 | 自研 pipeline（阶段函数 + interface）；复杂再考虑工作流引擎 |
| 存储（可选） | SQLite / Postgres（历史用例、追溯矩阵） |
| 可观测 | `slog` + prompt/用例版本号 |

### 目录结构建议（Go）

```text
autoWriteTestCase/
├── cmd/
│   └── awtc/                 # CLI 入口
├── internal/
│   ├── ingest/               # 各输入 Adapter
│   ├── parse/                # 归一化与切分
│   ├── analyze/              # 场景推导
│   ├── generate/             # Prompt + LLM + Schema
│   ├── validate/             # 校验与追溯
│   ├── export/               # 各输出 Adapter
│   ├── model/                # 领域结构体
│   └── pipeline/             # 流水线编排
├── prompts/                  # 版本化 Prompt 模板
├── configs/
├── api/                      # 可选 HTTP 服务
└── README.md
```

### 核心设计原则

- **Schema First**：用例用 Go struct + JSON Schema 定义，LLM 只填结构化字段。
- **Adapter First**：输入/输出用 interface，流水线不绑死文档或测试平台。
- **可追溯**：每条用例绑定 `requirement_id` / `acceptance_criteria_id`。
- **人机协同**：默认「生成 → 校验 → 人工确认 → 导出」。
- **Prompt 版本化**：Prompt 与模型参数入 Git，便于回归对比生成质量。

---

## 推荐落地路径

1. **MVP**：Go CLI + Markdown 输入 → JSON/Excel 用例（功能 + 边界 + 异常）。  
2. **增强**：验收标准追溯矩阵、去重、人工 review 标记。  
3. **产品化**：HTTP API、对接 Jira/飞书、导出 pytest/接口用例、Web 审核台。

---

## 一句话结论

**核心逻辑可以用 Go**：流水线、校验、追溯、导出都适合；LLM 用 HTTP + JSON Schema 即可。架构仍是 **「输入适配 → 解析 → 场景分析 → Schema 约束生成 → 校验追溯 → 导出」**。复杂 Word/PDF 解析不够用时，再单独加 Python sidecar，不必把整个核心改回 Python。
