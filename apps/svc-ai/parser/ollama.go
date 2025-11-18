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

// OllamaParser implements IngredientParser using a local Ollama instance
type OllamaParser struct {
	baseURL   string
	modelName string
	timeout   time.Duration
	client    *http.Client
	logger    logrus.FieldLogger
}

// NewOllamaParser creates a new Ollama-based ingredient parser
func NewOllamaParser(baseURL, modelName string, timeout time.Duration, logger logrus.FieldLogger) *OllamaParser {
	return &OllamaParser{
		baseURL:   baseURL,
		modelName: modelName,
		timeout:   timeout,
		client: &http.Client{
			Timeout: timeout,
		},
		logger: logger.WithField("parser", "ollama").WithField("model", modelName),
	}
}

// Name returns the parser name
func (p *OllamaParser) Name() string {
	return "local_ollama"
}

// Parse parses a single ingredient line
func (p *OllamaParser) Parse(ctx context.Context, line string, opts ParseOptions) (ParseResult, error) {
	startTime := time.Now()

	// Build prompt
	systemPrompt := SystemPrompt()
	userPrompt := UserPrompt(line, opts)

	// Create Ollama request
	reqBody := map[string]interface{}{
		"model":  p.modelName,
		"prompt": fmt.Sprintf("%s\n\n%s", systemPrompt, userPrompt),
		"stream": false,
		"options": map[string]interface{}{
			"temperature": 0.1, // Low temperature for consistent parsing
			"top_p":       0.9,
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return ParseResult{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/generate", bytes.NewBuffer(jsonBody))
	if err != nil {
		return ParseResult{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return ParseResult{}, fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return ParseResult{}, fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse Ollama response
	var ollamaResp struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return ParseResult{}, fmt.Errorf("failed to decode ollama response: %w", err)
	}

	// Parse the LLM's JSON response
	parsed, err := ParseLLMResponse(ollamaResp.Response)
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
func (p *OllamaParser) ParseBatch(ctx context.Context, lines []string, opts ParseOptions) ([]ParseResult, error) {
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
func (p *OllamaParser) ParseRecipe(ctx context.Context, recipeText string, opts ParseOptions) ([]ParseResult, error) {
	startTime := time.Now()

	p.logger.WithFields(logrus.Fields{
		"recipe_length": len(recipeText),
		"locale":        opts.Locale,
	}).Debug("Starting recipe parsing")

	// Build prompt for recipe extraction
	systemPrompt := RecipeExtractionPrompt()
	userPrompt := RecipeExtractionUserPrompt(recipeText, opts)

	// Create Ollama request
	reqBody := map[string]interface{}{
		"model":  p.modelName,
		"prompt": fmt.Sprintf("%s\n\n%s", systemPrompt, userPrompt),
		"stream": false,
		"options": map[string]interface{}{
			"temperature": 0.1, // Low temperature for consistent parsing
			"top_p":       0.9,
		},
	}

	p.logger.WithField("base_url", p.baseURL).Debug("Sending request to Ollama")

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/generate", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse Ollama response
	var ollamaResp struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode ollama response: %w", err)
	}

	p.logger.WithField("response_length", len(ollamaResp.Response)).Debug("Received LLM response")

	// Log first 500 chars of response for debugging
	responseSample := ollamaResp.Response
	if len(responseSample) > 500 {
		responseSample = responseSample[:500] + "..."
	}
	p.logger.WithField("response_sample", responseSample).Debug("Raw LLM response")

	// Parse the LLM's JSON array response
	parsedIngredients, err := ParseRecipeExtractionResponse(ollamaResp.Response, p.logger)
	if err != nil {
		p.logger.WithError(err).Error("Failed to parse LLM response")
		return nil, fmt.Errorf("failed to parse recipe extraction response: %w", err)
	}

	p.logger.WithField("ingredient_count", len(parsedIngredients)).Debug("Successfully parsed ingredients from LLM response")

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

// HealthCheck verifies the Ollama instance is reachable
func (p *OllamaParser) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/api/tags", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("ollama health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama health check returned status %d", resp.StatusCode)
	}

	return nil
}

// generateWarnings generates warnings based on parsed data
func generateWarnings(parsed ParsedIngredient) []string {
	var warnings []string

	if parsed.Confidence < 0.7 {
		warnings = append(warnings, "low_confidence")
	}

	if parsed.Quantity == nil && parsed.QuantityRaw != "" {
		warnings = append(warnings, "low_confidence_quantity")
	}

	return warnings
}
