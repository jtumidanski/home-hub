package appcontext

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/account-service/internal/household"
	"github.com/jtumidanski/home-hub/services/account-service/internal/invitation"
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
	db.AutoMigrate(&invitation.Entity{})
	return db
}

func TestResolve(t *testing.T) {
	tests := []struct {
		name               string
		setup              func(t *testing.T, db *gorm.DB) (tenantID, userID uuid.UUID)
		wantErr            bool
		wantActiveHH       bool
		wantRole           string
		wantCanCreate      bool
		wantMembershipCount int
	}{
		{
			name: "happy path with owner membership",
			setup: func(t *testing.T, db *gorm.DB) (uuid.UUID, uuid.UUID) {
				l, _ := test.NewNullLogger()
				ctx := context.Background()
				tenantProc := tenant.NewProcessor(l, ctx, db)
				ten, err := tenantProc.Create("Test Tenant")
				if err != nil {
					t.Fatalf("failed to create tenant: %v", err)
				}
				userID := uuid.New()
				hhProc := household.NewProcessor(l, ctx, db)
				if _, err := hhProc.CreateWithOwner(ten.Id(), userID, "Test Home", "UTC", "metric"); err != nil {
					t.Fatalf("failed to create household: %v", err)
				}
				return ten.Id(), userID
			},
			wantActiveHH:        true,
			wantRole:            "owner",
			wantCanCreate:       true,
			wantMembershipCount: 1,
		},
		{
			name: "no memberships",
			setup: func(t *testing.T, db *gorm.DB) (uuid.UUID, uuid.UUID) {
				l, _ := test.NewNullLogger()
				ctx := context.Background()
				tenantProc := tenant.NewProcessor(l, ctx, db)
				ten, err := tenantProc.Create("Empty Tenant")
				if err != nil {
					t.Fatalf("failed to create tenant: %v", err)
				}
				return ten.Id(), uuid.New()
			},
			wantActiveHH:        false,
			wantRole:            "",
			wantCanCreate:       false,
			wantMembershipCount: 0,
		},
		{
			name: "resolves first household when no active set",
			setup: func(t *testing.T, db *gorm.DB) (uuid.UUID, uuid.UUID) {
				l, _ := test.NewNullLogger()
				ctx := context.Background()
				tenantProc := tenant.NewProcessor(l, ctx, db)
				ten, err := tenantProc.Create("Resolve Tenant")
				if err != nil {
					t.Fatalf("failed to create tenant: %v", err)
				}
				userID := uuid.New()
				hhProc := household.NewProcessor(l, ctx, db)
				if _, err := hhProc.CreateWithOwner(ten.Id(), userID, "First Home", "UTC", "metric"); err != nil {
					t.Fatalf("failed to create household: %v", err)
				}
				return ten.Id(), userID
			},
			wantActiveHH:        true,
			wantRole:            "owner",
			wantCanCreate:       true,
			wantMembershipCount: 1,
		},
		{
			name: "tenant not found",
			setup: func(_ *testing.T, _ *gorm.DB) (uuid.UUID, uuid.UUID) {
				return uuid.New(), uuid.New()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			l, _ := test.NewNullLogger()
			ctx := context.Background()

			tenantID, userID := tt.setup(t, db)
			resolved, err := Resolve(l, ctx, db, tenantID, userID, "")

			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantActiveHH && resolved.ActiveHousehold == nil {
				t.Error("expected active household")
			}
			if !tt.wantActiveHH && resolved.ActiveHousehold != nil {
				t.Error("expected nil active household")
			}
			if resolved.ResolvedRole != tt.wantRole {
				t.Errorf("expected role %q, got %q", tt.wantRole, resolved.ResolvedRole)
			}
			if resolved.CanCreateHousehold != tt.wantCanCreate {
				t.Errorf("expected canCreateHousehold=%v, got %v", tt.wantCanCreate, resolved.CanCreateHousehold)
			}
			if len(resolved.Memberships) != tt.wantMembershipCount {
				t.Errorf("expected %d memberships, got %d", tt.wantMembershipCount, len(resolved.Memberships))
			}
		})
	}
}
