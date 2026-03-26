package invitation

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/account-service/internal/household"
	"github.com/jtumidanski/home-hub/services/account-service/internal/membership"
	"github.com/jtumidanski/home-hub/services/account-service/internal/preference"
	"github.com/jtumidanski/home-hub/shared/go/database"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
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
	for _, m := range []func(*gorm.DB) error{
		func(db *gorm.DB) error { return db.AutoMigrate(&Entity{}) },
		membership.Migration,
		preference.Migration,
		household.Migration,
	} {
		if err := m(db); err != nil {
			t.Fatalf("migration failed: %v", err)
		}
	}
	return db
}

func withTenantCtx(tenantID, userID uuid.UUID) context.Context {
	return tenantctx.WithContext(context.Background(), tenantctx.New(tenantID, uuid.Nil, userID))
}

func createTestHousehold(t *testing.T, db *gorm.DB, tenantID uuid.UUID, name string) household.Model {
	t.Helper()
	l, _ := test.NewNullLogger()
	ctx := withTenantCtx(tenantID, uuid.New())
	proc := household.NewProcessor(l, ctx, db)
	m, err := proc.Create(tenantID, name, "UTC", "imperial")
	if err != nil {
		t.Fatalf("failed to create household: %v", err)
	}
	return m
}

func createTestMembership(t *testing.T, db *gorm.DB, tenantID, householdID, userID uuid.UUID, role string) membership.Model {
	t.Helper()
	l, _ := test.NewNullLogger()
	ctx := withTenantCtx(tenantID, userID)
	proc := membership.NewProcessor(l, ctx, db)
	m, err := proc.Create(tenantID, householdID, userID, role)
	if err != nil {
		t.Fatalf("failed to create membership: %v", err)
	}
	return m
}

