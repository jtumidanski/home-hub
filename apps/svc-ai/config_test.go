package main

import (
	"os"
	"testing"
)

func TestLoadConfig_Disabled(t *testing.T) {
	// Setup
	os.Setenv("AI_INGREDIENT_PROVIDER", "disabled")
	defer os.Unsetenv("AI_INGREDIENT_PROVIDER")

	// Execute
	cfg, err := LoadConfig()

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.Provider != ProviderDisabled {
		t.Errorf("expected provider to be disabled, got %s", cfg.Provider)
	}
	if cfg.IsEnabled() {
		t.Error("expected IsEnabled() to return false for disabled provider")
	}
}

func TestLoadConfig_LocalOllama_Valid(t *testing.T) {
	// Setup
	os.Setenv("AI_INGREDIENT_PROVIDER", "local_ollama")
	os.Setenv("OLLAMA_BASE_URL", "http://localhost:11434")
	os.Setenv("OLLAMA_MODEL_NAME", "llama3")
	os.Setenv("OLLAMA_TIMEOUT_MS", "5000")
	defer func() {
		os.Unsetenv("AI_INGREDIENT_PROVIDER")
		os.Unsetenv("OLLAMA_BASE_URL")
		os.Unsetenv("OLLAMA_MODEL_NAME")
		os.Unsetenv("OLLAMA_TIMEOUT_MS")
	}()

	// Execute
	cfg, err := LoadConfig()

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.Provider != ProviderLocalOllama {
		t.Errorf("expected provider to be local_ollama, got %s", cfg.Provider)
	}
	if cfg.Ollama.BaseURL != "http://localhost:11434" {
		t.Errorf("expected BaseURL to be http://localhost:11434, got %s", cfg.Ollama.BaseURL)
	}
	if cfg.Ollama.ModelName != "llama3" {
		t.Errorf("expected ModelName to be llama3, got %s", cfg.Ollama.ModelName)
	}
	if !cfg.IsEnabled() {
		t.Error("expected IsEnabled() to return true for local_ollama provider")
	}
}

func TestLoadConfig_LocalOllama_MissingBaseURL(t *testing.T) {
	// Setup
	os.Setenv("AI_INGREDIENT_PROVIDER", "local_ollama")
	// Don't set OLLAMA_BASE_URL - it will use default
	os.Setenv("OLLAMA_MODEL_NAME", "llama3")
	defer func() {
		os.Unsetenv("AI_INGREDIENT_PROVIDER")
		os.Unsetenv("OLLAMA_MODEL_NAME")
	}()

	// Execute - should succeed with default base URL
	cfg, err := LoadConfig()

	// Assert
	if err != nil {
		t.Fatalf("expected no error with default BaseURL, got %v", err)
	}
	if cfg.Ollama.BaseURL != "http://ollama:11434" {
		t.Errorf("expected default BaseURL, got %s", cfg.Ollama.BaseURL)
	}
}

func TestLoadConfig_CloudFreeTier_Valid(t *testing.T) {
	// Setup
	os.Setenv("AI_INGREDIENT_PROVIDER", "cloud_free_tier")
	os.Setenv("CLOUD_AI_BASE_URL", "https://api.openai.com/v1")
	os.Setenv("CLOUD_AI_MODEL_NAME", "gpt-4o-mini")
	os.Setenv("CLOUD_AI_API_KEY", "sk-test-key")
	os.Setenv("CLOUD_AI_TIMEOUT_MS", "15000")
	os.Setenv("CLOUD_AI_MAX_RPS", "2")
	defer func() {
		os.Unsetenv("AI_INGREDIENT_PROVIDER")
		os.Unsetenv("CLOUD_AI_BASE_URL")
		os.Unsetenv("CLOUD_AI_MODEL_NAME")
		os.Unsetenv("CLOUD_AI_API_KEY")
		os.Unsetenv("CLOUD_AI_TIMEOUT_MS")
		os.Unsetenv("CLOUD_AI_MAX_RPS")
	}()

	// Execute
	cfg, err := LoadConfig()

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.Provider != ProviderCloudFreeTier {
		t.Errorf("expected provider to be cloud_free_tier, got %s", cfg.Provider)
	}
	if cfg.Cloud.BaseURL != "https://api.openai.com/v1" {
		t.Errorf("expected BaseURL to be https://api.openai.com/v1, got %s", cfg.Cloud.BaseURL)
	}
	if cfg.Cloud.APIKey != "sk-test-key" {
		t.Errorf("expected APIKey to be sk-test-key, got %s", cfg.Cloud.APIKey)
	}
	if cfg.Cloud.MaxRPS != 2 {
		t.Errorf("expected MaxRPS to be 2, got %d", cfg.Cloud.MaxRPS)
	}
	if !cfg.IsEnabled() {
		t.Error("expected IsEnabled() to return true for cloud_free_tier provider")
	}
}

func TestLoadConfig_CloudFreeTier_MissingAPIKey(t *testing.T) {
	// Setup
	os.Setenv("AI_INGREDIENT_PROVIDER", "cloud_free_tier")
	os.Setenv("CLOUD_AI_BASE_URL", "https://api.openai.com/v1")
	os.Setenv("CLOUD_AI_MODEL_NAME", "gpt-4o-mini")
	os.Setenv("CLOUD_AI_API_KEY", "")
	defer func() {
		os.Unsetenv("AI_INGREDIENT_PROVIDER")
		os.Unsetenv("CLOUD_AI_BASE_URL")
		os.Unsetenv("CLOUD_AI_MODEL_NAME")
		os.Unsetenv("CLOUD_AI_API_KEY")
	}()

	// Execute
	_, err := LoadConfig()

	// Assert
	if err == nil {
		t.Fatal("expected error for missing CLOUD_AI_API_KEY, got nil")
	}
}

func TestLoadConfig_InvalidProvider(t *testing.T) {
	// Setup
	os.Setenv("AI_INGREDIENT_PROVIDER", "invalid_provider")
	defer os.Unsetenv("AI_INGREDIENT_PROVIDER")

	// Execute
	_, err := LoadConfig()

	// Assert
	if err == nil {
		t.Fatal("expected error for invalid provider, got nil")
	}
}

func TestLoadConfig_Defaults(t *testing.T) {
	// Setup - use local_ollama with minimal config
	os.Setenv("AI_INGREDIENT_PROVIDER", "local_ollama")
	os.Setenv("OLLAMA_BASE_URL", "http://test:11434")
	// Don't set MODEL_NAME or TIMEOUT - should use defaults
	defer func() {
		os.Unsetenv("AI_INGREDIENT_PROVIDER")
		os.Unsetenv("OLLAMA_BASE_URL")
	}()

	// Execute
	cfg, err := LoadConfig()

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.Ollama.ModelName != "llama3" {
		t.Errorf("expected default ModelName to be llama3, got %s", cfg.Ollama.ModelName)
	}
	if cfg.Ollama.Timeout.Milliseconds() != 10000 {
		t.Errorf("expected default timeout to be 10000ms, got %d", cfg.Ollama.Timeout.Milliseconds())
	}
}
