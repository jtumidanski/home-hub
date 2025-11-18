package parser

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// CloudParser implements IngredientParser using a cloud LLM API (OpenAI compatible)
type CloudParser struct {
	baseURL   string
	modelName string
	apiKey    string
	timeout   time.Duration
	client    *http.Client
	logger    logrus.FieldLogger
}

// NewCloudParser creates a new cloud-based ingredient parser
func NewCloudParser(baseURL, modelName, apiKey string, timeout time.Duration, logger logrus.FieldLogger) *CloudParser {
	return &CloudParser{
		baseURL:   baseURL,
		modelName: modelName,
		apiKey:    apiKey,
		timeout:   timeout,
		client: &http.Client{
			Timeout: timeout,
		},
		logger: logger.WithField("parser", "cloud").WithField("model", modelName),
	}
}

// Name returns the parser name
func (p *CloudParser) Name() string {
	return "cloud_free_tier"
}

// Parse parses a single ingredient line
func (p *CloudParser) Parse(ctx context.Context, line string, opts ParseOptions) (ParseResult, error) {
	startTime := time.Now()

	// Build prompt
	systemPrompt := SystemPrompt()
	userPrompt := UserPrompt(line, opts)

	// Create OpenAI-compatible request
	reqBody := map[string]interface{}{
		"model": p.modelName,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0.1, // Low temperature for consistent parsing
		"max_tokens":  500,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return ParseResult{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return ParseResult{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return ParseResult{}, fmt.Errorf("cloud API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return ParseResult{}, fmt.Errorf("cloud API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse OpenAI response
	var cloudResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&cloudResp); err != nil {
		return ParseResult{}, fmt.Errorf("failed to decode cloud response: %w", err)
	}

	if len(cloudResp.Choices) == 0 {
		return ParseResult{}, fmt.Errorf("cloud API returned no choices")
	}

	// Parse the LLM's JSON response
	parsed, err := ParseLLMResponse(cloudResp.Choices[0].Message.Content)
	if err != nil {
		return ParseResult{}, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	latency := time.Since(startTime)

	return ParseResult{
		Line:   line,
		Parsed: *parsed,
		Provider: ProviderInfo{
			Name:      p.Name(),
			Model:     p.modelName,
			LatencyMs: latency.Milliseconds(),
		},
		Warnings: generateWarnings(*parsed),
	}, nil
}

// ParseBatch parses multiple ingredient lines
func (p *CloudParser) ParseBatch(ctx context.Context, lines []string, opts ParseOptions) ([]ParseResult, error) {
	results := make([]ParseResult, len(lines))

	for i, line := range lines {
		result, err := p.Parse(ctx, line, opts)
		if err != nil {
			// For batch processing, we don't want to fail the entire batch
			// Instead, return a partial result with low confidence
			results[i] = ParseResult{
				Line: line,
				Parsed: ParsedIngredient{
					QuantityRaw: line,
					Ingredient:  line, // Fallback: use entire line as ingredient
					Preparation: []string{},
					Notes:       []string{},
					Confidence:  0.0,
				},
				Provider: ProviderInfo{
					Name:  p.Name(),
					Model: p.modelName,
				},
				Warnings: []string{"parsing_failed"},
			}
			continue
		}
		results[i] = result
	}

	return results, nil
}

// ParseRecipe extracts and parses ingredients from full recipe text
func (p *CloudParser) ParseRecipe(ctx context.Context, recipeText string, opts ParseOptions) ([]ParseResult, error) {
	startTime := time.Now()

	// Build prompt for recipe extraction
	systemPrompt := RecipeExtractionPrompt()
	userPrompt := RecipeExtractionUserPrompt(recipeText, opts)

	// Create OpenAI-compatible request
	reqBody := map[string]interface{}{
		"model": p.modelName,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0.1, // Low temperature for consistent parsing
		"max_tokens":  2000, // More tokens for recipe extraction
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cloud API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("cloud API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse OpenAI response
	var cloudResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&cloudResp); err != nil {
		return nil, fmt.Errorf("failed to decode cloud response: %w", err)
	}

	if len(cloudResp.Choices) == 0 {
		return nil, fmt.Errorf("cloud API returned no choices")
	}

	// Parse the LLM's JSON array response
	parsedIngredients, err := ParseRecipeExtractionResponse(cloudResp.Choices[0].Message.Content, p.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to parse recipe extraction response: %w", err)
	}

	latency := time.Since(startTime)

	// Convert to ParseResult format
	results := make([]ParseResult, len(parsedIngredients))
	for i, parsed := range parsedIngredients {
		results[i] = ParseResult{
			Line:   parsed.QuantityRaw + " " + parsed.Ingredient, // Reconstruct line for reference
			Parsed: parsed,
			Provider: ProviderInfo{
				Name:      p.Name(),
				Model:     p.modelName,
				LatencyMs: latency.Milliseconds(),
			},
			Warnings: generateWarnings(parsed),
		}
	}

	return results, nil
}

// HealthCheck verifies the cloud API is reachable
func (p *CloudParser) HealthCheck(ctx context.Context) error {
	// Simple health check: make a minimal API request
	reqBody := map[string]interface{}{
		"model": p.modelName,
		"messages": []map[string]string{
			{"role": "user", "content": "ping"},
		},
		"max_tokens": 1,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal health check request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("cloud API health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cloud API health check returned status %d", resp.StatusCode)
	}

	return nil
}
