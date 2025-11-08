package server

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/packages/shared-go/tenant"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type SpanHandler func(logrus.FieldLogger, context.Context) http.HandlerFunc

//goland:noinspection GoUnusedExportedFunction
func RetrieveSpan(l logrus.FieldLogger, name string, ctx context.Context, next SpanHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		propagator := otel.GetTextMapPropagator()
		sctx := propagator.Extract(ctx, propagation.HeaderCarrier(r.Header))
		sctx, span := otel.GetTracerProvider().Tracer("atlas-rest").Start(sctx, name)
		sl := l.WithField("trace.id", span.SpanContext().TraceID().String()).WithField("span.id", span.SpanContext().SpanID().String())
		defer span.End()
		next(sl, sctx)(w, r)
	}
}

type TenantHandler func(logrus.FieldLogger, context.Context) http.HandlerFunc

//goland:noinspection GoUnusedExportedFunction
func ParseTenant(l logrus.FieldLogger, ctx context.Context, next TenantHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.Header.Get(tenant.ID)
		if idStr == "" {
			l.Errorf("%s is not supplied.", tenant.ID)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		id, err := uuid.Parse(idStr)
		if err != nil {
			l.Errorf("%s is not supplied.", tenant.ID)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		tl := l.WithField("tenant", id.String())

		t, err := tenant.Create(id)
		if err != nil {
			l.Errorf("Failed to create tenant with provided data.")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		tctx := tenant.WithContext(ctx, t)
		next(tl, tctx)(w, r)
	}
}
