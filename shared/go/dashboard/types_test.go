package dashboard

import (
	"encoding/json"
	"os"
	"sort"
	"testing"
)

func TestIsKnownWidgetType(t *testing.T) {
	for _, typ := range []string{
		"weather", "tasks-summary", "reminders-summary", "overdue-summary",
		"meal-plan-today", "calendar-today", "packages-summary",
		"habits-today", "workout-today",
		"tasks-today", "reminders-today",
		"weather-tomorrow", "calendar-tomorrow", "tasks-tomorrow",
	} {
		if !IsKnownWidgetType(typ) {
			t.Fatalf("expected %q to be known", typ)
		}
	}
	if IsKnownWidgetType("foo") {
		t.Fatalf("expected unknown type to be rejected")
	}
}

func TestWidgetTypesParityFixture(t *testing.T) {
	data, err := os.ReadFile("fixtures/widget-types.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var fixture []string
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatalf("parse fixture: %v", err)
	}
	got := make([]string, 0, len(WidgetTypes))
	for k := range WidgetTypes {
		got = append(got, k)
	}
	sort.Strings(got)
	want := append([]string(nil), fixture...)
	sort.Strings(want)
	if len(got) != len(want) {
		t.Fatalf("got %d types, fixture has %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("mismatch at %d: go=%q fixture=%q", i, got[i], want[i])
		}
	}
}

func TestLayoutConstants(t *testing.T) {
	if LayoutSchemaVersion != 1 {
		t.Fatalf("schema version must be 1")
	}
	if MaxWidgets != 40 || GridColumns != 12 {
		t.Fatalf("cap constants drifted")
	}
	if MaxLayoutBytes != 64*1024 {
		t.Fatalf("layout byte cap drifted")
	}
	if MaxWidgetConfigBytes != 4*1024 {
		t.Fatalf("widget config byte cap drifted")
	}
	if MaxWidgetConfigDepth != 5 {
		t.Fatalf("config depth cap drifted")
	}
}
