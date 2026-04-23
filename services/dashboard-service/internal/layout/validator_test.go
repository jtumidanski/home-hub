package layout

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func mustJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

func TestValidateEmptyWidgets(t *testing.T) {
	raw := mustJSON(map[string]any{"version": 1, "widgets": []any{}})
	l, err := Validate(raw)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if l.Version != 1 || len(l.Widgets) != 0 {
		t.Fatalf("bad layout: %+v", l)
	}
}

func TestValidateRejectsWrongVersion(t *testing.T) {
	raw := mustJSON(map[string]any{"version": 2, "widgets": []any{}})
	_, err := Validate(raw)
	if err == nil {
		t.Fatal("expected error")
	}
	ve, ok := err.(ValidationError)
	if !ok || ve.Code != CodeUnsupportedSchemaVersion {
		t.Fatalf("expected %s, got %v", CodeUnsupportedSchemaVersion, err)
	}
}

func TestValidateRejectsTooManyWidgets(t *testing.T) {
	widgets := make([]map[string]any, 41)
	for i := range widgets {
		widgets[i] = map[string]any{"id": uuid.New().String(), "type": "weather", "x": 0, "y": i, "w": 12, "h": 1, "config": map[string]any{}}
	}
	raw := mustJSON(map[string]any{"version": 1, "widgets": widgets})
	_, err := Validate(raw)
	ve, ok := err.(ValidationError)
	if !ok || ve.Code != CodeWidgetCountExceeded {
		t.Fatalf("expected %s, got %v", CodeWidgetCountExceeded, err)
	}
}

func TestValidateRejectsUnknownWidgetType(t *testing.T) {
	raw := mustJSON(map[string]any{"version": 1, "widgets": []any{
		map[string]any{"id": uuid.New().String(), "type": "foo", "x": 0, "y": 0, "w": 1, "h": 1, "config": map[string]any{}},
	}})
	_, err := Validate(raw)
	ve, ok := err.(ValidationError)
	if !ok || ve.Code != CodeWidgetUnknownType || ve.Pointer != "/data/attributes/layout/widgets/0/type" {
		t.Fatalf("got %v", err)
	}
}
