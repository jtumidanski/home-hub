package source

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/database"
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
	l, _ := test.NewNullLogger()
	database.RegisterTenantCallbacks(l, db)
	if err := db.AutoMigrate(&Entity{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func newProcessor(t *testing.T, db *gorm.DB) *Processor {
	t.Helper()
	l, _ := test.NewNullLogger()
	return NewProcessor(l, context.Background(), db)
}

func TestCreateOrUpdate_CreatesNew(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)

	tenantID := uuid.New()
	householdID := uuid.New()
	connID := uuid.New()

	m, err := p.CreateOrUpdate(tenantID, householdID, connID, "ext-cal-1", "Work Calendar", true, "#4285F4")
	if err != nil {
		t.Fatalf("CreateOrUpdate (create): %v", err)
	}

	if m.ExternalID() != "ext-cal-1" {
		t.Errorf("expected externalID 'ext-cal-1', got %q", m.ExternalID())
	}
	if m.Name() != "Work Calendar" {
		t.Errorf("expected name 'Work Calendar', got %q", m.Name())
	}
	if !m.Primary() {
		t.Error("expected primary to be true")
	}
	if !m.Visible() {
		t.Error("expected visible to be true by default")
	}
	if m.Color() != "#4285F4" {
		t.Errorf("expected color '#4285F4', got %q", m.Color())
	}
}

func TestCreateOrUpdate_UpdatesExisting(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)

	tenantID := uuid.New()
	householdID := uuid.New()
	connID := uuid.New()

	original, err := p.CreateOrUpdate(tenantID, householdID, connID, "ext-cal-1", "Old Name", false, "#FF0000")
	if err != nil {
		t.Fatalf("CreateOrUpdate (create): %v", err)
	}

	updated, err := p.CreateOrUpdate(tenantID, householdID, connID, "ext-cal-1", "New Name", true, "#00FF00")
	if err != nil {
		t.Fatalf("CreateOrUpdate (update): %v", err)
	}

	if updated.Id() != original.Id() {
		t.Errorf("expected same ID after update, got %v != %v", original.Id(), updated.Id())
	}
	if updated.Name() != "New Name" {
		t.Errorf("expected name 'New Name', got %q", updated.Name())
	}
	if updated.Color() != "#00FF00" {
		t.Errorf("expected color '#00FF00', got %q", updated.Color())
	}

	var count int64
	db.WithContext(database.WithoutTenantFilter(context.Background())).Model(&Entity{}).
		Where("connection_id = ?", connID).Count(&count)
	if count != 1 {
		t.Errorf("expected 1 source (no duplicates), got %d", count)
	}
}

func TestListByConnection(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)

	tenantID := uuid.New()
	householdID := uuid.New()
	connID := uuid.New()

	if _, err := p.CreateOrUpdate(tenantID, householdID, connID, "cal-1", "Calendar 1", true, "#FF0000"); err != nil {
		t.Fatalf("CreateOrUpdate: %v", err)
	}
	if _, err := p.CreateOrUpdate(tenantID, householdID, connID, "cal-2", "Calendar 2", false, "#00FF00"); err != nil {
		t.Fatalf("CreateOrUpdate: %v", err)
	}
	if _, err := p.CreateOrUpdate(tenantID, householdID, uuid.New(), "cal-3", "Other Conn", false, "#0000FF"); err != nil {
		t.Fatalf("CreateOrUpdate: %v", err)
	}

	sources, err := p.ListByConnection(connID)
	if err != nil {
		t.Fatalf("ListByConnection: %v", err)
	}
	if len(sources) != 2 {
		t.Errorf("expected 2 sources for connection, got %d", len(sources))
	}
}

