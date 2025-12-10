package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration
type Config struct {
	LLM      LLMConfig      `yaml:"llm"`
	Database DatabaseConfig `yaml:"database"`
	Agent    AgentConfig    `yaml:"agent"`
}

type LLMConfig struct {
	APIKey  string `yaml:"api_key"`
	BaseURL string `yaml:"base_url"`
	Model   string `yaml:"model"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type AgentConfig struct {
	MaxHistory int `yaml:"max_history"` // Number of messages to keep in context
}

// LoadConfig loads configuration from file or environment variables
func LoadConfig(path string) (*Config, error) {
	// Default configuration
	cfg := &Config{
		LLM: LLMConfig{
			BaseURL: "https://api.deepseek.com/v1",
			Model:   "deepseek-chat",
		},
		Database: DatabaseConfig{
			Path: "gomentum.db",
		},
		Agent: AgentConfig{
			MaxHistory: 20,
		},
	}

	// Try to load from file
	f, err := os.Open(path)
	if err == nil {
		defer f.Close()
		decoder := yaml.NewDecoder(f)
		if err := decoder.Decode(cfg); err != nil {
			return nil, fmt.Errorf("failed to decode config file: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}

	// Override with environment variables if set
	if apiKey := os.Getenv("LLM_API_KEY"); apiKey != "" {
		cfg.LLM.APIKey = apiKey
	}
	if baseURL := os.Getenv("LLM_BASE_URL"); baseURL != "" {
		cfg.LLM.BaseURL = baseURL
	}
	if model := os.Getenv("LLM_MODEL"); model != "" {
		cfg.LLM.Model = model
	}

	// Validate
	if cfg.LLM.APIKey == "" {
		return nil, fmt.Errorf("LLM API Key is missing. Please set LLM_API_KEY env var or configure it in %s", path)
	}

	return cfg, nil
}

// SaveConfig saves the configuration to a file
func SaveConfig(path string, cfg *Config) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer f.Close()

	encoder := yaml.NewEncoder(f)
	encoder.SetIndent(2)
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}
	return nil
}
