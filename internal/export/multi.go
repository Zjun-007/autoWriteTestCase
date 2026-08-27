package export

import (
	"fmt"
	"strings"
)

// Multi resolves exporters by format name.
type Multi struct {
	exporters map[string]Exporter
}

func NewMulti(list ...Exporter) *Multi {
	m := &Multi{exporters: map[string]Exporter{}}
	for _, e := range list {
		m.exporters[strings.ToLower(e.Name())] = e
	}
	return m
}

func DefaultMulti() *Multi {
	return NewMulti(NewJSON(), NewExcel(), NewMarkdown())
}

func (m *Multi) Resolve(formats []string) ([]Exporter, error) {
	var out []Exporter
	seen := map[string]bool{}
	for _, f := range formats {
		name := strings.ToLower(strings.TrimSpace(f))
		if name == "" || seen[name] {
			continue
		}
		e, ok := m.exporters[name]
		if !ok {
			return nil, fmt.Errorf("unknown export format %q (supported: json, excel, markdown)", f)
		}
		seen[name] = true
		out = append(out, e)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no export formats specified")
	}
	return out, nil
}