func TestToggleVisibility(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)

	tenantID := uuid.New()
	householdID := uuid.New()
	connID := uuid.New()

	m, err := p.CreateOrUpdate(tenantID, householdID, connID, "cal-1", "Calendar", false, "#FF0000")
	if err != nil {
		t.Fatalf("CreateOrUpdate: %v", err)
	}
	if !m.Visible() {
		t.Fatal("expected visible=true initially")
	}

	if err := p.ToggleVisibility(m.Id(), false); err != nil {
		t.Fatalf("ToggleVisibility: %v", err)
	}

	var e Entity
	db.WithContext(database.WithoutTenantFilter(context.Background())).First(&e, "id = ?", m.Id())
	if e.Visible {
		t.Error("expected visible=false after toggle")
	}
}

func TestUpdateSyncToken(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)

	m, err := p.CreateOrUpdate(uuid.New(), uuid.New(), uuid.New(), "cal-1", "Calendar", false, "#FF0000")
	if err != nil {
		t.Fatalf("CreateOrUpdate: %v", err)
	}

	if err := p.UpdateSyncToken(m.Id(), "sync-token-abc"); err != nil {
		t.Fatalf("UpdateSyncToken: %v", err)
	}

	var e Entity
	db.WithContext(database.WithoutTenantFilter(context.Background())).First(&e, "id = ?", m.Id())
	if e.SyncToken != "sync-token-abc" {
		t.Errorf("expected sync token 'sync-token-abc', got %q", e.SyncToken)
	}
}

func TestClearSyncToken(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)

	m, err := p.CreateOrUpdate(uuid.New(), uuid.New(), uuid.New(), "cal-1", "Calendar", false, "#FF0000")
	if err != nil {
		t.Fatalf("CreateOrUpdate: %v", err)
	}
	if err := p.UpdateSyncToken(m.Id(), "some-token"); err != nil {
		t.Fatalf("UpdateSyncToken: %v", err)
	}

	if err := p.ClearSyncToken(m.Id()); err != nil {
		t.Fatalf("ClearSyncToken: %v", err)
	}

	var e Entity
	db.WithContext(database.WithoutTenantFilter(context.Background())).First(&e, "id = ?", m.Id())
	if e.SyncToken != "" {
		t.Errorf("expected empty sync token, got %q", e.SyncToken)
	}
}

func TestDeleteByConnection_RemovesAllSources(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)

	tenantID := uuid.New()
	householdID := uuid.New()
	connID := uuid.New()

	if _, err := p.CreateOrUpdate(tenantID, householdID, connID, "cal-1", "Calendar 1", true, "#FF0000"); err != nil {
		t.Fatalf("CreateOrUpdate: %v", err)
	}
	if _, err := p.CreateOrUpdate(tenantID, householdID, connID, "cal-2", "Calendar 2", false, "#00FF00"); err != nil {
		t.Fatalf("CreateOrUpdate: %v", err)
	}
	if _, err := p.CreateOrUpdate(tenantID, householdID, uuid.New(), "cal-3", "Other", false, "#0000FF"); err != nil {
		t.Fatalf("CreateOrUpdate: %v", err)
	}

	if err := p.DeleteByConnection(connID); err != nil {
		t.Fatalf("DeleteByConnection: %v", err)
	}

	var count int64
	db.WithContext(database.WithoutTenantFilter(context.Background())).Model(&Entity{}).Count(&count)
	if count != 1 {
		t.Errorf("expected 1 remaining source, got %d", count)
	}
}

func TestByIDProvider_ReturnsModel(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)

	created, err := p.CreateOrUpdate(uuid.New(), uuid.New(), uuid.New(), "cal-1", "Calendar", false, "#FF0000")
	if err != nil {
		t.Fatalf("CreateOrUpdate: %v", err)
	}

	got, err := p.ByIDProvider(created.Id())()
	if err != nil {
		t.Fatalf("ByIDProvider: %v", err)
	}
	if got.Id() != created.Id() {
		t.Errorf("expected id %v, got %v", created.Id(), got.Id())
	}
	if got.Name() != "Calendar" {
		t.Errorf("expected name 'Calendar', got %q", got.Name())
	}
}

func TestByIDProvider_NotFound(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)

	_, err := p.ByIDProvider(uuid.New())()
	if err == nil {
		t.Fatal("expected error for non-existent source")
	}
}
