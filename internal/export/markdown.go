package export

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/webnode/autoWriteTestCase/internal/model"
)

// MarkdownExporter writes a human-readable cases.md.
type MarkdownExporter struct{}

func NewMarkdown() *MarkdownExporter { return &MarkdownExporter{} }

func (e *MarkdownExporter) Name() string { return "markdown" }

func (e *MarkdownExporter) Export(outDir string, result model.GenerateResult) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	var b strings.Builder
	b.WriteString("# 测试用例\n\n")
	b.WriteString(fmt.Sprintf("- 需求：%s\n", result.Graph.Doc.Title))
	b.WriteString(fmt.Sprintf("- 用例数：%d\n", len(result.Cases)))
	b.WriteString(fmt.Sprintf("- 校验通过：%v\n\n", result.Report.OK))

	if len(result.Report.Warnings) > 0 {
		b.WriteString("## 警告\n\n")
		for _, w := range result.Report.Warnings {
			b.WriteString("- " + w + "\n")
		}
		b.WriteString("\n")
	}

	for _, c := range result.Cases {
		b.WriteString(fmt.Sprintf("## %s %s\n\n", c.ID, c.Title))
		b.WriteString(fmt.Sprintf("- 模块：%s\n", c.Module))
		b.WriteString(fmt.Sprintf("- 功能：%s\n", c.Feature))
		b.WriteString(fmt.Sprintf("- 类型：%s\n", c.Type))
		b.WriteString(fmt.Sprintf("- 优先级：%s\n", c.Priority))
		b.WriteString(fmt.Sprintf("- 关联 AC：%s\n", c.ACID))
		if len(c.Tags) > 0 {
			b.WriteString(fmt.Sprintf("- 标签：%s\n", strings.Join(c.Tags, ", ")))
		}
		b.WriteString("\n### 前置条件\n\n")
		for _, p := range c.Preconditions {
			b.WriteString("- " + p + "\n")
		}
		b.WriteString("\n### 步骤\n\n")
		for i, s := range c.Steps {
			b.WriteString(fmt.Sprintf("%d. **操作**：%s  \n   **期望**：%s\n", i+1, s.Action, s.Expected))
		}
		b.WriteString("\n### 总体期望\n\n")
		b.WriteString(c.Expected + "\n\n")
	}

	path := filepath.Join(outDir, "cases.md")
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
