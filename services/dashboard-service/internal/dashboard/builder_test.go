package dashboard

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

func TestBuilderRequiresIDs(t *testing.T) {
	_, err := NewBuilder().Build()
	if err == nil {
		t.Fatal("expected error for missing ids")
	}
}

func TestBuilderRoundTrip(t *testing.T) {
	tid, hid, uid := uuid.New(), uuid.New(), uuid.New()
	raw, _ := json.Marshal(map[string]any{"version": 1, "widgets": []any{}})
	m, err := NewBuilder().
		SetId(uuid.New()).
		SetTenantID(tid).
		SetHouseholdID(hid).
		SetUserID(&uid).
		SetName("Weekend").
		SetSortOrder(2).
		SetLayout(datatypes.JSON(raw)).
		SetSchemaVersion(1).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	if m.Name() != "Weekend" || m.SortOrder() != 2 {
		t.Fatalf("round trip mismatch: %+v", m)
	}
	if m.UserID() == nil || *m.UserID() != uid {
		t.Fatalf("user id mismatch")
	}
}

func TestBuilderTrimsName(t *testing.T) {
	tid, hid := uuid.New(), uuid.New()
	raw, _ := json.Marshal(map[string]any{"version": 1, "widgets": []any{}})
	m, err := NewBuilder().
		SetId(uuid.New()).
		SetTenantID(tid).
		SetHouseholdID(hid).
		SetName("   Home   ").
		SetLayout(datatypes.JSON(raw)).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	if m.Name() != "Home" {
		t.Fatalf("expected trim, got %q", m.Name())
	}
}

func TestBuilderRejectsEmptyName(t *testing.T) {
	tid, hid := uuid.New(), uuid.New()
	raw, _ := json.Marshal(map[string]any{"version": 1, "widgets": []any{}})
	_, err := NewBuilder().
		SetId(uuid.New()).
		SetTenantID(tid).
		SetHouseholdID(hid).
		SetName("  ").
		SetLayout(datatypes.JSON(raw)).
		Build()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBuilderRejectsLongName(t *testing.T) {
	tid, hid := uuid.New(), uuid.New()
	raw, _ := json.Marshal(map[string]any{"version": 1, "widgets": []any{}})
	_, err := NewBuilder().
		SetId(uuid.New()).
		SetTenantID(tid).
		SetHouseholdID(hid).
		SetName(string(make([]byte, 81))).
		SetLayout(datatypes.JSON(raw)).
		Build()
	if err == nil {
		t.Fatal("expected error")
	}
}
