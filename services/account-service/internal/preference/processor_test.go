package preference

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus/hooks/test"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	db.AutoMigrate(&Entity{})
	return db
}

func TestFindOrCreate_Creates(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	tenantID := uuid.New()
	userID := uuid.New()

	m, err := p.FindOrCreate(tenantID, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Theme() != "light" {
		t.Errorf("expected default theme light, got %s", m.Theme())
	}
	if m.ActiveHouseholdID() != nil {
		t.Error("expected nil active household for new preference")
	}
}

func TestFindOrCreate_FindsExisting(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	tenantID := uuid.New()
	userID := uuid.New()

	first, _ := p.FindOrCreate(tenantID, userID)
	second, err := p.FindOrCreate(tenantID, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if first.Id() != second.Id() {
		t.Errorf("expected same preference ID, got %s and %s", first.Id(), second.Id())
	}
}

func TestUpdateTheme(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	m, _ := p.FindOrCreate(uuid.New(), uuid.New())
	updated, err := p.UpdateTheme(m.Id(), "dark")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Theme() != "dark" {
		t.Errorf("expected theme dark, got %s", updated.Theme())
	}
}

func TestSetActiveHousehold(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	m, _ := p.FindOrCreate(uuid.New(), uuid.New())
	hhID := uuid.New()
	updated, err := p.SetActiveHousehold(m.Id(), hhID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.ActiveHouseholdID() == nil {
		t.Fatal("expected non-nil active household")
	}
	if *updated.ActiveHouseholdID() != hhID {
		t.Errorf("expected household %s, got %s", hhID, *updated.ActiveHouseholdID())
	}
}
