package retention

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/normalization"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/planitem"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/recipe"
	sr "github.com/jtumidanski/home-hub/shared/go/retention"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(
		&recipe.Entity{},
		&recipe.TagEntity{},
		&recipe.RestorationEntity{},
		&normalization.Entity{},
		&planitem.Entity{},
		&sr.RunEntity{},
	); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestDeletedRecipesRestoreWindowBoundary(t *testing.T) {
	db := newDB(t)
	tenantID := uuid.New()
	householdID := uuid.New()
	now := time.Now()

	past := now.Add(-31 * 24 * time.Hour)
	recent := now.Add(-10 * 24 * time.Hour)

	mkRecipe := func(deletedAt *time.Time) uuid.UUID {
		id := uuid.New()
		db.Create(&recipe.Entity{
			Id: id, TenantId: tenantID, HouseholdId: householdID,
			Title: "Test Recipe", Source: "manual",
			DeletedAt: deletedAt, CreatedAt: now, UpdatedAt: now,
		})
		return id
	}

	oldID := mkRecipe(&past)   // past 30-day window — should be reaped
	mkRecipe(&recent)          // inside window — should survive
	mkRecipe(nil)              // not deleted — should survive

	// Cascade targets for the old recipe.
	db.Create(&recipe.TagEntity{Id: uuid.New(), RecipeId: oldID, Tag: "dinner"})
	db.Create(&normalization.Entity{
		Id: uuid.New(), TenantId: tenantID, HouseholdId: householdID,
		RecipeId: oldID, RawName: "flour", Position: 1,
		NormalizationStatus: "unresolved",
		CreatedAt: now, UpdatedAt: now,
	})
	db.Create(&recipe.RestorationEntity{Id: uuid.New(), RecipeId: oldID, RestoredAt: now})
	db.Create(&planitem.Entity{
		Id: uuid.New(), PlanWeekId: uuid.New(), Day: now,
		Slot: "dinner", RecipeId: oldID, Position: 0,
		CreatedAt: now, UpdatedAt: now,
	})

	scope := sr.Scope{TenantId: tenantID, Kind: sr.ScopeHousehold, ScopeId: householdID}
	res, err := DeletedRecipesRestoreWindow{}.Reap(context.Background(), db, scope, 30, false)
	if err != nil {
		t.Fatal(err)
	}
	if res.Scanned != 1 {
		t.Errorf("scanned = %d, want 1", res.Scanned)
	}
	// 1 tag + 1 ingredient + 1 restoration + 1 plan_item + 1 recipe = 5
	if res.Deleted != 5 {
		t.Errorf("deleted = %d, want 5", res.Deleted)
	}

	var recipeCount int64
	db.Model(&recipe.Entity{}).Count(&recipeCount)
	if recipeCount != 2 {
		t.Errorf("remaining recipes = %d, want 2", recipeCount)
	}
	var ingredientCount int64
	db.Model(&normalization.Entity{}).Count(&ingredientCount)
	if ingredientCount != 0 {
		t.Errorf("remaining recipe_ingredients = %d, want 0", ingredientCount)
	}
}

func TestRestorationAuditBoundary(t *testing.T) {
	db := newDB(t)
	tenantID := uuid.New()
	householdID := uuid.New()
	now := time.Now()

	recipeID := uuid.New()
	db.Create(&recipe.Entity{
		Id: recipeID, TenantId: tenantID, HouseholdId: householdID,
		Title: "Test", Source: "manual", CreatedAt: now, UpdatedAt: now,
	})

	old := now.Add(-100 * 24 * time.Hour)
	recent := now.Add(-30 * 24 * time.Hour)
	db.Create(&recipe.RestorationEntity{Id: uuid.New(), RecipeId: recipeID, RestoredAt: old})
	db.Create(&recipe.RestorationEntity{Id: uuid.New(), RecipeId: recipeID, RestoredAt: recent})

	scope := sr.Scope{TenantId: tenantID, Kind: sr.ScopeHousehold, ScopeId: householdID}
	res, err := RestorationAudit{}.Reap(context.Background(), db, scope, 90, false)
	if err != nil {
		t.Fatal(err)
	}
	if res.Deleted != 1 {
		t.Errorf("deleted = %d, want 1", res.Deleted)
	}

	var remaining int64
	db.Model(&recipe.RestorationEntity{}).Count(&remaining)
	if remaining != 1 {
		t.Errorf("remaining restorations = %d, want 1", remaining)
	}
}
