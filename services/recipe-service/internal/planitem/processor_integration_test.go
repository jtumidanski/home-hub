package planitem

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/audit"
	database "github.com/jtumidanski/home-hub/shared/go/database"
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
	if err := db.AutoMigrate(&Entity{}, &audit.Entity{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func TestProcessorIntegration(t *testing.T) {
	tests := []struct {
		name    string
		slot    string
		wantErr bool
	}{
		{
			name: "add and retrieve item",
			slot: "dinner",
		},
		{
			name:    "invalid slot returns error",
			slot:    "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			l, _ := test.NewNullLogger()
			proc := NewProcessor(l, context.Background(), db)

			planWeekID := uuid.New()
			startsOn := time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)
			day := startsOn

			m, err := proc.AddItem(planWeekID, startsOn, AddAttrs{
				Day:      day,
				Slot:     tt.slot,
				RecipeID: uuid.New(),
			})
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			items, err := proc.GetByPlanWeekID(planWeekID)
			if err != nil {
				t.Fatalf("GetByPlanWeekID error: %v", err)
			}
			if len(items) != 1 {
				t.Fatalf("expected 1 item, got %d", len(items))
			}
			if items[0].Id() != m.Id() {
				t.Errorf("item ID mismatch: got %s, want %s", items[0].Id(), m.Id())
			}
		})
	}
}
