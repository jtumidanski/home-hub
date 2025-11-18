package parser

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// ProviderType represents the type of AI provider
type ProviderType string

const (
	ProviderLocalOllama   ProviderType = "local_ollama"
	ProviderCloudFreeTier ProviderType = "cloud_free_tier"
	ProviderDisabled      ProviderType = "disabled"
)

// Config holds configuration for creating a parser
type Config struct {
	Provider ProviderType
	Ollama   OllamaConfig
	Cloud    CloudConfig
}

// OllamaConfig holds Ollama-specific configuration
type OllamaConfig struct {
	BaseURL   string
	ModelName string
	Timeout   time.Duration
}

// CloudConfig holds cloud provider configuration
type CloudConfig struct {
	BaseURL   string
	ModelName string
	APIKey    string
	Timeout   time.Duration
}

// NewParser creates an IngredientParser based on the provided configuration
func NewParser(cfg Config, logger logrus.FieldLogger) (IngredientParser, error) {
	switch cfg.Provider {
	case ProviderLocalOllama:
		return NewOllamaParser(
			cfg.Ollama.BaseURL,
			cfg.Ollama.ModelName,
			cfg.Ollama.Timeout,
			logger,
		), nil

	case ProviderCloudFreeTier:
		return NewCloudParser(
			cfg.Cloud.BaseURL,
			cfg.Cloud.ModelName,
			cfg.Cloud.APIKey,
			cfg.Cloud.Timeout,
			logger,
		), nil

	case ProviderDisabled:
		return nil, fmt.Errorf("AI provider is disabled")

	default:
		return nil, fmt.Errorf("unknown provider type: %s", cfg.Provider)
	}
}
