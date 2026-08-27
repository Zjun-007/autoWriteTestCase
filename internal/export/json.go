package export

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/webnode/autoWriteTestCase/internal/model"
)

// Exporter writes generate results to a sink.
type Exporter interface {
	Export(outDir string, result model.GenerateResult) error
	Name() string
}

// JSONExporter writes cases.json and result.json.
type JSONExporter struct{}

func NewJSON() *JSONExporter { return &JSONExporter{} }

func (e *JSONExporter) Name() string { return "json" }

func (e *JSONExporter) Export(outDir string, result model.GenerateResult) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(outDir, "cases.json"), result.Cases); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(outDir, "result.json"), result); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(outDir, "scenarios.json"), result.Scenarios); err != nil {
		return err
	}
	return writeJSON(filepath.Join(outDir, "criteria.json"), result.Graph.Criteria)
}

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
