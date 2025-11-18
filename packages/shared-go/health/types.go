package health

import (
	"context"
	"time"
)

// HealthStatus represents the overall health status of a service or component
type HealthStatus string

const (
	// StatusHealthy indicates the component is fully operational
	StatusHealthy HealthStatus = "healthy"
	// StatusDegraded indicates the component is operational but with reduced functionality
	StatusDegraded HealthStatus = "degraded"
	// StatusUnhealthy indicates the component is not operational
	StatusUnhealthy HealthStatus = "unhealthy"
)

// String returns the string representation of a HealthStatus
func (s HealthStatus) String() string {
	return string(s)
}

// Check represents a health check that can be executed
type Check interface {
	// Name returns the name of this health check
	Name() string
	// Execute performs the health check and returns the result
	Execute(ctx context.Context) CheckResult
}

// CheckResult represents the result of a single health check
type CheckResult struct {
	// Status is the health status of this check
	Status HealthStatus `json:"status"`
	// ResponseTimeMs is the time taken to execute the check in milliseconds
	ResponseTimeMs int64 `json:"response_time_ms,omitempty"`
	// Error is the error message if the check failed (optional)
	Error string `json:"error,omitempty"`
	// Metadata contains additional check-specific information (optional)
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// HealthResponse represents the complete health status of a service
type HealthResponse struct {
	// Status is the overall health status
	Status HealthStatus `json:"status"`
	// Service is the name of the service
	Service string `json:"service"`
	// Version is the version of the service
	Version string `json:"version"`
	// UptimeSeconds is how long the service has been running
	UptimeSeconds int64 `json:"uptime_seconds"`
	// Timestamp is when this health check was performed (ISO 8601)
	Timestamp string `json:"timestamp"`
	// Checks contains the results of individual health checks
	Checks map[string]CheckResult `json:"checks"`
}

// StatusFromError determines health status from an error
// Returns StatusHealthy if err is nil, StatusUnhealthy otherwise
func StatusFromError(err error) HealthStatus {
	if err == nil {
		return StatusHealthy
	}
	return StatusUnhealthy
}

// WorstStatus returns the worst status from a list of statuses
// Unhealthy > Degraded > Healthy
func WorstStatus(statuses ...HealthStatus) HealthStatus {
	worst := StatusHealthy
	for _, status := range statuses {
		if status == StatusUnhealthy {
			return StatusUnhealthy
		}
		if status == StatusDegraded {
			worst = StatusDegraded
		}
	}
	return worst
}

// measureTime measures the execution time of a function and returns duration in milliseconds
func measureTime(fn func() error) (int64, error) {
	start := time.Now()
	err := fn()
	duration := time.Since(start)
	return duration.Milliseconds(), err
}
