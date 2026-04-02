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
	tests := []struct {
		name      string
		email     string
		role      string
		inviterFn func(tenantID uuid.UUID, householdID uuid.UUID, db *gorm.DB, t *testing.T) uuid.UUID
		setup     func(t *testing.T, db *gorm.DB, tenantID uuid.UUID, householdID uuid.UUID, ownerID uuid.UUID)
		wantErr   error
		wantEmail string
		wantRole  string
	}{
		{
			name:  "create invitation",
			email: "invitee@example.com",
			role:  "viewer",
			inviterFn: func(tenantID uuid.UUID, householdID uuid.UUID, db *gorm.DB, t *testing.T) uuid.UUID {
				ownerID := uuid.New()
				createTestMembership(t, db, tenantID, householdID, ownerID, "owner")
				return ownerID
			},
			wantEmail: "invitee@example.com",
			wantRole:  "viewer",
		},
		{
			name:  "duplicate invitation returns error",
			email: "dup@example.com",
			role:  "editor",
			inviterFn: func(tenantID uuid.UUID, householdID uuid.UUID, db *gorm.DB, t *testing.T) uuid.UUID {
				ownerID := uuid.New()
				createTestMembership(t, db, tenantID, householdID, ownerID, "owner")
				return ownerID
			},
			setup: func(t *testing.T, db *gorm.DB, tenantID uuid.UUID, householdID uuid.UUID, ownerID uuid.UUID) {
				l, _ := test.NewNullLogger()
				ctx := withTenantCtx(tenantID, ownerID)
				proc := NewProcessor(l, ctx, db)
				proc.Create(tenantID, householdID, "dup@example.com", "viewer", ownerID)
			},
			wantErr: ErrAlreadyInvited,
		},
		{
			name:  "non-privileged user cannot create",
			email: "another@example.com",
			role:  "viewer",
			inviterFn: func(tenantID uuid.UUID, householdID uuid.UUID, db *gorm.DB, t *testing.T) uuid.UUID {
				ownerID := uuid.New()
				createTestMembership(t, db, tenantID, householdID, ownerID, "owner")
				viewerID := uuid.New()
				createTestMembership(t, db, tenantID, householdID, viewerID, "viewer")
				return viewerID
			},
			wantErr: ErrNotAuthorized,
		},
		{
			name:  "default role to viewer",
			email: "defaultrole@example.com",
			role:  "",
			inviterFn: func(tenantID uuid.UUID, householdID uuid.UUID, db *gorm.DB, t *testing.T) uuid.UUID {
				ownerID := uuid.New()
				createTestMembership(t, db, tenantID, householdID, ownerID, "owner")
				return ownerID
			},
			wantEmail: "defaultrole@example.com",
			wantRole:  "viewer",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			tenantID := uuid.New()
			hh := createTestHousehold(t, db, tenantID, "Test Home")
			inviterID := tt.inviterFn(tenantID, hh.Id(), db, t)

			if tt.setup != nil {
				tt.setup(t, db, tenantID, hh.Id(), inviterID)
			}

			l, _ := test.NewNullLogger()
			ctx := withTenantCtx(tenantID, inviterID)
			proc := NewProcessor(l, ctx, db)
			m, err := proc.Create(tenantID, hh.Id(), tt.email, tt.role, inviterID)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("expected %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if m.Email() != tt.wantEmail {
				t.Errorf("expected email %s, got %s", tt.wantEmail, m.Email())
			}
			if m.Role() != tt.wantRole {
				t.Errorf("expected role %s, got %s", tt.wantRole, m.Role())
			}
			if m.Status() != "pending" {
				t.Errorf("expected status pending, got %s", m.Status())
			}
		})
	}
}

func TestProcessorRevoke(t *testing.T) {
	tests := []struct {
		name      string
		revokerFn func(tenantID uuid.UUID, householdID uuid.UUID, db *gorm.DB, t *testing.T) uuid.UUID
		wantErr   error
	}{
		{
			name: "revoke pending invitation",
			revokerFn: func(tenantID uuid.UUID, householdID uuid.UUID, db *gorm.DB, t *testing.T) uuid.UUID {
				ownerID := uuid.New()
				createTestMembership(t, db, tenantID, householdID, ownerID, "owner")
				return ownerID
			},
		},
		{
			name: "non-privileged user cannot revoke",
			revokerFn: func(tenantID uuid.UUID, householdID uuid.UUID, db *gorm.DB, t *testing.T) uuid.UUID {
				ownerID := uuid.New()
				createTestMembership(t, db, tenantID, householdID, ownerID, "owner")
				viewerID := uuid.New()
				createTestMembership(t, db, tenantID, householdID, viewerID, "viewer")
				return viewerID
			},
			wantErr: ErrNotAuthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			tenantID := uuid.New()
			hh := createTestHousehold(t, db, tenantID, "Test Home")

			// Create invitation as owner
			ownerID := uuid.New()
			createTestMembership(t, db, tenantID, hh.Id(), ownerID, "owner")
			l, _ := test.NewNullLogger()
			ctx := withTenantCtx(tenantID, ownerID)
			proc := NewProcessor(l, ctx, db)
			inv, _ := proc.Create(tenantID, hh.Id(), "revoke@example.com", "viewer", ownerID)

			// Attempt revoke as test subject
			revokerID := tt.revokerFn(tenantID, hh.Id(), db, t)
			ctx2 := withTenantCtx(tenantID, revokerID)
			proc2 := NewProcessor(l, ctx2, db)
			err := proc2.Revoke(inv.Id(), revokerID)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("expected %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			updated, _ := proc.ByIDProvider(inv.Id())()
			if updated.Status() != "revoked" {
				t.Errorf("expected status revoked, got %s", updated.Status())
			}
		})
	}
}

