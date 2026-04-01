package normalization

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/ingredient"
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
	if err := db.AutoMigrate(&Entity{}, &ingredient.Entity{}, &ingredient.AliasEntity{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func TestProcessorIntegration(t *testing.T) {
	tests := []struct {
		name        string
		ingredients []ParsedIngredient
		wantCount   int
	}{
		{
			name:        "normalize empty list",
			ingredients: []ParsedIngredient{},
			wantCount:   0,
		},
		{
			name: "normalize single ingredient",
			ingredients: []ParsedIngredient{
				{Name: "flour", Quantity: "2", Unit: "cups"},
			},
			wantCount: 1,
		},
		{
			name: "normalize multiple ingredients",
			ingredients: []ParsedIngredient{
				{Name: "flour", Quantity: "2", Unit: "cups"},
				{Name: "sugar", Quantity: "1", Unit: "cup"},
				{Name: "salt", Quantity: "1", Unit: "tsp"},
			},
			wantCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			l, _ := test.NewNullLogger()
			proc := NewProcessor(l, context.Background(), db)

			tenantID := uuid.New()
			householdID := uuid.New()
			recipeID := uuid.New()

			models, err := proc.NormalizeIngredients(tenantID, householdID, recipeID, tt.ingredients)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(models) != tt.wantCount {
				t.Errorf("got %d models, want %d", len(models), tt.wantCount)
			}

			// Verify GetByRecipeID returns the same count
			retrieved, err := proc.GetByRecipeID(recipeID)
			if err != nil {
				t.Fatalf("GetByRecipeID error: %v", err)
			}
			if len(retrieved) != tt.wantCount {
				t.Errorf("GetByRecipeID returned %d, want %d", len(retrieved), tt.wantCount)
			}
		})
	}
}
