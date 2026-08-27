package analyze

import (
	"fmt"
	"strings"

	"github.com/webnode/autoWriteTestCase/internal/model"
)

// Analyzer derives test scenarios from a requirement graph.
type Analyzer interface {
	Analyze(graph model.RequirementGraph) ([]model.TestScenario, error)
}

// RuleAnalyzer creates normal / boundary / abnormal scenarios per AC.
type RuleAnalyzer struct{}

func NewRule() *RuleAnalyzer {
	return &RuleAnalyzer{}
}

func (a *RuleAnalyzer) Analyze(graph model.RequirementGraph) ([]model.TestScenario, error) {
	var out []model.TestScenario
	idx := 0
	for _, ac := range graph.Criteria {
		templates := []struct {
			typ    model.ScenarioType
			title  string
			intent string
		}{
			{
				typ:    model.ScenarioNormal,
				title:  fmt.Sprintf("%s - 正常路径", short(ac.Description, 40)),
				intent: fmt.Sprintf("验证验收点在合法输入与正常流程下成立：%s", ac.Description),
			},
			{
				typ:    model.ScenarioBoundary,
				title:  fmt.Sprintf("%s - 边界条件", short(ac.Description, 40)),
				intent: fmt.Sprintf("针对边界值/临界条件验证：%s", ac.Description),
			},
			{
				typ:    model.ScenarioAbnormal,
				title:  fmt.Sprintf("%s - 异常路径", short(ac.Description, 40)),
				intent: fmt.Sprintf("验证非法输入、权限不足或失败恢复：%s", ac.Description),
			},
		}
		for _, t := range templates {
			idx++
			out = append(out, model.TestScenario{
				ID:      fmt.Sprintf("SCN-%03d", idx),
				ACID:    ac.ID,
				Type:    t.typ,
				Title:   t.title,
				Intent:  t.intent,
				Module:  ac.Module,
				Feature: ac.Feature,
			})
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no scenarios derived")
	}
	return out, nil
}

func short(s string, n int) string {
	s = strings.TrimSpace(s)
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
