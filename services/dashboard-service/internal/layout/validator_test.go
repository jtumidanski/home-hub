package layout

import (
	"encoding/json"
	"testing"
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
