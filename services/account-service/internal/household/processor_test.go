package household

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/jtumidanski/home-hub/services/account-service/internal/membership"
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
	db.AutoMigrate(&membership.Entity{})
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

func TestAllProvider(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	tenantID := uuid.New()
	p.Create(tenantID, "Home 1", "UTC", "metric")
	p.Create(tenantID, "Home 2", "UTC", "metric")

	models, err := p.AllProvider()()
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

func TestCreateWithOwner(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	tenantID := uuid.New()
	userID := uuid.New()
	m, err := p.CreateWithOwner(tenantID, userID, "My House", "UTC", "metric")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Name() != "My House" {
		t.Errorf("expected name My House, got %s", m.Name())
	}

	// Verify owner membership was created
	memProc := membership.NewProcessor(l, context.Background(), db)
	mem, err := memProc.ByHouseholdAndUserProvider(m.Id(), userID)()
	if err != nil {
		t.Fatalf("expected owner membership, got error: %v", err)
	}
	if mem.Role() != "owner" {
		t.Errorf("expected role owner, got %s", mem.Role())
	}
}
