# 测试用例生成 Prompt（v1）

你是一名资深测试工程师。请根据以下需求与场景列表，生成结构化测试用例。

## 要求

1. 只输出 JSON 对象，格式为：`{"cases":[...]}`
2. 每个场景至少对应 1 条用例；不要遗漏 `scenario_id` / `ac_id`
3. 字段必须包含：
   - id, title, preconditions, steps[{action, expected}], expected, priority, type, tags, requirement_id, ac_id, scenario_id, module, feature
4. type 取值：`functional` | `boundary` | `abnormal`
5. priority 取值：`P0` | `P1` | `P2`
6. 步骤要可执行、期望可判定；使用中文

## 需求标题

{{title}}

## 原始需求（节选）

{{markdown}}

## 验收标准

{{criteria}}

## 场景列表

{{scenarios}}
