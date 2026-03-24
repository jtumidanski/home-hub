package household

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

func TestCreate(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	tenantID := uuid.New()
	m, err := p.Create(tenantID, "Main Home", "America/Detroit", "imperial")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Name() != "Main Home" {
		t.Errorf("expected name Main Home, got %s", m.Name())
	}
	if m.TenantID() != tenantID {
		t.Errorf("expected tenant ID %s, got %s", tenantID, m.TenantID())
	}
	if m.Timezone() != "America/Detroit" {
		t.Errorf("expected timezone America/Detroit, got %s", m.Timezone())
	}
}

func TestByTenantIDProvider(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	tenantID := uuid.New()
	p.Create(tenantID, "Home 1", "UTC", "metric")
	p.Create(tenantID, "Home 2", "UTC", "metric")
	p.Create(uuid.New(), "Other Tenant Home", "UTC", "metric")

	models, err := p.ByTenantIDProvider(tenantID)()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(models) != 2 {
		t.Errorf("expected 2 households, got %d", len(models))
	}
}

func TestUpdate(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	m, _ := p.Create(uuid.New(), "Old Name", "UTC", "metric")
	updated, err := p.Update(m.Id(), "New Name", "America/Chicago", "imperial")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name() != "New Name" {
		t.Errorf("expected name New Name, got %s", updated.Name())
	}
	if updated.Timezone() != "America/Chicago" {
		t.Errorf("expected timezone America/Chicago, got %s", updated.Timezone())
	}
}
