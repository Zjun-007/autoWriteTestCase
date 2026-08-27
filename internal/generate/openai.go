package generate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/webnode/autoWriteTestCase/internal/config"
	"github.com/webnode/autoWriteTestCase/internal/model"
)

// OpenAIGenerator calls an OpenAI-compatible chat completions API.
type OpenAIGenerator struct {
	cfg       config.LLMConfig
	promptDir string
	version   string
	client    *http.Client
	fallback  Generator
}

func NewOpenAI(cfg config.LLMConfig, promptDir, version string) *OpenAIGenerator {
	return &OpenAIGenerator{
		cfg:       cfg,
		promptDir: promptDir,
		version:   version,
		client: &http.Client{
			Timeout: time.Duration(cfg.TimeoutSec) * time.Second,
		},
		fallback: NewMock(),
	}
}

type chatRequest struct {
	Model          string        `json:"model"`
	Temperature    float64       `json:"temperature"`
	Messages       []chatMessage `json:"messages"`
	ResponseFormat *respFormat   `json:"response_format,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type respFormat struct {
	Type string `json:"type"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type llmCasePayload struct {
	Cases []model.TestCase `json:"cases"`
}

func (g *OpenAIGenerator) Generate(ctx context.Context, graph model.RequirementGraph, scenarios []model.TestScenario) ([]model.TestCase, error) {
	if g.cfg.APIKey == "" {
		return nil, fmt.Errorf("llm api key is empty; set AWTC_LLM_API_KEY / OPENAI_API_KEY or use provider=mock")
	}

	tmpl, err := LoadPrompt(g.promptDir, g.version)
	if err != nil {
		return nil, err
	}

	acJSON, _ := json.MarshalIndent(graph.Criteria, "", "  ")
	scnJSON, _ := json.MarshalIndent(scenarios, "", "  ")
	user := RenderPrompt(tmpl, map[string]string{
		"title":     graph.Doc.Title,
		"criteria":  string(acJSON),
		"scenarios": string(scnJSON),
		"markdown":  trimMarkdown(graph.Doc.RawMarkdown, 8000),
	})

	reqBody := chatRequest{
		Model:       g.cfg.Model,
		Temperature: g.cfg.Temperature,
		Messages: []chatMessage{
			{Role: "system", Content: "You are a senior QA engineer. Output valid JSON only."},
			{Role: "user", Content: user},
		},
		ResponseFormat: &respFormat{Type: "json_object"},
	}

	var lastErr error
	attempts := g.cfg.MaxRetries + 1
	for i := 0; i < attempts; i++ {
		cases, err := g.doRequest(ctx, reqBody)
		if err == nil {
			return normalizeCases(cases, graph, scenarios), nil
		}
		lastErr = err
	}

	// Fall back to mock so CLI remains usable when LLM fails.
	mockCases, mockErr := g.fallback.Generate(ctx, graph, scenarios)
	if mockErr != nil {
		return nil, fmt.Errorf("llm failed: %v; mock fallback failed: %w", lastErr, mockErr)
	}
	return mockCases, nil
}

func (g *OpenAIGenerator) doRequest(ctx context.Context, body chatRequest) ([]model.TestCase, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	url := strings.TrimRight(g.cfg.BaseURL, "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+g.cfg.APIKey)

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("llm http %d: %s", resp.StatusCode, truncate(string(raw), 500))
	}

	var cr chatResponse
	if err := json.Unmarshal(raw, &cr); err != nil {
		return nil, fmt.Errorf("decode llm response: %w", err)
	}
	if cr.Error != nil {
		return nil, fmt.Errorf("llm error: %s", cr.Error.Message)
	}
	if len(cr.Choices) == 0 {
		return nil, fmt.Errorf("llm returned no choices")
	}
	content := strings.TrimSpace(cr.Choices[0].Message.Content)
	content = stripCodeFence(content)

	var payloadOut llmCasePayload
	if err := json.Unmarshal([]byte(content), &payloadOut); err != nil {
		// Try bare array
		var arr []model.TestCase
		if err2 := json.Unmarshal([]byte(content), &arr); err2 != nil {
			return nil, fmt.Errorf("parse cases json: %w", err)
		}
		return arr, nil
	}
	return payloadOut.Cases, nil
}

func normalizeCases(cases []model.TestCase, graph model.RequirementGraph, scenarios []model.TestScenario) []model.TestCase {
	scnByID := map[string]model.TestScenario{}
	for _, s := range scenarios {
		scnByID[s.ID] = s
	}
	out := make([]model.TestCase, 0, len(cases))
	for i, c := range cases {
		if c.ID == "" {
			c.ID = fmt.Sprintf("TC-%03d", i+1)
		}
		if c.RequirementID == "" {
			c.RequirementID = graph.Doc.Title
		}
		if c.Priority == "" {
			c.Priority = "P1"
		}
		if c.Type == "" {
			c.Type = model.CaseFunctional
		}
		if c.ACID == "" && c.ScenarioID != "" {
			if s, ok := scnByID[c.ScenarioID]; ok {
				c.ACID = s.ACID
				c.Module = s.Module
				c.Feature = s.Feature
			}
		}
		if len(c.Steps) == 0 {
			c.Steps = []model.TestStep{{Action: "执行测试", Expected: c.Expected}}
		}
		out = append(out, c)
	}
	return out
}

func stripCodeFence(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimPrefix(s, "```JSON")
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSuffix(s, "```")
		s = strings.TrimSpace(s)
	}
	return s
}

func trimMarkdown(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "\n…(truncated)"
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
