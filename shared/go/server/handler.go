package server

import (
	"io"
	"net/http"

	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/sirupsen/logrus"
)

// HandlerDependency provides pre-configured dependencies to handlers.
type HandlerDependency struct {
	logger logrus.FieldLogger
}

// Logger returns the pre-configured logger with handler context.
func (d *HandlerDependency) Logger() logrus.FieldLogger { return d.logger }

// HandlerContext provides JSON:API context to handlers.
type HandlerContext struct {
	si jsonapi.ServerInformation
}

// ServerInformation returns the JSON:API server configuration.
func (c *HandlerContext) ServerInformation() jsonapi.ServerInformation { return c.si }

// GetHandler is the signature for handlers that don't accept a request body.
type GetHandler func(d *HandlerDependency, c *HandlerContext) http.HandlerFunc

// InputHandler is the signature for handlers that accept a typed request body.
type InputHandler[T any] func(d *HandlerDependency, c *HandlerContext, input T) http.HandlerFunc

// RegisterHandler wraps a GetHandler with logging, tracing, and JSON:API context.
func RegisterHandler(l logrus.FieldLogger) func(si jsonapi.ServerInformation) func(name string, handler GetHandler) http.HandlerFunc {
	return func(si jsonapi.ServerInformation) func(name string, handler GetHandler) http.HandlerFunc {
		return func(name string, handler GetHandler) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				entry := l.WithField("handler", name)
				d := &HandlerDependency{logger: entry}
				c := &HandlerContext{si: si}
				handler(d, c)(w, r)
			}
		}
	}
}

// RegisterInputHandler wraps an InputHandler with JSON:API unmarshaling, logging, and context.
func RegisterInputHandler[T jsonapi.EntityNamer](l logrus.FieldLogger) func(si jsonapi.ServerInformation) func(name string, handler InputHandler[T]) http.HandlerFunc {
	return func(si jsonapi.ServerInformation) func(name string, handler InputHandler[T]) http.HandlerFunc {
		return func(name string, handler InputHandler[T]) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				entry := l.WithField("handler", name)
				d := &HandlerDependency{logger: entry}
				c := &HandlerContext{si: si}

				var input T
				body, _ := io.ReadAll(r.Body)
				if err := jsonapi.Unmarshal(body, &input); err != nil {
					entry.WithError(err).Warn("invalid request body")
					WriteError(w, http.StatusBadRequest, "Invalid Request", "Could not parse request body")
					return
				}

				handler(d, c, input)(w, r)
			}
		}
	}
}
