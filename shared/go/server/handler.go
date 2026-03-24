package server

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
)

// jsonapiError represents a JSON:API error object.
type jsonapiError struct {
	Status string `json:"status"`
	Code   string `json:"code,omitempty"`
	Title  string `json:"title"`
	Detail string `json:"detail,omitempty"`
}

type jsonapiErrors struct {
	Errors []jsonapiError `json:"errors"`
}

// WriteError writes a JSON:API error response.
func WriteError(w http.ResponseWriter, status int, title string, detail string) {
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(jsonapiErrors{
		Errors: []jsonapiError{
			{
				Status: http.StatusText(status),
				Title:  title,
				Detail: detail,
			},
		},
	})
}

// MarshalResponse writes a JSON:API success response.
func MarshalResponse[T any](w http.ResponseWriter, status int, data T) {
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]T{"data": data})
}

// MarshalSliceResponse writes a JSON:API success response for a list.
func MarshalSliceResponse[T any](w http.ResponseWriter, status int, data []T) {
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string][]T{"data": data})
}

// GetHandler is the signature for handlers that don't accept a request body.
type GetHandler func(l logrus.FieldLogger, w http.ResponseWriter, r *http.Request)

// InputHandler is the signature for handlers that accept a typed request body.
type InputHandler[T any] func(l logrus.FieldLogger, w http.ResponseWriter, r *http.Request, input T)

// RegisterHandler wraps a GetHandler with logging and error recovery.
func RegisterHandler(l *logrus.Logger) func(name string, handler GetHandler) http.HandlerFunc {
	return func(name string, handler GetHandler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			entry := l.WithField("handler", name)
			handler(entry, w, r)
		}
	}
}

// RegisterInputHandler wraps an InputHandler with JSON decoding, logging, and error recovery.
func RegisterInputHandler[T any](l *logrus.Logger) func(name string, handler InputHandler[T]) http.HandlerFunc {
	return func(name string, handler InputHandler[T]) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			entry := l.WithField("handler", name)

			var input T
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
				entry.WithError(err).Warn("invalid request body")
				WriteError(w, http.StatusBadRequest, "Invalid Request", "Could not parse request body")
				return
			}

			handler(entry, w, r, input)
		}
	}
}