func TestProcessorCreate(t *testing.T) {
	db := setupTestDB(t)
	tenantID := uuid.New()
	ownerID := uuid.New()
	hh := createTestHousehold(t, db, tenantID, "Test Home")
	createTestMembership(t, db, tenantID, hh.Id(), ownerID, "owner")

	l, _ := test.NewNullLogger()

	t.Run("create invitation", func(t *testing.T) {
		ctx := withTenantCtx(tenantID, ownerID)
		proc := NewProcessor(l, ctx, db)
		m, err := proc.Create(tenantID, hh.Id(), "invitee@example.com", "viewer", ownerID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if m.Email() != "invitee@example.com" {
			t.Errorf("expected email invitee@example.com, got %s", m.Email())
		}
		if m.Role() != "viewer" {
			t.Errorf("expected role viewer, got %s", m.Role())
		}
		if m.Status() != "pending" {
			t.Errorf("expected status pending, got %s", m.Status())
		}
	})

	t.Run("duplicate invitation returns error", func(t *testing.T) {
		ctx := withTenantCtx(tenantID, ownerID)
		proc := NewProcessor(l, ctx, db)
		_, err := proc.Create(tenantID, hh.Id(), "invitee@example.com", "editor", ownerID)
		if err != ErrAlreadyInvited {
			t.Errorf("expected ErrAlreadyInvited, got %v", err)
		}
	})

	t.Run("non-privileged user cannot create", func(t *testing.T) {
		viewerID := uuid.New()
		createTestMembership(t, db, tenantID, hh.Id(), viewerID, "viewer")
		ctx := withTenantCtx(tenantID, viewerID)
		proc := NewProcessor(l, ctx, db)
		_, err := proc.Create(tenantID, hh.Id(), "another@example.com", "viewer", viewerID)
		if err != ErrNotAuthorized {
			t.Errorf("expected ErrNotAuthorized, got %v", err)
		}
	})

	t.Run("default role to viewer", func(t *testing.T) {
		ctx := withTenantCtx(tenantID, ownerID)
		proc := NewProcessor(l, ctx, db)
		m, err := proc.Create(tenantID, hh.Id(), "defaultrole@example.com", "", ownerID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if m.Role() != "viewer" {
			t.Errorf("expected default role viewer, got %s", m.Role())
		}
	})
}

func TestProcessorRevoke(t *testing.T) {
	db := setupTestDB(t)
	tenantID := uuid.New()
	ownerID := uuid.New()
	hh := createTestHousehold(t, db, tenantID, "Test Home")
	createTestMembership(t, db, tenantID, hh.Id(), ownerID, "owner")

	l, _ := test.NewNullLogger()

	t.Run("revoke pending invitation", func(t *testing.T) {
		ctx := withTenantCtx(tenantID, ownerID)
		proc := NewProcessor(l, ctx, db)
		inv, _ := proc.Create(tenantID, hh.Id(), "revoke@example.com", "viewer", ownerID)

		err := proc.Revoke(inv.Id(), ownerID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		updated, _ := proc.ByIDProvider(inv.Id())()
		if updated.Status() != "revoked" {
			t.Errorf("expected status revoked, got %s", updated.Status())
		}
	})

	t.Run("non-privileged user cannot revoke", func(t *testing.T) {
		viewerID := uuid.New()
		createTestMembership(t, db, tenantID, hh.Id(), viewerID, "viewer")
		ctx := withTenantCtx(tenantID, ownerID)
		proc := NewProcessor(l, ctx, db)
		inv, _ := proc.Create(tenantID, hh.Id(), "revokeauth@example.com", "viewer", ownerID)

		ctx2 := withTenantCtx(tenantID, viewerID)
		proc2 := NewProcessor(l, ctx2, db)
		err := proc2.Revoke(inv.Id(), viewerID)
		if err != ErrNotAuthorized {
			t.Errorf("expected ErrNotAuthorized, got %v", err)
		}
	})
}

func TestProcessorAccept(t *testing.T) {
	db := setupTestDB(t)
	tenantID := uuid.New()
	ownerID := uuid.New()
	hh := createTestHousehold(t, db, tenantID, "Test Home")
	createTestMembership(t, db, tenantID, hh.Id(), ownerID, "owner")

	l, _ := test.NewNullLogger()

	t.Run("accept creates membership", func(t *testing.T) {
		ctx := withTenantCtx(tenantID, ownerID)
		proc := NewProcessor(l, ctx, db)
		inv, _ := proc.Create(tenantID, hh.Id(), "accept@example.com", "editor", ownerID)

		acceptorID := uuid.New()
		result, err := proc.Accept(inv.Id(), acceptorID, "accept@example.com", tenantID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Status() != "accepted" {
			t.Errorf("expected status accepted, got %s", result.Status())
		}

		// Verify membership was created
		memProc := membership.NewProcessor(l, ctx, db)
		_, memErr := memProc.ByHouseholdAndUserProvider(hh.Id(), acceptorID)()
		if memErr != nil {
			t.Error("expected membership to exist after accept")
		}
	})

	t.Run("email mismatch rejected", func(t *testing.T) {
		ctx := withTenantCtx(tenantID, ownerID)
		proc := NewProcessor(l, ctx, db)
		inv, _ := proc.Create(tenantID, hh.Id(), "mismatch@example.com", "viewer", ownerID)

		_, err := proc.Accept(inv.Id(), uuid.New(), "wrong@example.com", tenantID)
		if err != ErrEmailMismatch {
			t.Errorf("expected ErrEmailMismatch, got %v", err)
		}
	})

	t.Run("cross-tenant rejected", func(t *testing.T) {
		ctx := withTenantCtx(tenantID, ownerID)
		proc := NewProcessor(l, ctx, db)
		inv, _ := proc.Create(tenantID, hh.Id(), "crosstenant@example.com", "viewer", ownerID)

		differentTenantID := uuid.New()
		_, err := proc.Accept(inv.Id(), uuid.New(), "crosstenant@example.com", differentTenantID)
		if err != ErrCrossTenant {
			t.Errorf("expected ErrCrossTenant, got %v", err)
		}
	})

	t.Run("expired invitation rejected", func(t *testing.T) {
		ctx := withTenantCtx(tenantID, ownerID)
		proc := NewProcessor(l, ctx, db)
		inv, _ := proc.Create(tenantID, hh.Id(), "expired@example.com", "viewer", ownerID)

		// Override timeNow to simulate expiration
		original := timeNow
		timeNow = func() time.Time { return time.Now().Add(8 * 24 * time.Hour) }
		defer func() { timeNow = original }()

		_, err := proc.Accept(inv.Id(), uuid.New(), "expired@example.com", tenantID)
		if err != ErrExpired {
			t.Errorf("expected ErrExpired, got %v", err)
		}
	})
}

func TestProcessorDecline(t *testing.T) {
	db := setupTestDB(t)
	tenantID := uuid.New()
	ownerID := uuid.New()
	hh := createTestHousehold(t, db, tenantID, "Test Home")
	createTestMembership(t, db, tenantID, hh.Id(), ownerID, "owner")

	l, _ := test.NewNullLogger()

	t.Run("decline pending invitation", func(t *testing.T) {
		ctx := withTenantCtx(tenantID, ownerID)
		proc := NewProcessor(l, ctx, db)
		inv, _ := proc.Create(tenantID, hh.Id(), "decline@example.com", "viewer", ownerID)

		result, err := proc.Decline(inv.Id(), "decline@example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Status() != "declined" {
			t.Errorf("expected status declined, got %s", result.Status())
		}
	})

	t.Run("email mismatch rejected", func(t *testing.T) {
		ctx := withTenantCtx(tenantID, ownerID)
		proc := NewProcessor(l, ctx, db)
		inv, _ := proc.Create(tenantID, hh.Id(), "declinemismatch@example.com", "viewer", ownerID)

		_, err := proc.Decline(inv.Id(), "wrong@example.com")
		if err != ErrEmailMismatch {
			t.Errorf("expected ErrEmailMismatch, got %v", err)
		}
	})

	t.Run("already declined invitation rejected", func(t *testing.T) {
		ctx := withTenantCtx(tenantID, ownerID)
		proc := NewProcessor(l, ctx, db)
		inv, _ := proc.Create(tenantID, hh.Id(), "doubledecline@example.com", "viewer", ownerID)

		proc.Decline(inv.Id(), "doubledecline@example.com")
		_, err := proc.Decline(inv.Id(), "doubledecline@example.com")
		if err != ErrNotPending {
			t.Errorf("expected ErrNotPending, got %v", err)
		}
	})
}
