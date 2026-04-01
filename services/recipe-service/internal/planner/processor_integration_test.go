package planner

import (
	"context"
	"testing"

	"github.com/google/uuid"
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
	if err := db.AutoMigrate(&Entity{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func strPtr(s string) *string { return &s }

func TestProcessorIntegration(t *testing.T) {
	tests := []struct {
		name               string
		classification     *string
		servingsYield      *int
		wantClassification string
	}{
		{
			name:               "create config with classification",
			classification:     strPtr("dinner"),
			servingsYield:      ptrInt(4),
			wantClassification: "dinner",
		},
		{
			name:               "create config without classification",
			servingsYield:      ptrInt(2),
			wantClassification: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			l, _ := test.NewNullLogger()
			proc := NewProcessor(l, context.Background(), db)

			recipeID := uuid.New()
			m, err := proc.CreateOrUpdate(recipeID, ConfigAttrs{
				Classification: tt.classification,
				ServingsYield:  tt.servingsYield,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if m.Classification() != tt.wantClassification {
				t.Errorf("Classification() = %q, want %q", m.Classification(), tt.wantClassification)
			}

			// Verify update path
			updated, err := proc.CreateOrUpdate(recipeID, ConfigAttrs{
				Classification: strPtr("lunch"),
				ServingsYield:  tt.servingsYield,
			})
			if err != nil {
				t.Fatalf("update error: %v", err)
			}
			if updated.Classification() != "lunch" {
				t.Errorf("after update Classification() = %q, want %q", updated.Classification(), "lunch")
			}
		})
	}
}
