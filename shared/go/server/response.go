package server

import (
	"encoding/json"
	"net/http"

	"github.com/jtumidanski/api2go/jsonapi"
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

// WriteJSONAPIError writes a single JSON:API error object with an optional
// stable error code and source pointer. Use this when the caller needs to
// machine-distinguish validation failures (e.g., UI widget highlighting).
func WriteJSONAPIError(w http.ResponseWriter, status int, code, title, detail, pointer string) {
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(status)
	type sourceObj struct {
		Pointer string `json:"pointer,omitempty"`
	}
	type detailed struct {
		Status string     `json:"status"`
		Code   string     `json:"code,omitempty"`
		Title  string     `json:"title"`
		Detail string     `json:"detail,omitempty"`
		Source *sourceObj `json:"source,omitempty"`
	}
	body := struct {
		Errors []detailed `json:"errors"`
	}{Errors: []detailed{{Status: http.StatusText(status), Code: code, Title: title, Detail: detail}}}
	if pointer != "" {
		body.Errors[0].Source = &sourceObj{Pointer: pointer}
	}
	json.NewEncoder(w).Encode(body)
}

// MarshalResponse marshals a single resource using api2go and writes the response.
// The curried signature matches the guidelines pattern:
//
//	server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(queryParams)(model)
func MarshalResponse[T jsonapi.MarshalIdentifier](l logrus.FieldLogger) func(w http.ResponseWriter) func(si jsonapi.ServerInformation) func(params map[string][]string) func(data T) {
	return func(w http.ResponseWriter) func(si jsonapi.ServerInformation) func(params map[string][]string) func(data T) {
		return func(si jsonapi.ServerInformation) func(params map[string][]string) func(data T) {
			return func(params map[string][]string) func(data T) {
				return func(data T) {
					result, err := jsonapi.MarshalWithURLs(data, si)
					if err != nil {
						l.WithError(err).Error("failed to marshal response")
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					w.Header().Set("Content-Type", "application/vnd.api+json")
					w.WriteHeader(http.StatusOK)
					w.Write(result)
				}
			}
		}
	}
}

// MarshalCreatedResponse is like MarshalResponse but returns 201.
func MarshalCreatedResponse[T jsonapi.MarshalIdentifier](l logrus.FieldLogger) func(w http.ResponseWriter) func(si jsonapi.ServerInformation) func(data T) {
	return func(w http.ResponseWriter) func(si jsonapi.ServerInformation) func(data T) {
		return func(si jsonapi.ServerInformation) func(data T) {
			return func(data T) {
				result, err := jsonapi.MarshalWithURLs(data, si)
				if err != nil {
					l.WithError(err).Error("failed to marshal response")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/vnd.api+json")
				w.WriteHeader(http.StatusCreated)
				w.Write(result)
			}
		}
	}
}

// MarshalSliceResponse marshals a slice of resources using api2go.
func MarshalSliceResponse[T jsonapi.MarshalIdentifier](l logrus.FieldLogger) func(w http.ResponseWriter) func(si jsonapi.ServerInformation) func(data []T) {
	return func(w http.ResponseWriter) func(si jsonapi.ServerInformation) func(data []T) {
		return func(si jsonapi.ServerInformation) func(data []T) {
			return func(data []T) {
				// Convert to []jsonapi.MarshalIdentifier for api2go
				items := make([]jsonapi.MarshalIdentifier, len(data))
				for i, d := range data {
					items[i] = d
				}
				result, err := jsonapi.MarshalWithURLs(items, si)
				if err != nil {
					l.WithError(err).Error("failed to marshal slice response")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/vnd.api+json")
				w.WriteHeader(http.StatusOK)
				w.Write(result)
			}
		}
	}
}

// ServerInfo implements jsonapi.ServerInformation for api2go marshaling.
type ServerInfo struct {
	BaseURL string
	Prefix  string
}

func (s ServerInfo) GetBaseURL() string { return s.BaseURL }
func (s ServerInfo) GetPrefix() string  { return s.Prefix }

// GetServerInformation returns the default server information.
func GetServerInformation() jsonapi.ServerInformation {
	return ServerInfo{BaseURL: "", Prefix: "/api/v1"}
}
