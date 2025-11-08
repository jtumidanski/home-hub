package requests

import (
	"context"
	"net/http"

	"github.com/jtumidanski/home-hub/packages/shared-go/tenant"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type HeaderDecorator func(header http.Header)

//goland:noinspection GoUnusedExportedFunction
func SpanHeaderDecorator(ctx context.Context) HeaderDecorator {
	return func(h http.Header) {
		carrier := propagation.MapCarrier{}
		propagator := otel.GetTextMapPropagator()
		propagator.Inject(ctx, carrier)
		for _, k := range carrier.Keys() {
			h.Set(k, carrier.Get(k))
		}
	}
}

//goland:noinspection GoUnusedExportedFunction
func TenantHeaderDecorator(ctx context.Context) HeaderDecorator {
	return func(h http.Header) {
		h.Set("Content-Type", "application/json; charset=utf-8")

		t, err := tenant.FromContext(ctx)()
		if err != nil {
			return
		}

		h.Set(tenant.ID, t.Id().String())
	}
}
