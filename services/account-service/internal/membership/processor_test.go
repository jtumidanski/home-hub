package membership

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

func TestProcessor(t *testing.T) {
	t.Run("Create", func(t *testing.T) {
		tests := []struct {
			name     string
			role     string
			wantRole string
		}{
			{"owner role", "owner", "owner"},
			{"admin role", "admin", "admin"},
			{"editor role", "editor", "editor"},
			{"viewer role", "viewer", "viewer"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db := setupTestDB(t)
				l, _ := test.NewNullLogger()
				p := NewProcessor(l, context.Background(), db)

				householdID := uuid.New()
				m, err := p.Create(uuid.New(), householdID, uuid.New(), tt.role)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if m.Role() != tt.wantRole {
					t.Errorf("expected role %s, got %s", tt.wantRole, m.Role())
				}
				if m.HouseholdID() != householdID {
					t.Errorf("expected household ID %s, got %s", householdID, m.HouseholdID())
				}
			})
		}
	})

	t.Run("UpdateRole", func(t *testing.T) {
		tests := []struct {
			name        string
			initialRole string
			newRole     string
		}{
			{"viewer to admin", "viewer", "admin"},
			{"admin to owner", "admin", "owner"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db := setupTestDB(t)
				l, _ := test.NewNullLogger()
				p := NewProcessor(l, context.Background(), db)

				m, _ := p.Create(uuid.New(), uuid.New(), uuid.New(), tt.initialRole)
				updated, err := p.UpdateRole(m.Id(), tt.newRole)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if updated.Role() != tt.newRole {
					t.Errorf("expected role %s, got %s", tt.newRole, updated.Role())
				}
			})
		}
	})

	t.Run("Delete", func(t *testing.T) {
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
	})

	t.Run("ByUserProvider", func(t *testing.T) {
		db := setupTestDB(t)
		l, _ := test.NewNullLogger()
		p := NewProcessor(l, context.Background(), db)

		tenantID := uuid.New()
		userID := uuid.New()
		p.Create(tenantID, uuid.New(), userID, "owner")
		p.Create(tenantID, uuid.New(), userID, "editor")
		p.Create(uuid.New(), uuid.New(), uuid.New(), "viewer") // different user

		models, err := p.ByUserProvider(userID)()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(models) != 2 {
			t.Errorf("expected 2 memberships, got %d", len(models))
		}
	})
}
