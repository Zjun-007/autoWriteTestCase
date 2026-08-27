// Package model defines domain types shared across the generate pipeline.
package model

// RequirementDoc is the normalized requirement after ingest.
type RequirementDoc struct {
	SourcePath string `json:"source_path"`
	Title      string `json:"title"`
	RawMarkdown string `json:"raw_markdown"`
}

// AcceptanceCriteria is a single verifiable acceptance point.
type AcceptanceCriteria struct {
	ID          string `json:"id"`
	Module      string `json:"module"`
	Feature     string `json:"feature"`
	Description string `json:"description"`
}

// RequirementGraph is the parsed structure used by analyze/generate.
type RequirementGraph struct {
	Doc        RequirementDoc         `json:"doc"`
	Modules    []string               `json:"modules"`
	Criteria   []AcceptanceCriteria   `json:"criteria"`
}

// ScenarioType classifies a test scenario intent.
type ScenarioType string

const (
	ScenarioNormal   ScenarioType = "normal"
	ScenarioBoundary ScenarioType = "boundary"
	ScenarioAbnormal ScenarioType = "abnormal"
)

// TestScenario is a scenario intent before concrete steps are written.
type TestScenario struct {
	ID          string       `json:"id"`
	ACID        string       `json:"ac_id"`
	Type        ScenarioType `json:"type"`
	Title       string       `json:"title"`
	Intent      string       `json:"intent"`
	Module      string       `json:"module"`
	Feature     string       `json:"feature"`
}

// CaseType classifies a concrete test case.
type CaseType string

const (
	CaseFunctional CaseType = "functional"
	CaseBoundary   CaseType = "boundary"
	CaseAbnormal   CaseType = "abnormal"
)

// TestStep is one action/expectation pair inside a case.
type TestStep struct {
	Action   string `json:"action"`
	Expected string `json:"expected"`
}

// TestCase is the structured output unit.
type TestCase struct {
	ID            string     `json:"id"`
	Title         string     `json:"title"`
	Preconditions []string   `json:"preconditions"`
	Steps         []TestStep `json:"steps"`
	Expected      string     `json:"expected"`
	Priority      string     `json:"priority"`
	Type          CaseType   `json:"type"`
	Tags          []string   `json:"tags"`
	RequirementID string     `json:"requirement_id"`
	ACID          string     `json:"ac_id"`
	ScenarioID    string     `json:"scenario_id,omitempty"`
	Module        string     `json:"module,omitempty"`
	Feature       string     `json:"feature,omitempty"`
}

// CoverageGap reports an AC without enough cases.
type CoverageGap struct {
	ACID    string `json:"ac_id"`
	Message string `json:"message"`
}

// ValidationReport summarizes validation results.
type ValidationReport struct {
	OK       bool           `json:"ok"`
	Errors   []string       `json:"errors,omitempty"`
	Warnings []string       `json:"warnings,omitempty"`
	Gaps     []CoverageGap  `json:"gaps,omitempty"`
}

// GenerateResult is the pipeline output bundle.
type GenerateResult struct {
	Graph     RequirementGraph  `json:"graph"`
	Scenarios []TestScenario    `json:"scenarios"`
	Cases     []TestCase        `json:"cases"`
	Report    ValidationReport  `json:"report"`
}
