package parse

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/webnode/autoWriteTestCase/internal/model"
)

// Parser turns a RequirementDoc into a RequirementGraph.
type Parser interface {
	Parse(doc model.RequirementDoc) (model.RequirementGraph, error)
}

var (
	headingRE = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)
	listRE    = regexp.MustCompile(`^[-*+]\s+(.+)$`)
	orderedRE = regexp.MustCompile(`^\d+[.)]\s+(.+)$`)
)

// MarkdownParser splits by headings and collects acceptance criteria lists.
type MarkdownParser struct{}

func NewMarkdown() *MarkdownParser {
	return &MarkdownParser{}
}

func (p *MarkdownParser) Parse(doc model.RequirementDoc) (model.RequirementGraph, error) {
	graph := model.RequirementGraph{Doc: doc}
	lines := strings.Split(doc.RawMarkdown, "\n")

	var (
		module      = "默认模块"
		feature     = "默认功能"
		inACSection bool
		modules     []string
		seenModule  = map[string]bool{}
		acIdx       int
	)

	flushFeature := func() {}
	_ = flushFeature

	for _, raw := range lines {
		line := strings.TrimRight(raw, "\r")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if m := headingRE.FindStringSubmatch(trimmed); m != nil {
			level := len(m[1])
			text := strings.TrimSpace(m[2])
			lower := strings.ToLower(text)

			inACSection = isACHeading(lower)

			switch {
			case level <= 2:
				module = text
				feature = text
				if !seenModule[module] {
					seenModule[module] = true
					modules = append(modules, module)
				}
			case level == 3:
				feature = text
			default:
				if inACSection {
					feature = text
				}
			}
			continue
		}

		item := ""
		if m := listRE.FindStringSubmatch(trimmed); m != nil {
			item = strings.TrimSpace(m[1])
		} else if m := orderedRE.FindStringSubmatch(trimmed); m != nil {
			item = strings.TrimSpace(m[1])
		}

		if item == "" {
			continue
		}

		// Collect list items under AC section, or any list that looks like a criterion.
		if inACSection || looksLikeCriterion(item) {
			acIdx++
			graph.Criteria = append(graph.Criteria, model.AcceptanceCriteria{
				ID:          fmt.Sprintf("AC-%03d", acIdx),
				Module:      module,
				Feature:     feature,
				Description: stripACPrefix(item),
			})
		}
	}

	// Fallback: if no AC found, treat non-heading paragraphs / bullets as criteria.
	if len(graph.Criteria) == 0 {
		acIdx = 0
		module = doc.Title
		feature = doc.Title
		for _, raw := range lines {
			trimmed := strings.TrimSpace(raw)
			if trimmed == "" || headingRE.MatchString(trimmed) {
				continue
			}
			item := trimmed
			if m := listRE.FindStringSubmatch(trimmed); m != nil {
				item = strings.TrimSpace(m[1])
			} else if m := orderedRE.FindStringSubmatch(trimmed); m != nil {
				item = strings.TrimSpace(m[1])
			}
			if len([]rune(item)) < 4 {
				continue
			}
			acIdx++
			graph.Criteria = append(graph.Criteria, model.AcceptanceCriteria{
				ID:          fmt.Sprintf("AC-%03d", acIdx),
				Module:      module,
				Feature:     feature,
				Description: item,
			})
		}
	}

	if len(graph.Criteria) == 0 {
		return graph, fmt.Errorf("no acceptance criteria found in %s", doc.SourcePath)
	}

	graph.Modules = modules
	if len(graph.Modules) == 0 {
		graph.Modules = []string{doc.Title}
	}
	return graph, nil
}

func isACHeading(lower string) bool {
	keys := []string{
		"验收", "验收标准", "验收条件", "acceptance", "ac", "given", "需求点", "功能点",
	}
	for _, k := range keys {
		if strings.Contains(lower, k) {
			return true
		}
	}
	return false
}

func looksLikeCriterion(item string) bool {
	lower := strings.ToLower(item)
	prefixes := []string{"应", "必须", "支持", "可以", "能够", "用户", "系统", "当", "如果", "shall", "should", "must", "given", "when"}
	for _, p := range prefixes {
		if strings.HasPrefix(lower, p) || strings.HasPrefix(item, p) {
			return true
		}
	}
	return false
}

func stripACPrefix(item string) string {
	item = strings.TrimSpace(item)
	item = regexp.MustCompile(`^(?i)(ac[-_]?\d+[:：.\s]+|验收[标准条件]?[:：]\s*)`).ReplaceAllString(item, "")
	return strings.TrimSpace(item)
}
