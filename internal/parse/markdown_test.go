package parse_test

import (
	"testing"

	"github.com/webnode/autoWriteTestCase/internal/model"
	"github.com/webnode/autoWriteTestCase/internal/parse"
)

func TestMarkdownParser_ExtractsAC(t *testing.T) {
	doc := model.RequirementDoc{
		SourcePath: "login.md",
		Title:      "用户登录",
		RawMarkdown: `# 用户登录

## 功能：账号密码登录

### 验收标准

- 正确密码应登录成功
- 错误密码应提示失败
`,
	}
	graph, err := parse.NewMarkdown().Parse(doc)
	if err != nil {
		t.Fatal(err)
	}
	if len(graph.Criteria) != 2 {
		t.Fatalf("want 2 criteria, got %d", len(graph.Criteria))
	}
	if graph.Criteria[0].ID != "AC-001" {
		t.Fatalf("want AC-001, got %s", graph.Criteria[0].ID)
	}
}
