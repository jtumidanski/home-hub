package logging

import (
	"context"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// InitTracing initializes OpenTelemetry tracing with an OTLP HTTP exporter.
// Returns a shutdown function that should be called on service exit.
// If no OTLP endpoint is configured, a no-op provider is used.
func InitTracing(l *logrus.Logger, serviceName string) func(context.Context) error {
	ctx := context.Background()

	exporter, err := otlptracehttp.New(ctx)
	if err != nil {
		l.WithError(err).Warn("failed to create OTLP exporter, tracing disabled")
		return func(context.Context) error { return nil }
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		l.WithError(err).Warn("failed to create tracing resource")
		return func(context.Context) error { return nil }
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	l.WithField("service", serviceName).Info("tracing initialized")

	return func(ctx context.Context) error {
		return tp.Shutdown(ctx)
	}
}
