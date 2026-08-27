package validate_test

import (
	"testing"

	"github.com/webnode/autoWriteTestCase/internal/model"
	"github.com/webnode/autoWriteTestCase/internal/validate"
)

func TestBasicValidator_Coverage(t *testing.T) {
	graph := model.RequirementGraph{
		Criteria: []model.AcceptanceCriteria{
			{ID: "AC-001", Description: "a"},
			{ID: "AC-002", Description: "b"},
		},
	}
	cases := []model.TestCase{
		{ID: "TC-001", Title: "t1", ACID: "AC-001", Expected: "ok", Steps: []model.TestStep{{Action: "do", Expected: "ok"}}},
	}
	report := validate.NewBasic().Validate(graph, cases)
	if len(report.Gaps) != 1 || report.Gaps[0].ACID != "AC-002" {
		t.Fatalf("expected gap AC-002, got %+v", report.Gaps)
	}
}
