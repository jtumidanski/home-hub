package layout

import (
	"encoding/json"
	"fmt"
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

func TestValidateRejectsBadGeometry(t *testing.T) {
	cases := []struct {
		name       string
		x, y, w, h int
	}{
		{"negative x", -1, 0, 1, 1},
		{"negative y", 0, -1, 1, 1},
		{"zero w", 0, 0, 0, 1},
		{"zero h", 0, 0, 1, 0},
		{"overflows grid", 10, 0, 4, 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			raw := mustJSON(map[string]any{"version": 1, "widgets": []any{
				map[string]any{"id": uuid.New().String(), "type": "weather", "x": c.x, "y": c.y, "w": c.w, "h": c.h, "config": map[string]any{}},
			}})
			_, err := Validate(raw)
			ve, ok := err.(ValidationError)
			if !ok || ve.Code != CodeWidgetBadGeometry {
				t.Fatalf("%s: got %v", c.name, err)
			}
		})
	}
}

func TestValidateRejectsBadID(t *testing.T) {
	raw := mustJSON(map[string]any{"version": 1, "widgets": []any{
		map[string]any{"id": "not-a-uuid", "type": "weather", "x": 0, "y": 0, "w": 1, "h": 1, "config": map[string]any{}},
	}})
	_, err := Validate(raw)
	ve, ok := err.(ValidationError)
	if !ok || (ve.Code != CodeWidgetBadID && ve.Code != CodeMalformed) {
		t.Fatalf("got %v", err)
	}
}

func TestValidateRejectsDuplicateID(t *testing.T) {
	id := uuid.New().String()
	raw := mustJSON(map[string]any{"version": 1, "widgets": []any{
		map[string]any{"id": id, "type": "weather", "x": 0, "y": 0, "w": 1, "h": 1, "config": map[string]any{}},
		map[string]any{"id": id, "type": "weather", "x": 0, "y": 1, "w": 1, "h": 1, "config": map[string]any{}},
	}})
	_, err := Validate(raw)
	ve, ok := err.(ValidationError)
	if !ok || ve.Code != CodeWidgetDuplicateID {
		t.Fatalf("got %v", err)
	}
}

func TestValidateRejectsNonObjectConfig(t *testing.T) {
	raw := mustJSON(map[string]any{"version": 1, "widgets": []any{
		map[string]any{"id": uuid.New().String(), "type": "weather", "x": 0, "y": 0, "w": 1, "h": 1, "config": []any{1, 2, 3}},
	}})
	_, err := Validate(raw)
	ve, _ := err.(ValidationError)
	if ve.Code != CodeConfigNotObject {
		t.Fatalf("got %v", err)
	}
}

func TestValidateRejectsOversizedConfig(t *testing.T) {
	big := make(map[string]string)
	for i := 0; i < 500; i++ {
		big[fmt.Sprintf("k%d", i)] = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	}
	raw := mustJSON(map[string]any{"version": 1, "widgets": []any{
		map[string]any{"id": uuid.New().String(), "type": "weather", "x": 0, "y": 0, "w": 1, "h": 1, "config": big},
	}})
	_, err := Validate(raw)
	ve, _ := err.(ValidationError)
	if ve.Code != CodeConfigTooLarge {
		t.Fatalf("got %v", err)
	}
}

func TestValidateRejectsDeepConfig(t *testing.T) {
	var nest any = "leaf"
	for i := 0; i < 10; i++ {
		nest = map[string]any{"x": nest}
	}
	raw := mustJSON(map[string]any{"version": 1, "widgets": []any{
		map[string]any{"id": uuid.New().String(), "type": "weather", "x": 0, "y": 0, "w": 1, "h": 1, "config": nest},
	}})
	_, err := Validate(raw)
	ve, _ := err.(ValidationError)
	if ve.Code != CodeConfigTooDeep {
		t.Fatalf("got %v", err)
	}
}
