package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jtumidanski/home-hub/apps/svc-meals/ingredient"
	"github.com/sirupsen/logrus"
)

// Client is an HTTP client for the svc-ai service
type Client struct {
	baseURL string
	timeout time.Duration
	client  *http.Client
	logger  logrus.FieldLogger
}

// NewClient creates a new AI service client
func NewClient(baseURL string, timeout time.Duration, logger logrus.FieldLogger) *Client {
	return &Client{
		baseURL: baseURL,
		timeout: timeout,
		client: &http.Client{
			Timeout: timeout,
		},
		logger: logger.WithField("component", "ai_client"),
	}
}

// ParsedIngredientAttributes represents the attributes from JSON:API response
type ParsedIngredientAttributes struct {
	Line        string   `json:"line"`
	Quantity    *float64 `json:"quantity"`
	QuantityRaw string   `json:"quantityRaw"`
	Unit        *string  `json:"unit"`
	UnitRaw     *string  `json:"unitRaw"`
	Ingredient  string   `json:"ingredient"`
	Preparation []string `json:"preparation"`
	Notes       []string `json:"notes"`
	Confidence  float64  `json:"confidence"`
}

// JsonApiIngredientResource represents a JSON:API resource
type JsonApiIngredientResource struct {
	Type       string                      `json:"type"`
	ID         string                      `json:"id"`
	Attributes ParsedIngredientAttributes  `json:"attributes"`
}

// JsonApiIngredientsResponse represents the JSON:API array response
type JsonApiIngredientsResponse struct {
	Data []JsonApiIngredientResource `json:"data"`
	Meta map[string]interface{}      `json:"meta,omitempty"`
}

// BatchParseRequest represents a request to parse multiple ingredient lines
type BatchParseRequest struct {
	Lines  []string `json:"lines"`
	Locale string   `json:"locale,omitempty"`
}

// RecipeParseRequest represents a request to parse a full recipe
type RecipeParseRequest struct {
	RecipeText string `json:"recipeText"`
	Locale     string `json:"locale,omitempty"`
}

// ParseIngredients parses ingredient lines using the AI service
// Returns builders (not complete models) because meal ID is not known yet
func (c *Client) ParseIngredients(ctx context.Context, lines []string) ([]*ingredient.Builder, error) {
	if len(lines) == 0 {
		return []*ingredient.Builder{}, nil
	}

	// Prepare request
	reqBody := BatchParseRequest{
		Lines:  lines,
		Locale: "en-US",
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/parse/ingredients", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("AI service request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("AI service returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse JSON:API response
	var aiResp JsonApiIngredientsResponse
	if err := json.NewDecoder(resp.Body).Decode(&aiResp); err != nil {
		return nil, fmt.Errorf("failed to decode AI JSON:API response: %w", err)
	}

	// Convert JSON:API resources to ingredient builders (not models, as meal ID is not yet known)
	builders := make([]*ingredient.Builder, len(aiResp.Data))
	for i, resource := range aiResp.Data {
		attrs := resource.Attributes

		builders[i] = ingredient.New().
			WithRawLine(attrs.Line).
			WithQuantity(attrs.Quantity).
			WithQuantityRaw(attrs.QuantityRaw).
			WithUnit(attrs.Unit).
			WithUnitRaw(attrs.UnitRaw).
			WithIngredient(attrs.Ingredient).
			WithPreparation(attrs.Preparation).
			WithNotes(attrs.Notes).
			WithConfidence(attrs.Confidence)
	}

	return builders, nil
}

// ParseRecipe parses a full recipe text and extracts ingredients using the AI service
// Returns builders (not complete models) because meal ID is not known yet
func (c *Client) ParseRecipe(ctx context.Context, recipeText string) ([]*ingredient.Builder, error) {
	if recipeText == "" {
		return []*ingredient.Builder{}, nil
	}

	// Prepare request
	reqBody := RecipeParseRequest{
		RecipeText: recipeText,
		Locale:     "en-US",
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/parse/recipe", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Log request details
	c.logger.WithFields(logrus.Fields{
		"url":           c.baseURL + "/api/parse/recipe",
		"recipe_length": len(recipeText),
		"timeout":       c.timeout,
	}).Debug("Sending request to AI service")

	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.WithError(err).Error("AI service request failed")
		return nil, fmt.Errorf("AI service request failed: %w", err)
	}
	defer resp.Body.Close()

	c.logger.WithFields(logrus.Fields{
		"status":         resp.StatusCode,
		"content_length": resp.Header.Get("Content-Length"),
	}).Debug("Received response from AI service")

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.WithFields(logrus.Fields{
			"status": resp.StatusCode,
			"body":   string(body),
		}).Error("AI service returned error status")
		return nil, fmt.Errorf("AI service returned status %d: %s", resp.StatusCode, string(body))
	}

	// Read response body
	c.logger.Debug("Reading response body from AI service")
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.WithError(err).Error("Failed to read response body")
		return nil, fmt.Errorf("failed to read AI response body: %w", err)
	}

	c.logger.WithField("bytes_read", len(bodyBytes)).Debug("Response body read successfully")

	// Parse JSON:API response
	var aiResp JsonApiIngredientsResponse
	if err := json.Unmarshal(bodyBytes, &aiResp); err != nil {
		sample := string(bodyBytes)
		if len(sample) > 500 {
			sample = sample[:500] + "..."
		}
		c.logger.WithFields(logrus.Fields{
			"error":          err.Error(),
			"response_sample": sample,
		}).Error("Failed to parse JSON:API response")
		return nil, fmt.Errorf("failed to decode AI JSON:API response: %w", err)
	}

	c.logger.WithField("ingredient_count", len(aiResp.Data)).Debug("Successfully parsed JSON:API response")

	// Convert JSON:API resources to ingredient builders (not models, as meal ID is not yet known)
	builders := make([]*ingredient.Builder, len(aiResp.Data))
	for i, resource := range aiResp.Data {
		attrs := resource.Attributes

		c.logger.WithFields(logrus.Fields{
			"index":      i + 1,
			"ingredient": attrs.Ingredient,
			"confidence": attrs.Confidence,
		}).Debug("Processing ingredient")

		builders[i] = ingredient.New().
			WithRawLine(attrs.Line).
			WithQuantity(attrs.Quantity).
			WithQuantityRaw(attrs.QuantityRaw).
			WithUnit(attrs.Unit).
			WithUnitRaw(attrs.UnitRaw).
			WithIngredient(attrs.Ingredient).
			WithPreparation(attrs.Preparation).
			WithNotes(attrs.Notes).
			WithConfidence(attrs.Confidence)
	}

	c.logger.WithField("builder_count", len(builders)).Debug("Successfully converted ingredients to builders")

	return builders, nil
}

// HealthCheck verifies the AI service is reachable
func (c *Client) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/health", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("AI service health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("AI service health check returned status %d", resp.StatusCode)
	}

	return nil
}
