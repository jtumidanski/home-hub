package server

import (
	"net/http"

	"github.com/jtumidanski/home-hub/shared/go/logging"
	"go.opentelemetry.io/otel/trace"
)

// TracingMiddleware adds trace and span IDs to the request context log fields.
// Requires OTel instrumentation to be active on the handler (e.g., via otelhttp).
func TracingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span := trace.SpanFromContext(r.Context())
		if span.SpanContext().IsValid() {
			ctx := logging.WithField(r.Context(), "trace_id", span.SpanContext().TraceID().String())
			ctx = logging.WithField(ctx, "span_id", span.SpanContext().SpanID().String())
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}
