package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// ProviderType represents the type of AI provider configured
type ProviderType string

const (
	// ProviderLocalOllama uses a local Ollama instance
	ProviderLocalOllama ProviderType = "local_ollama"
	// ProviderCloudFreeTier uses a cloud LLM API (OpenAI compatible)
	ProviderCloudFreeTier ProviderType = "cloud_free_tier"
	// ProviderDisabled disables AI parsing functionality
	ProviderDisabled ProviderType = "disabled"
)

// Config holds all configuration for the AI parsing service
type Config struct {
	Provider ProviderType
	Ollama   OllamaConfig
	Cloud    CloudConfig
}

// OllamaConfig holds configuration for local Ollama provider
type OllamaConfig struct {
	BaseURL   string
	ModelName string
	Timeout   time.Duration
}

// CloudConfig holds configuration for cloud LLM provider
type CloudConfig struct {
	BaseURL   string
	ModelName string
	APIKey    string
	Timeout   time.Duration
	MaxRPS    int // Maximum requests per second for rate limiting
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	providerStr := getEnv("AI_INGREDIENT_PROVIDER", "disabled")
	provider := ProviderType(providerStr)

	// Validate provider type
	if provider != ProviderLocalOllama && provider != ProviderCloudFreeTier && provider != ProviderDisabled {
		return nil, fmt.Errorf("invalid AI_INGREDIENT_PROVIDER: %s (must be: local_ollama, cloud_free_tier, or disabled)", providerStr)
	}

	cfg := &Config{
		Provider: provider,
		Ollama:   loadOllamaConfig(),
		Cloud:    loadCloudConfig(),
	}

	// Validate configuration based on provider type
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// loadOllamaConfig loads Ollama-specific configuration
func loadOllamaConfig() OllamaConfig {
	timeoutMs := getEnvInt("OLLAMA_TIMEOUT_MS", 10000)

	return OllamaConfig{
		BaseURL:   getEnv("OLLAMA_BASE_URL", "http://ollama:11434"),
		ModelName: getEnv("OLLAMA_MODEL_NAME", "llama3"),
		Timeout:   time.Duration(timeoutMs) * time.Millisecond,
	}
}

// loadCloudConfig loads cloud provider configuration
func loadCloudConfig() CloudConfig {
	timeoutMs := getEnvInt("CLOUD_AI_TIMEOUT_MS", 10000)
	maxRPS := getEnvInt("CLOUD_AI_MAX_RPS", 1)

	return CloudConfig{
		BaseURL:   getEnv("CLOUD_AI_BASE_URL", ""),
		ModelName: getEnv("CLOUD_AI_MODEL_NAME", "gpt-4o-mini"),
		APIKey:    getEnv("CLOUD_AI_API_KEY", ""),
		Timeout:   time.Duration(timeoutMs) * time.Millisecond,
		MaxRPS:    maxRPS,
	}
}

// Validate validates the configuration based on the selected provider
func (c *Config) Validate() error {
	switch c.Provider {
	case ProviderLocalOllama:
		return c.validateOllamaConfig()
	case ProviderCloudFreeTier:
		return c.validateCloudConfig()
	case ProviderDisabled:
		return nil // No validation needed for disabled mode
	default:
		return fmt.Errorf("unknown provider type: %s", c.Provider)
	}
}

// validateOllamaConfig validates Ollama-specific configuration
func (c *Config) validateOllamaConfig() error {
	if c.Ollama.BaseURL == "" {
		return fmt.Errorf("OLLAMA_BASE_URL is required when provider is local_ollama")
	}
	if c.Ollama.ModelName == "" {
		return fmt.Errorf("OLLAMA_MODEL_NAME is required when provider is local_ollama")
	}
	if c.Ollama.Timeout <= 0 {
		return fmt.Errorf("OLLAMA_TIMEOUT_MS must be positive")
	}
	return nil
}

// validateCloudConfig validates cloud provider configuration
func (c *Config) validateCloudConfig() error {
	if c.Cloud.BaseURL == "" {
		return fmt.Errorf("CLOUD_AI_BASE_URL is required when provider is cloud_free_tier")
	}
	if c.Cloud.ModelName == "" {
		return fmt.Errorf("CLOUD_AI_MODEL_NAME is required when provider is cloud_free_tier")
	}
	if c.Cloud.APIKey == "" {
		return fmt.Errorf("CLOUD_AI_API_KEY is required when provider is cloud_free_tier")
	}
	if c.Cloud.Timeout <= 0 {
		return fmt.Errorf("CLOUD_AI_TIMEOUT_MS must be positive")
	}
	if c.Cloud.MaxRPS <= 0 {
		return fmt.Errorf("CLOUD_AI_MAX_RPS must be positive")
	}
	return nil
}

// IsEnabled returns true if AI parsing is enabled (not disabled)
func (c *Config) IsEnabled() bool {
	return c.Provider != ProviderDisabled
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt retrieves an integer environment variable or returns a default value
func getEnvInt(key string, defaultValue int) int {
	if valueStr := os.Getenv(key); valueStr != "" {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}
