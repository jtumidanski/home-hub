package health

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// mockCheck is a mock health check for testing
type mockCheck struct {
	name   string
	result CheckResult
}

func (m *mockCheck) Name() string {
	return m.name
}

func (m *mockCheck) Execute(ctx context.Context) CheckResult {
	return m.result
}

func TestHandler_AllHealthy(t *testing.T) {
	// Create aggregator with healthy checks
	aggregator := NewAggregator(
		&mockCheck{
			name: "database",
			result: CheckResult{
				Status:         StatusHealthy,
				ResponseTimeMs: 5,
			},
		},
		&mockCheck{
			name: "redis",
			result: CheckResult{
				Status:         StatusHealthy,
				ResponseTimeMs: 3,
			},
		},
	)

	// Create handler
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Suppress logs in tests
	handler := Handler("test-service", "1.0.0", aggregator, logger)

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	// Execute handler
	handler(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response HealthResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, StatusHealthy, response.Status)
	assert.Equal(t, "test-service", response.Service)
	assert.Equal(t, "1.0.0", response.Version)
	assert.GreaterOrEqual(t, response.UptimeSeconds, int64(0))
	assert.NotEmpty(t, response.Timestamp)
	assert.Len(t, response.Checks, 2)
	assert.Equal(t, StatusHealthy, response.Checks["database"].Status)
	assert.Equal(t, StatusHealthy, response.Checks["redis"].Status)
}

func TestHandler_OneUnhealthy(t *testing.T) {
	// Create aggregator with one unhealthy check
	aggregator := NewAggregator(
		&mockCheck{
			name: "database",
			result: CheckResult{
				Status:         StatusHealthy,
				ResponseTimeMs: 5,
			},
		},
		&mockCheck{
			name: "redis",
			result: CheckResult{
				Status:         StatusUnhealthy,
				ResponseTimeMs: 1000,
				Error:          "connection refused",
			},
		},
	)

	// Create handler
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)
	handler := Handler("test-service", "1.0.0", aggregator, logger)

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	// Execute handler
	handler(w, req)

	// Verify response
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var response HealthResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, StatusUnhealthy, response.Status)
	assert.Equal(t, StatusHealthy, response.Checks["database"].Status)
	assert.Equal(t, StatusUnhealthy, response.Checks["redis"].Status)
	assert.Equal(t, "connection refused", response.Checks["redis"].Error)
}

func TestHandler_OneDegraded(t *testing.T) {
	// Create aggregator with one degraded check
	aggregator := NewAggregator(
		&mockCheck{
			name: "database",
			result: CheckResult{
				Status:         StatusHealthy,
				ResponseTimeMs: 5,
			},
		},
		&mockCheck{
			name: "dependency",
			result: CheckResult{
				Status:         StatusDegraded,
				ResponseTimeMs: 2000,
				Error:          "slow response",
			},
		},
	)

	// Create handler
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)
	handler := Handler("test-service", "1.0.0", aggregator, logger)

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	// Execute handler
	handler(w, req)

	// Verify response - degraded should return 200 OK
	assert.Equal(t, http.StatusOK, w.Code)

	var response HealthResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, StatusDegraded, response.Status)
}

func TestSimpleHandler(t *testing.T) {
	// Create simple handler (no checks)
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)
	handler := SimpleHandler("test-service", "1.0.0", logger)

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	// Execute handler
	handler(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response HealthResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, StatusHealthy, response.Status)
	assert.Len(t, response.Checks, 0)
}

func TestWorstStatus(t *testing.T) {
	tests := []struct {
		name     string
		statuses []HealthStatus
		expected HealthStatus
	}{
		{
			name:     "all healthy",
			statuses: []HealthStatus{StatusHealthy, StatusHealthy},
			expected: StatusHealthy,
		},
		{
			name:     "one degraded",
			statuses: []HealthStatus{StatusHealthy, StatusDegraded},
			expected: StatusDegraded,
		},
		{
			name:     "one unhealthy",
			statuses: []HealthStatus{StatusHealthy, StatusUnhealthy},
			expected: StatusUnhealthy,
		},
		{
			name:     "unhealthy trumps degraded",
			statuses: []HealthStatus{StatusDegraded, StatusUnhealthy},
			expected: StatusUnhealthy,
		},
		{
			name:     "empty",
			statuses: []HealthStatus{},
			expected: StatusHealthy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WorstStatus(tt.statuses...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStatusFromError(t *testing.T) {
	assert.Equal(t, StatusHealthy, StatusFromError(nil))
	assert.Equal(t, StatusUnhealthy, StatusFromError(assert.AnError))
}
