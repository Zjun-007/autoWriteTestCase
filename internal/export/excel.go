package export

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
	"github.com/webnode/autoWriteTestCase/internal/model"
)

// ExcelExporter writes cases.xlsx.
type ExcelExporter struct{}

func NewExcel() *ExcelExporter { return &ExcelExporter{} }

func (e *ExcelExporter) Name() string { return "excel" }

func (e *ExcelExporter) Export(outDir string, result model.GenerateResult) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	f := excelize.NewFile()
	sheet := "Cases"
	idx, err := f.NewSheet(sheet)
	if err != nil {
		return err
	}
	_ = f.DeleteSheet("Sheet1")
	f.SetActiveSheet(idx)

	headers := []string{
		"ID", "Title", "Module", "Feature", "Type", "Priority",
		"AC_ID", "Requirement", "Preconditions", "Steps", "Expected", "Tags",
	}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}

	for r, c := range result.Cases {
		row := r + 2
		steps := formatSteps(c.Steps)
		vals := []any{
			c.ID, c.Title, c.Module, c.Feature, string(c.Type), c.Priority,
			c.ACID, c.RequirementID, strings.Join(c.Preconditions, "; "),
			steps, c.Expected, strings.Join(c.Tags, ", "),
		}
		for i, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(i+1, row)
			_ = f.SetCellValue(sheet, cell, v)
		}
	}

	path := filepath.Join(outDir, "cases.xlsx")
	if err := f.SaveAs(path); err != nil {
		return fmt.Errorf("save excel: %w", err)
	}
	return nil
}

func formatSteps(steps []model.TestStep) string {
	parts := make([]string, 0, len(steps))
	for i, s := range steps {
		parts = append(parts, fmt.Sprintf("%d) %s => %s", i+1, s.Action, s.Expected))
	}
	return strings.Join(parts, " | ")
}
