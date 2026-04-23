// Package layout validates the dashboard layout JSON document. It is a pure
// function: no DB, no HTTP. It enforces PRD §4.9 rules and returns stable
// error codes that the REST layer maps to JSON:API error objects.
package layout

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	shared "github.com/jtumidanski/home-hub/shared/go/dashboard"
)

type Widget struct {
	ID     uuid.UUID       `json:"id"`
	Type   string          `json:"type"`
	X      int             `json:"x"`
	Y      int             `json:"y"`
	W      int             `json:"w"`
	H      int             `json:"h"`
	Config json.RawMessage `json:"config"`
}

type Layout struct {
	Version int      `json:"version"`
	Widgets []Widget `json:"widgets"`
}

type Code string

const (
	CodeUnsupportedSchemaVersion Code = "layout.unsupported_schema_version"
	CodeWidgetCountExceeded      Code = "layout.widget_count_exceeded"
	CodeWidgetUnknownType        Code = "layout.widget_unknown_type"
	CodeWidgetBadGeometry        Code = "layout.widget_bad_geometry"
	CodeWidgetBadID              Code = "layout.widget_bad_id"
	CodeWidgetDuplicateID        Code = "layout.widget_duplicate_id"
	CodeConfigTooLarge           Code = "layout.config_too_large"
	CodeConfigTooDeep            Code = "layout.config_too_deep"
	CodeConfigNotObject          Code = "layout.config_not_object"
	CodePayloadTooLarge          Code = "layout.payload_too_large"
	CodeMalformed                Code = "layout.malformed"
)

type ValidationError struct {
	Code    Code
	Pointer string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s at %s: %s", e.Code, e.Pointer, e.Message)
}

func Validate(raw json.RawMessage) (Layout, error) {
	if len(raw) > shared.MaxLayoutBytes {
		return Layout{}, ValidationError{Code: CodePayloadTooLarge, Pointer: "/data/attributes/layout",
			Message: fmt.Sprintf("layout exceeds %d bytes", shared.MaxLayoutBytes)}
	}
	var out Layout
	if err := json.Unmarshal(raw, &out); err != nil {
		return Layout{}, ValidationError{Code: CodeMalformed, Pointer: "/data/attributes/layout", Message: err.Error()}
	}
	if out.Version != shared.LayoutSchemaVersion {
		return Layout{}, ValidationError{Code: CodeUnsupportedSchemaVersion, Pointer: "/data/attributes/layout/version",
			Message: fmt.Sprintf("expected version %d, got %d", shared.LayoutSchemaVersion, out.Version)}
	}
	// Widget validations are added in later tasks.
	return out, nil
}
