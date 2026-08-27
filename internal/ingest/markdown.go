package ingest

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/webnode/autoWriteTestCase/internal/model"
)

// Ingester loads a requirement source into a RequirementDoc.
type Ingester interface {
	Ingest(path string) (model.RequirementDoc, error)
}

// MarkdownIngester reads local Markdown / plain text files.
type MarkdownIngester struct{}

func NewMarkdown() *MarkdownIngester {
	return &MarkdownIngester{}
}

func (m *MarkdownIngester) Ingest(path string) (model.RequirementDoc, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return model.RequirementDoc{}, fmt.Errorf("read requirement: %w", err)
	}
	raw := string(data)
	title := deriveTitle(raw, path)
	return model.RequirementDoc{
		SourcePath:  path,
		Title:       title,
		RawMarkdown: raw,
	}, nil
}

func deriveTitle(raw, path string) string {
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	base := filepath.Base(path)
	return strings.TrimSuffix(base, filepath.Ext(base))
}
