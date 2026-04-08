package ingredient

import (
	"context"
	"testing"
	"time"

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
	if err := db.AutoMigrate(&Entity{}, &AliasEntity{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func TestProcessorIntegration(t *testing.T) {
	tests := []struct {
		name        string
		createName  string
		displayName string
		unitFamily  string
		wantErr     bool
		wantErrIs   error
	}{
		{
			name:        "create and retrieve ingredient",
			createName:  "flour",
			displayName: "All-Purpose Flour",
			unitFamily:  "weight",
		},
		{
			name:      "empty name returns error",
			wantErr:   true,
			wantErrIs: ErrNameRequired,
		},
		{
			name:       "invalid unit family returns error",
			createName: "sugar",
			unitFamily: "invalid",
			wantErr:    true,
			wantErrIs:  ErrInvalidUnitFamily,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			l, _ := test.NewNullLogger()
			proc := NewProcessor(l, context.Background(), db)

			m, err := proc.Create(uuid.New(), tt.createName, tt.displayName, tt.unitFamily, nil)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrIs != nil && err.Error() != tt.wantErrIs.Error() {
					t.Errorf("expected error %q, got %q", tt.wantErrIs, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got, err := proc.Get(m.Id())
			if err != nil {
				t.Fatalf("failed to get ingredient: %v", err)
			}
			if got.Name() != tt.createName {
				t.Errorf("Name() = %q, want %q", got.Name(), tt.createName)
			}
		})
	}
}

func TestGetIngredientRecipesIncludesRecipeName(t *testing.T) {
	db := setupTestDB(t)
	if err := db.Exec(`CREATE TABLE recipes (
		id text primary key,
		title text not null,
		deleted_at datetime,
		created_at datetime not null,
		updated_at datetime not null
	)`).Error; err != nil {
		t.Fatalf("failed to create recipes table: %v", err)
	}
	if err := db.Exec(`CREATE TABLE recipe_ingredients (
		id text primary key,
		recipe_id text not null,
		canonical_ingredient_id text,
		raw_name text not null,
		created_at datetime not null,
		updated_at datetime not null
	)`).Error; err != nil {
		t.Fatalf("failed to create recipe_ingredients table: %v", err)
	}

	canonID := uuid.New()
	now := time.Now().UTC()
	recipe1ID := uuid.New()
	recipe2ID := uuid.New()
	if err := db.Exec(
		`INSERT INTO recipes (id, title, created_at, updated_at) VALUES (?, ?, ?, ?), (?, ?, ?, ?)`,
		recipe1ID.String(), "Pancakes", now, now,
		recipe2ID.String(), "Waffles", now, now,
	).Error; err != nil {
		t.Fatalf("failed to insert recipes: %v", err)
	}
	if err := db.Exec(
		`INSERT INTO recipe_ingredients (id, recipe_id, canonical_ingredient_id, raw_name, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?), (?, ?, ?, ?, ?, ?)`,
		uuid.New().String(), recipe1ID.String(), canonID.String(), "1 cup flour", now, now,
		uuid.New().String(), recipe2ID.String(), canonID.String(), "2 cups flour", now, now.Add(time.Second),
	).Error; err != nil {
		t.Fatalf("failed to insert recipe_ingredients: %v", err)
	}

	refs, total, err := getIngredientRecipes(db, canonID, 1, 10)
	if err != nil {
		t.Fatalf("getIngredientRecipes error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected 2 results, got %d", total)
	}
	if len(refs) != 2 {
		t.Fatalf("expected 2 refs, got %d", len(refs))
	}
	names := map[string]bool{refs[0].RecipeName: true, refs[1].RecipeName: true}
	if !names["Pancakes"] || !names["Waffles"] {
		t.Errorf("expected recipe names Pancakes and Waffles, got %+v", names)
	}

	// Soft-deleted recipe should be excluded.
	if err := db.Exec(`UPDATE recipes SET deleted_at = ? WHERE id = ?`, now, recipe2ID.String()).Error; err != nil {
		t.Fatalf("failed to soft-delete recipe: %v", err)
	}
	refs, total, err = getIngredientRecipes(db, canonID, 1, 10)
	if err != nil {
		t.Fatalf("getIngredientRecipes error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1 result after soft-delete, got %d", total)
	}
	if len(refs) != 1 || refs[0].RecipeName != "Pancakes" {
		t.Errorf("expected only Pancakes, got %+v", refs)
	}
}
