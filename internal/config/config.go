package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds runtime settings for awtc.
type Config struct {
	LLM      LLMConfig      `yaml:"llm"`
	Generate GenerateConfig `yaml:"generate"`
	Paths    PathsConfig    `yaml:"paths"`
}

type LLMConfig struct {
	Provider    string  `yaml:"provider"` // mock | openai
	BaseURL     string  `yaml:"base_url"`
	APIKey      string  `yaml:"api_key"`
	Model       string  `yaml:"model"`
	Temperature float64 `yaml:"temperature"`
	TimeoutSec  int     `yaml:"timeout_sec"`
	MaxRetries  int     `yaml:"max_retries"`
}

type GenerateConfig struct {
	PromptVersion string   `yaml:"prompt_version"`
	Formats       []string `yaml:"formats"`
}

type PathsConfig struct {
	PromptDir string `yaml:"prompt_dir"`
}

// Default returns sensible MVP defaults (mock generator, no API key required).
func Default() Config {
	return Config{
		LLM: LLMConfig{
			Provider:    "mock",
			BaseURL:     "https://api.openai.com/v1",
			Model:       "gpt-4o-mini",
			Temperature: 0.2,
			TimeoutSec:  60,
			MaxRetries:  2,
		},
		Generate: GenerateConfig{
			PromptVersion: "v1",
			Formats:       []string{"json", "excel", "markdown"},
		},
		Paths: PathsConfig{
			PromptDir: "prompts",
		},
	}
}

// Load reads YAML config and overlays environment variables.
// Missing file falls back to Default() without error.
func Load(path string) (Config, error) {
	cfg := Default()
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				overlayEnv(&cfg)
				normalize(&cfg)
				return cfg, nil
			}
			return cfg, fmt.Errorf("read config: %w", err)
		}
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return cfg, fmt.Errorf("parse config: %w", err)
		}
	}
	overlayEnv(&cfg)
	normalize(&cfg)
	return cfg, nil
}

func overlayEnv(cfg *Config) {
	if v := os.Getenv("AWTC_LLM_PROVIDER"); v != "" {
		cfg.LLM.Provider = v
	}
	if v := os.Getenv("AWTC_LLM_BASE_URL"); v != "" {
		cfg.LLM.BaseURL = v
	}
	if v := os.Getenv("AWTC_LLM_API_KEY"); v != "" {
		cfg.LLM.APIKey = v
	} else if v := os.Getenv("OPENAI_API_KEY"); v != "" {
		cfg.LLM.APIKey = v
	}
	if v := os.Getenv("AWTC_LLM_MODEL"); v != "" {
		cfg.LLM.Model = v
	}
	if v := os.Getenv("AWTC_PROMPT_DIR"); v != "" {
		cfg.Paths.PromptDir = v
	}
}

func normalize(cfg *Config) {
	cfg.LLM.Provider = strings.ToLower(strings.TrimSpace(cfg.LLM.Provider))
	if cfg.LLM.Provider == "" {
		cfg.LLM.Provider = "mock"
	}
	if cfg.LLM.TimeoutSec <= 0 {
		cfg.LLM.TimeoutSec = 60
	}
	if cfg.LLM.MaxRetries < 0 {
		cfg.LLM.MaxRetries = 0
	}
	if cfg.Generate.PromptVersion == "" {
		cfg.Generate.PromptVersion = "v1"
	}
	if cfg.Paths.PromptDir == "" {
		cfg.Paths.PromptDir = "prompts"
	}
	if len(cfg.Generate.Formats) == 0 {
		cfg.Generate.Formats = []string{"json", "excel", "markdown"}
	}
	// Auto-upgrade to openai when key present and provider still mock.
	if cfg.LLM.Provider == "mock" && cfg.LLM.APIKey != "" {
		cfg.LLM.Provider = "openai"
	}
}
