package validate

import (
	"fmt"
	"strings"

	"github.com/webnode/autoWriteTestCase/internal/model"
)

// Validator checks generated cases and AC coverage.
type Validator interface {
	Validate(graph model.RequirementGraph, cases []model.TestCase) model.ValidationReport
}

// BasicValidator performs required-field and coverage checks.
type BasicValidator struct{}

func NewBasic() *BasicValidator {
	return &BasicValidator{}
}

func (v *BasicValidator) Validate(graph model.RequirementGraph, cases []model.TestCase) model.ValidationReport {
	report := model.ValidationReport{OK: true}
	covered := map[string]int{}

	for i, c := range cases {
		prefix := fmt.Sprintf("cases[%d](%s)", i, c.ID)
		if strings.TrimSpace(c.ID) == "" {
			report.Errors = append(report.Errors, prefix+": id is required")
		}
		if strings.TrimSpace(c.Title) == "" {
			report.Errors = append(report.Errors, prefix+": title is required")
		}
		if strings.TrimSpace(c.Expected) == "" && len(c.Steps) == 0 {
			report.Errors = append(report.Errors, prefix+": expected or steps required")
		}
		if strings.TrimSpace(c.ACID) == "" {
			report.Warnings = append(report.Warnings, prefix+": missing ac_id (traceability)")
		} else {
			covered[c.ACID]++
		}
		for j, step := range c.Steps {
			if strings.TrimSpace(step.Action) == "" {
				report.Errors = append(report.Errors, fmt.Sprintf("%s.steps[%d]: action required", prefix, j))
			}
		}
	}

	for _, ac := range graph.Criteria {
		if covered[ac.ID] == 0 {
			report.Gaps = append(report.Gaps, model.CoverageGap{
				ACID:    ac.ID,
				Message: fmt.Sprintf("acceptance criteria %s has no linked test case", ac.ID),
			})
			report.Warnings = append(report.Warnings, fmt.Sprintf("coverage gap: %s", ac.ID))
		}
	}

	if len(report.Errors) > 0 || len(report.Gaps) > 0 {
		report.OK = len(report.Errors) == 0
	}
	return report
}
