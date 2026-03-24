package appcontext

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/account-service/internal/household"
	"github.com/jtumidanski/home-hub/services/account-service/internal/membership"
	"github.com/jtumidanski/home-hub/services/account-service/internal/preference"
	"github.com/jtumidanski/home-hub/services/account-service/internal/tenant"
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
	db.AutoMigrate(&tenant.Entity{})
	db.AutoMigrate(&household.Entity{})
	db.AutoMigrate(&membership.Entity{})
	db.AutoMigrate(&preference.Entity{})
	return db
}

func TestResolve_HappyPath(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	ctx := context.Background()

	// Create tenant
	tenantProc := tenant.NewProcessor(l, ctx, db)
	ten, err := tenantProc.Create("Test Tenant")
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	userID := uuid.New()

	// Create household with owner
	hhProc := household.NewProcessor(l, ctx, db)
	hh, err := hhProc.CreateWithOwner(ten.Id(), userID, "Test Home", "UTC", "metric")
	if err != nil {
		t.Fatalf("failed to create household: %v", err)
	}

	resolved, err := Resolve(l, ctx, db, ten.Id(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resolved.Tenant.Id() != ten.Id() {
		t.Errorf("expected tenant %s, got %s", ten.Id(), resolved.Tenant.Id())
	}
	if resolved.ActiveHousehold == nil {
		t.Fatal("expected active household")
	}
	if resolved.ActiveHousehold.Id() != hh.Id() {
		t.Errorf("expected household %s, got %s", hh.Id(), resolved.ActiveHousehold.Id())
	}
	if resolved.ResolvedRole != "owner" {
		t.Errorf("expected role owner, got %s", resolved.ResolvedRole)
	}
	if !resolved.CanCreateHousehold {
		t.Error("expected canCreateHousehold to be true for owner")
	}
	if len(resolved.Memberships) != 1 {
		t.Errorf("expected 1 membership, got %d", len(resolved.Memberships))
	}
}

func TestResolve_NoMemberships(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	ctx := context.Background()

	tenantProc := tenant.NewProcessor(l, ctx, db)
	ten, err := tenantProc.Create("Empty Tenant")
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	userID := uuid.New()

	resolved, err := Resolve(l, ctx, db, ten.Id(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resolved.ActiveHousehold != nil {
		t.Error("expected nil active household when no memberships")
	}
	if resolved.ResolvedRole != "" {
		t.Errorf("expected empty role, got %s", resolved.ResolvedRole)
	}
	if resolved.CanCreateHousehold {
		t.Error("expected canCreateHousehold to be false with no memberships")
	}
}

func TestResolve_ResolvesFirstHouseholdWhenNoActiveSet(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	ctx := context.Background()

	tenantProc := tenant.NewProcessor(l, ctx, db)
	ten, err := tenantProc.Create("Resolve Tenant")
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	userID := uuid.New()

	// Create household with owner — preference will have no activeHouseholdId
	hhProc := household.NewProcessor(l, ctx, db)
	hh, err := hhProc.CreateWithOwner(ten.Id(), userID, "First Home", "UTC", "metric")
	if err != nil {
		t.Fatalf("failed to create household: %v", err)
	}

	resolved, err := Resolve(l, ctx, db, ten.Id(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resolved.ActiveHousehold == nil {
		t.Fatal("expected active household to be resolved from first membership")
	}
	if resolved.ActiveHousehold.Id() != hh.Id() {
		t.Errorf("expected household %s, got %s", hh.Id(), resolved.ActiveHousehold.Id())
	}
}

func TestResolve_TenantNotFound(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	ctx := context.Background()

	_, err := Resolve(l, ctx, db, uuid.New(), uuid.New())
	if err == nil {
		t.Error("expected error for non-existent tenant")
	}
}
