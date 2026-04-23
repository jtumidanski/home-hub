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
	if len(out.Widgets) > shared.MaxWidgets {
		return Layout{}, ValidationError{Code: CodeWidgetCountExceeded, Pointer: "/data/attributes/layout/widgets",
			Message: fmt.Sprintf("at most %d widgets allowed", shared.MaxWidgets)}
	}
	seen := make(map[uuid.UUID]struct{}, len(out.Widgets))
	for i, w := range out.Widgets {
		ptr := func(f string) string { return fmt.Sprintf("/data/attributes/layout/widgets/%d/%s", i, f) }
		if !shared.IsKnownWidgetType(w.Type) {
			return Layout{}, ValidationError{Code: CodeWidgetUnknownType, Pointer: ptr("type"),
				Message: fmt.Sprintf("widget type %q is not in the registry", w.Type)}
		}
		if w.X < 0 || w.Y < 0 || w.W < 1 || w.H < 1 || w.X+w.W > shared.GridColumns {
			return Layout{}, ValidationError{Code: CodeWidgetBadGeometry, Pointer: ptr(""), Message: "widget geometry out of grid"}
		}
		if w.ID == uuid.Nil {
			return Layout{}, ValidationError{Code: CodeWidgetBadID, Pointer: ptr("id"), Message: "widget id is required and must be a uuid"}
		}
		if _, dup := seen[w.ID]; dup {
			return Layout{}, ValidationError{Code: CodeWidgetDuplicateID, Pointer: ptr("id"), Message: "widget id is duplicated"}
		}
		seen[w.ID] = struct{}{}
	}
	return out, nil
}
