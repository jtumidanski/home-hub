package recipe

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/audit"
	database "github.com/jtumidanski/home-hub/shared/go/database"
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
	if err := db.AutoMigrate(&Entity{}, &TagEntity{}, &RestorationEntity{}, &audit.Entity{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func TestProcessorIntegration(t *testing.T) {
	tenantID := uuid.New()
	householdID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name string
		fn   func(t *testing.T, db *gorm.DB)
	}{
		{
			name: "create and get recipe",
			fn: func(t *testing.T, db *gorm.DB) {
				ctx := tenantctx.WithContext(context.Background(), tenantctx.New(tenantID, householdID, userID))
				l, _ := test.NewNullLogger()
				p := NewProcessor(l, ctx, db)

				attrs := CreateAttrs{
					Title:  "Test Recipe",
					Source: "Boil water.",
				}

				created, _, err := p.Create(tenantID, householdID, attrs)
				if err != nil {
					t.Fatalf("unexpected error creating recipe: %v", err)
				}
				if created.Title() != "Test Recipe" {
					t.Fatalf("expected title %q, got %q", "Test Recipe", created.Title())
				}

				got, _, err := p.Get(created.Id())
				if err != nil {
					t.Fatalf("unexpected error getting recipe: %v", err)
				}
				if got.Title() != created.Title() {
					t.Fatalf("expected title %q, got %q", created.Title(), got.Title())
				}
			},
		},
		{
			name: "create with missing title returns error",
			fn: func(t *testing.T, db *gorm.DB) {
				ctx := tenantctx.WithContext(context.Background(), tenantctx.New(tenantID, householdID, userID))
				l, _ := test.NewNullLogger()
				p := NewProcessor(l, ctx, db)

				attrs := CreateAttrs{
					Title:  "",
					Source: "Boil water.",
				}

				_, _, err := p.Create(tenantID, householdID, attrs)
				if err == nil {
					t.Fatal("expected error for missing title, got nil")
				}
				if !errors.Is(err, ErrTitleRequired) {
					t.Fatalf("expected ErrTitleRequired, got %v", err)
				}
			},
		},
		{
			name: "get nonexistent returns error",
			fn: func(t *testing.T, db *gorm.DB) {
				ctx := tenantctx.WithContext(context.Background(), tenantctx.New(tenantID, householdID, userID))
				l, _ := test.NewNullLogger()
				p := NewProcessor(l, ctx, db)

				_, _, err := p.Get(uuid.New())
				if err == nil {
					t.Fatal("expected error for nonexistent recipe, got nil")
				}
				if !errors.Is(err, ErrNotFound) {
					t.Fatalf("expected ErrNotFound, got %v", err)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db := setupTestDB(t)
			tc.fn(t, db)
		})
	}
}
