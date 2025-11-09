package server

import (
	"context"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/jtumidanski/api2go/jsonapi"
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

type HandlerDependency struct {
	l   logrus.FieldLogger
	ctx context.Context
}

func (h HandlerDependency) Logger() logrus.FieldLogger {
	return h.l
}

func (h HandlerDependency) Context() context.Context {
	return h.ctx
}

type HandlerContext struct {
	si jsonapi.ServerInformation
}

func (h HandlerContext) ServerInformation() jsonapi.ServerInformation {
	return h.si
}

type GetHandler func(d *HandlerDependency, c *HandlerContext) http.HandlerFunc

type InputHandler[M any] func(d *HandlerDependency, c *HandlerContext, model M) http.HandlerFunc

func ParseInput[M any](d *HandlerDependency, c *HandlerContext, next InputHandler[M]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var model M

		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		err = jsonapi.Unmarshal(body, &model)
		if err != nil {
			d.l.WithError(err).Errorln("Deserializing input", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		next(d, c, model)(w, r)
	}
}

//goland:noinspection GoUnusedExportedFunction
func RegisterHandler(l logrus.FieldLogger) func(si jsonapi.ServerInformation) func(handlerName string, handler GetHandler) http.HandlerFunc {
	return func(si jsonapi.ServerInformation) func(handlerName string, handler GetHandler) http.HandlerFunc {
		return func(handlerName string, handler GetHandler) http.HandlerFunc {
			return RetrieveSpan(l, handlerName, context.Background(), func(sl logrus.FieldLogger, sctx context.Context) http.HandlerFunc {
				fl := sl.WithFields(logrus.Fields{"originator": handlerName, "type": "rest_handler"})
				return ParseTenant(fl, sctx, func(tl logrus.FieldLogger, tctx context.Context) http.HandlerFunc {
					return handler(&HandlerDependency{l: tl, ctx: tctx}, &HandlerContext{si: si})
				})
			})
		}
	}
}

//goland:noinspection GoUnusedExportedFunction
func RegisterInputHandler[M any](l logrus.FieldLogger) func(si jsonapi.ServerInformation) func(handlerName string, handler InputHandler[M]) http.HandlerFunc {
	return func(si jsonapi.ServerInformation) func(handlerName string, handler InputHandler[M]) http.HandlerFunc {
		return func(handlerName string, handler InputHandler[M]) http.HandlerFunc {
			return RetrieveSpan(l, handlerName, context.Background(), func(sl logrus.FieldLogger, sctx context.Context) http.HandlerFunc {
				fl := sl.WithFields(logrus.Fields{"originator": handlerName, "type": "rest_handler"})
				return ParseTenant(fl, sctx, func(tl logrus.FieldLogger, tctx context.Context) http.HandlerFunc {
					return ParseInput[M](&HandlerDependency{l: tl, ctx: tctx}, &HandlerContext{si: si}, handler)
				})
			})
		}
	}
}
