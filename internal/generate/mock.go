package generate

import (
	"context"
	"fmt"

	"github.com/webnode/autoWriteTestCase/internal/model"
)

// Generator turns scenarios into concrete test cases.
type Generator interface {
	Generate(ctx context.Context, graph model.RequirementGraph, scenarios []model.TestScenario) ([]model.TestCase, error)
}

// MockGenerator builds deterministic cases without calling an LLM.
type MockGenerator struct{}

func NewMock() *MockGenerator {
	return &MockGenerator{}
}

func (g *MockGenerator) Generate(_ context.Context, graph model.RequirementGraph, scenarios []model.TestScenario) ([]model.TestCase, error) {
	acMap := map[string]model.AcceptanceCriteria{}
	for _, ac := range graph.Criteria {
		acMap[ac.ID] = ac
	}

	cases := make([]model.TestCase, 0, len(scenarios))
	for i, scn := range scenarios {
		ac := acMap[scn.ACID]
		tc := model.TestCase{
			ID:            fmt.Sprintf("TC-%03d", i+1),
			Title:         scn.Title,
			Preconditions: []string{"测试环境可用", "测试账号具备所需权限"},
			Steps:         stepsFor(scn, ac),
			Expected:      expectedFor(scn, ac),
			Priority:      priorityFor(scn.Type),
			Type:          caseTypeFor(scn.Type),
			Tags:          []string{string(scn.Type), scn.Module},
			RequirementID: graph.Doc.Title,
			ACID:          scn.ACID,
			ScenarioID:    scn.ID,
			Module:        scn.Module,
			Feature:       scn.Feature,
		}
		cases = append(cases, tc)
	}
	return cases, nil
}

func caseTypeFor(t model.ScenarioType) model.CaseType {
	switch t {
	case model.ScenarioBoundary:
		return model.CaseBoundary
	case model.ScenarioAbnormal:
		return model.CaseAbnormal
	default:
		return model.CaseFunctional
	}
}

func priorityFor(t model.ScenarioType) string {
	if t == model.ScenarioNormal {
		return "P0"
	}
	if t == model.ScenarioBoundary {
		return "P1"
	}
	return "P2"
}

func stepsFor(scn model.TestScenario, ac model.AcceptanceCriteria) []model.TestStep {
	desc := ac.Description
	switch scn.Type {
	case model.ScenarioBoundary:
		return []model.TestStep{
			{Action: "准备处于边界的输入/状态（最小值、最大值或临界配置）", Expected: "边界数据准备完成"},
			{Action: fmt.Sprintf("执行与验收点相关的操作：%s", desc), Expected: "系统接受边界值并按规则处理"},
			{Action: "检查关键结果与提示信息", Expected: "结果符合边界规则，无崩溃或错误数据"},
		}
	case model.ScenarioAbnormal:
		return []model.TestStep{
			{Action: "准备非法输入、缺失依赖或无权限账号", Expected: "异常前置条件就绪"},
			{Action: fmt.Sprintf("尝试执行：%s", desc), Expected: "操作被拒绝或进入失败分支"},
			{Action: "检查错误提示、日志与数据一致性", Expected: "有明确错误提示，核心数据未被破坏"},
		}
	default:
		return []model.TestStep{
			{Action: "使用合法数据与正常账号进入功能入口", Expected: "页面/接口可访问"},
			{Action: fmt.Sprintf("按需求完成操作以验证：%s", desc), Expected: "关键执行成功"},
			{Action: "核对界面展示与后端数据", Expected: fmt.Sprintf("满足验收点：%s", desc)},
		}
	}
}

func expectedFor(scn model.TestScenario, ac model.AcceptanceCriteria) string {
	switch scn.Type {
	case model.ScenarioBoundary:
		return fmt.Sprintf("边界条件下系统行为符合规则，验收点仍成立或按文档降级：%s", ac.Description)
	case model.ScenarioAbnormal:
		return fmt.Sprintf("异常被正确拦截并提示，不影响系统稳定性；相关验收点：%s", ac.Description)
	default:
		return fmt.Sprintf("正常流程下验收点通过：%s", ac.Description)
	}
}