func TestProcessorAccept(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		acceptEmail string
		tenantID    func(invTenantID uuid.UUID) uuid.UUID
		timeOverride func()
		timeRestore  func()
		wantErr     error
		wantStatus  string
	}{
		{
			name:        "accept creates membership",
			email:       "accept@example.com",
			acceptEmail: "accept@example.com",
			tenantID:    func(id uuid.UUID) uuid.UUID { return id },
			wantStatus:  "accepted",
		},
		{
			name:        "email mismatch rejected",
			email:       "mismatch@example.com",
			acceptEmail: "wrong@example.com",
			tenantID:    func(id uuid.UUID) uuid.UUID { return id },
			wantErr:     ErrEmailMismatch,
		},
		{
			name:        "cross-tenant rejected",
			email:       "crosstenant@example.com",
			acceptEmail: "crosstenant@example.com",
			tenantID:    func(_ uuid.UUID) uuid.UUID { return uuid.New() },
			wantErr:     ErrCrossTenant,
		},
		{
			name:        "expired invitation rejected",
			email:       "expired@example.com",
			acceptEmail: "expired@example.com",
			tenantID:    func(id uuid.UUID) uuid.UUID { return id },
			timeOverride: func() {
				timeNow = func() time.Time { return time.Now().Add(8 * 24 * time.Hour) }
			},
			timeRestore: func() {
				timeNow = func() time.Time { return time.Now().UTC() }
			},
			wantErr: ErrExpired,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			tenantID := uuid.New()
			ownerID := uuid.New()
			hh := createTestHousehold(t, db, tenantID, "Test Home")
			createTestMembership(t, db, tenantID, hh.Id(), ownerID, "owner")

			l, _ := test.NewNullLogger()
			ctx := withTenantCtx(tenantID, ownerID)
			proc := NewProcessor(l, ctx, db)
			inv, _ := proc.Create(tenantID, hh.Id(), tt.email, "editor", ownerID)

			if tt.timeOverride != nil {
				tt.timeOverride()
				defer tt.timeRestore()
			}

			acceptorID := uuid.New()
			result, err := proc.Accept(inv.Id(), acceptorID, tt.acceptEmail, tt.tenantID(tenantID))
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("expected %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Status() != tt.wantStatus {
				t.Errorf("expected status %s, got %s", tt.wantStatus, result.Status())
			}

			memProc := membership.NewProcessor(l, ctx, db)
			_, memErr := memProc.ByHouseholdAndUserProvider(hh.Id(), acceptorID)()
			if memErr != nil {
				t.Error("expected membership to exist after accept")
			}
		})
	}
}

func TestProcessorDecline(t *testing.T) {
	tests := []struct {
		name         string
		email        string
		declineEmail string
		preDecline   bool
		wantErr      error
		wantStatus   string
	}{
		{
			name:         "decline pending invitation",
			email:        "decline@example.com",
			declineEmail: "decline@example.com",
			wantStatus:   "declined",
		},
		{
			name:         "email mismatch rejected",
			email:        "declinemismatch@example.com",
			declineEmail: "wrong@example.com",
			wantErr:      ErrEmailMismatch,
		},
		{
			name:         "already declined invitation rejected",
			email:        "doubledecline@example.com",
			declineEmail: "doubledecline@example.com",
			preDecline:   true,
			wantErr:      ErrNotPending,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			tenantID := uuid.New()
			ownerID := uuid.New()
			hh := createTestHousehold(t, db, tenantID, "Test Home")
			createTestMembership(t, db, tenantID, hh.Id(), ownerID, "owner")

			l, _ := test.NewNullLogger()
			ctx := withTenantCtx(tenantID, ownerID)
			proc := NewProcessor(l, ctx, db)
			inv, _ := proc.Create(tenantID, hh.Id(), tt.email, "viewer", ownerID)

			if tt.preDecline {
				proc.Decline(inv.Id(), tt.email)
			}

			result, err := proc.Decline(inv.Id(), tt.declineEmail)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("expected %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Status() != tt.wantStatus {
				t.Errorf("expected status %s, got %s", tt.wantStatus, result.Status())
			}
		})
	}
}
