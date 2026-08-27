package pipeline

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/webnode/autoWriteTestCase/internal/analyze"
	"github.com/webnode/autoWriteTestCase/internal/config"
	"github.com/webnode/autoWriteTestCase/internal/export"
	"github.com/webnode/autoWriteTestCase/internal/generate"
	"github.com/webnode/autoWriteTestCase/internal/ingest"
	"github.com/webnode/autoWriteTestCase/internal/model"
	"github.com/webnode/autoWriteTestCase/internal/parse"
	"github.com/webnode/autoWriteTestCase/internal/validate"
)

// Options controls a single generate run.
type Options struct {
	Input   string
	OutDir  string
	Formats []string
}

// Pipeline wires ingest → parse → analyze → generate → validate → export.
type Pipeline struct {
	cfg       config.Config
	log       *slog.Logger
	ingester  ingest.Ingester
	parser    parse.Parser
	analyzer  analyze.Analyzer
	generator generate.Generator
	validator validate.Validator
	exporters *export.Multi
}

func New(cfg config.Config, log *slog.Logger) (*Pipeline, error) {
	if log == nil {
		log = slog.Default()
	}
	gen, err := buildGenerator(cfg)
	if err != nil {
		return nil, err
	}
	return &Pipeline{
		cfg:       cfg,
		log:       log,
		ingester:  ingest.NewMarkdown(),
		parser:    parse.NewMarkdown(),
		analyzer:  analyze.NewRule(),
		generator: gen,
		validator: validate.NewBasic(),
		exporters: export.DefaultMulti(),
	}, nil
}

func buildGenerator(cfg config.Config) (generate.Generator, error) {
	switch strings.ToLower(cfg.LLM.Provider) {
	case "mock", "":
		return generate.NewMock(), nil
	case "openai":
		return generate.NewOpenAI(cfg.LLM, cfg.Paths.PromptDir, cfg.Generate.PromptVersion), nil
	default:
		return nil, fmt.Errorf("unsupported llm provider %q (use mock or openai)", cfg.LLM.Provider)
	}
}

// Run executes the full pipeline.
func (p *Pipeline) Run(ctx context.Context, opt Options) (model.GenerateResult, error) {
	var result model.GenerateResult

	p.log.Info("ingest", "input", opt.Input)
	doc, err := p.ingester.Ingest(opt.Input)
	if err != nil {
		return result, err
	}

	p.log.Info("parse", "title", doc.Title)
	graph, err := p.parser.Parse(doc)
	if err != nil {
		return result, err
	}
	result.Graph = graph
	p.log.Info("parsed criteria", "count", len(graph.Criteria))

	p.log.Info("analyze")
	scenarios, err := p.analyzer.Analyze(graph)
	if err != nil {
		return result, err
	}
	result.Scenarios = scenarios
	p.log.Info("scenarios", "count", len(scenarios))

	p.log.Info("generate", "provider", p.cfg.LLM.Provider)
	cases, err := p.generator.Generate(ctx, graph, scenarios)
	if err != nil {
		return result, err
	}
	result.Cases = cases
	p.log.Info("cases", "count", len(cases))

	report := p.validator.Validate(graph, cases)
	result.Report = report
	if len(report.Errors) > 0 {
		p.log.Warn("validation errors", "count", len(report.Errors))
	}
	if len(report.Warnings) > 0 {
		p.log.Warn("validation warnings", "count", len(report.Warnings))
	}

	formats := opt.Formats
	if len(formats) == 0 {
		formats = p.cfg.Generate.Formats
	}
	exporters, err := p.exporters.Resolve(formats)
	if err != nil {
		return result, err
	}
	for _, ex := range exporters {
		p.log.Info("export", "format", ex.Name(), "out", opt.OutDir)
		if err := ex.Export(opt.OutDir, result); err != nil {
			return result, fmt.Errorf("export %s: %w", ex.Name(), err)
		}
	}

	return result, nil
}
