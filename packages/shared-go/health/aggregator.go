package health

import (
	"context"
)

// Aggregator combines multiple health checks into a single result
type Aggregator struct {
	checks []Check
}

// NewAggregator creates a new health check aggregator
func NewAggregator(checks ...Check) *Aggregator {
	return &Aggregator{
		checks: checks,
	}
}

// AddCheck adds a health check to the aggregator
func (a *Aggregator) AddCheck(check Check) {
	a.checks = append(a.checks, check)
}

// Execute runs all health checks and returns the aggregated results
func (a *Aggregator) Execute(ctx context.Context) map[string]CheckResult {
	results := make(map[string]CheckResult)

	// Execute all checks sequentially
	// For now, we run them sequentially to keep it simple
	// Could be optimized with goroutines and channels if needed
	for _, check := range a.checks {
		result := check.Execute(ctx)
		results[check.Name()] = result
	}

	return results
}

// OverallStatus determines the overall health status from check results
// Returns the worst status found among all checks
func (a *Aggregator) OverallStatus(results map[string]CheckResult) HealthStatus {
	if len(results) == 0 {
		return StatusHealthy
	}

	statuses := make([]HealthStatus, 0, len(results))
	for _, result := range results {
		statuses = append(statuses, result.Status)
	}

	return WorstStatus(statuses...)
}
