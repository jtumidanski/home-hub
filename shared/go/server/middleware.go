package server

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/logging"
	"github.com/sirupsen/logrus"
)

// RequestIDMiddleware generates a unique request ID and adds it to the
// request context and response headers.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		ctx := logging.WithField(r.Context(), "request_id", requestID)
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LoggingMiddleware logs each request with structured fields.
func LoggingMiddleware(l *logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logging.WithContext(l, r.Context()).WithFields(logrus.Fields{
				"method": r.Method,
				"path":   r.URL.Path,
			}).Info("request received")
			next.ServeHTTP(w, r)
		})
	}
}
