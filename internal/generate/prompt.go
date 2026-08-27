package generate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LoadPrompt reads prompts/<version>/generate_cases.md.
func LoadPrompt(promptDir, version string) (string, error) {
	path := filepath.Join(promptDir, version, "generate_cases.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("load prompt %s: %w", path, err)
	}
	return string(data), nil
}

// RenderPrompt replaces simple {{placeholders}}.
func RenderPrompt(tmpl string, vars map[string]string) string {
	out := tmpl
	for k, v := range vars {
		out = strings.ReplaceAll(out, "{{"+k+"}}", v)
	}
	return out
}
