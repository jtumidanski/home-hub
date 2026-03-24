package tenant

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
	db.AutoMigrate(&Entity{})
	return db
}

func TestCreate(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	m, err := p.Create("Test Tenant")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Name() != "Test Tenant" {
		t.Errorf("expected name Test Tenant, got %s", m.Name())
	}
	if m.Id() == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
}

func TestByIDProvider(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	created, _ := p.Create("Lookup Tenant")
	found, err := p.ByIDProvider(created.Id())()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found.Name() != "Lookup Tenant" {
		t.Errorf("expected name Lookup Tenant, got %s", found.Name())
	}
}

func TestByIDProvider_NotFound(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	_, err := p.ByIDProvider(uuid.New())()
	if err == nil {
		t.Error("expected error for non-existent tenant")
	}
}
