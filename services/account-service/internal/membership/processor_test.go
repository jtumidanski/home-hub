package membership

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
	householdID := uuid.New()
	userID := uuid.New()

	m, err := p.Create(tenantID, householdID, userID, "owner")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Role() != "owner" {
		t.Errorf("expected role owner, got %s", m.Role())
	}
	if m.HouseholdID() != householdID {
		t.Errorf("expected household ID %s, got %s", householdID, m.HouseholdID())
	}
}

func TestUpdateRole(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	m, _ := p.Create(uuid.New(), uuid.New(), uuid.New(), "viewer")
	updated, err := p.UpdateRole(m.Id(), "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Role() != "admin" {
		t.Errorf("expected role admin, got %s", updated.Role())
	}
}

func TestDelete(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	m, _ := p.Create(uuid.New(), uuid.New(), uuid.New(), "editor")
	if err := p.Delete(m.Id()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := p.ByIDProvider(m.Id())()
	if err == nil {
		t.Error("expected error for deleted membership")
	}
}

func TestByUserAndTenantProvider(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	tenantID := uuid.New()
	userID := uuid.New()
	p.Create(tenantID, uuid.New(), userID, "owner")
	p.Create(tenantID, uuid.New(), userID, "editor")
	p.Create(uuid.New(), uuid.New(), userID, "viewer") // different tenant

	models, err := p.ByUserAndTenantProvider(userID, tenantID)()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(models) != 2 {
		t.Errorf("expected 2 memberships, got %d", len(models))
	}
}
