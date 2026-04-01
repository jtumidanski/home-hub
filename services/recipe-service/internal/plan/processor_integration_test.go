package plan

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
		name     string
		startsOn time.Time
		planName string
		wantErr  bool
	}{
		{
			name:     "create and get plan",
			startsOn: time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC),
			planName: "Test Week",
		},
		{
			name:    "missing starts_on returns error",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			l, _ := test.NewNullLogger()
			proc := NewProcessor(l, context.Background(), db)

			m, err := proc.Create(uuid.New(), uuid.New(), uuid.New(), CreateAttrs{
				StartsOn: tt.startsOn,
				Name:     tt.planName,
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

			got, err := proc.Get(m.Id())
			if err != nil {
				t.Fatalf("failed to get plan: %v", err)
			}
			if got.Name() != tt.planName {
				t.Errorf("Name() = %q, want %q", got.Name(), tt.planName)
			}
		})
	}
}
