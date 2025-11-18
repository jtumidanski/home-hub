package health

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	// serviceStartTime is set when the first handler is created
	serviceStartTime = time.Now()
)

// Handler creates an HTTP handler for the health check endpoint
// It returns a handler function that can be registered with a router
func Handler(serviceName, version string, aggregator *Aggregator, logger logrus.FieldLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set response headers
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

		// Execute all health checks
		checkResults := aggregator.Execute(r.Context())
		overallStatus := aggregator.OverallStatus(checkResults)

		// Calculate uptime
		uptime := time.Since(serviceStartTime)

		// Build response
		response := HealthResponse{
			Status:         overallStatus,
			Service:        serviceName,
			Version:        version,
			UptimeSeconds:  int64(uptime.Seconds()),
			Timestamp:      time.Now().UTC().Format(time.RFC3339),
			Checks:         checkResults,
		}

		// Set HTTP status code based on health
		statusCode := http.StatusOK
		if overallStatus == StatusUnhealthy {
			statusCode = http.StatusServiceUnavailable
		}

		// Log the health check
		logFields := logrus.Fields{
			"service":        serviceName,
			"overall_status": overallStatus,
			"uptime_seconds": response.UptimeSeconds,
		}

		if overallStatus == StatusUnhealthy {
			logger.WithFields(logFields).Warn("Health check returned unhealthy status")
		} else if overallStatus == StatusDegraded {
			logger.WithFields(logFields).Info("Health check returned degraded status")
		} else {
			logger.WithFields(logFields).Debug("Health check returned healthy status")
		}

		// Write response
		w.WriteHeader(statusCode)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.WithError(err).Error("Failed to encode health check response")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

// SimpleHandler creates a basic health handler without checks
// Useful for services that don't have dependencies to check
func SimpleHandler(serviceName, version string, logger logrus.FieldLogger) http.HandlerFunc {
	aggregator := NewAggregator()
	return Handler(serviceName, version, aggregator, logger)
}
