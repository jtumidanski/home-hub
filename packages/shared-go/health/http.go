package health

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPCheck checks the health of an HTTP service dependency
type HTTPCheck struct {
	serviceURL string
	name       string
	client     *http.Client
}

// NewHTTPCheck creates a new HTTP dependency health check
func NewHTTPCheck(serviceURL string, name string) Check {
	return &HTTPCheck{
		serviceURL: serviceURL,
		name:       name,
		client: &http.Client{
			Timeout: 2 * time.Second,
		},
	}
}

// Name returns the name of this check
func (c *HTTPCheck) Name() string {
	return c.name
}

// Execute performs the HTTP health check
func (c *HTTPCheck) Execute(ctx context.Context) CheckResult {
	// Measure execution time
	responseTime, err := measureTime(func() error {
		return c.checkHealth(ctx)
	})

	result := CheckResult{
		ResponseTimeMs: responseTime,
	}

	if err != nil {
		// Degraded instead of unhealthy for dependency failures
		// This allows the service to continue operating with reduced functionality
		result.Status = StatusDegraded
		result.Error = err.Error()
	} else {
		result.Status = StatusHealthy
	}

	return result
}

// checkHealth performs the actual HTTP health check
func (c *HTTPCheck) checkHealth(ctx context.Context) error {
	healthURL := fmt.Sprintf("%s/health", c.serviceURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to reach service: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check HTTP status code
	if resp.StatusCode >= 500 {
		return fmt.Errorf("service returned error status: %d", resp.StatusCode)
	}

	// Try to parse health response
	var healthResp HealthResponse
	if err := json.Unmarshal(body, &healthResp); err != nil {
		// If we can't parse the response but got 200, consider it healthy
		if resp.StatusCode == http.StatusOK {
			return nil
		}
		return fmt.Errorf("failed to parse health response: %w", err)
	}

	// Check the service's reported status
	if healthResp.Status == StatusUnhealthy {
		return fmt.Errorf("service reports unhealthy status")
	}

	return nil
}
