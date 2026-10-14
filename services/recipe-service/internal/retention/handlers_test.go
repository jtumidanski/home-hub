package retention

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/jtumidanski/home-hub/services/recipe-service/internal/normalization"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/planitem"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/recipe"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/recipeimage"
	sr "github.com/jtumidanski/home-hub/shared/go/retention"
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
		&recipeimage.Entity{},
		&recipeimage.ObjectDeletionEntity{},
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

	oldID := mkRecipe(&past) // past 30-day window — should be reaped
	mkRecipe(&recent)        // inside window — should survive
	mkRecipe(nil)            // not deleted — should survive

	// Cascade targets for the old recipe.
	db.Create(&recipe.TagEntity{Id: uuid.New(), RecipeId: oldID, Tag: "dinner"})
	db.Create(&normalization.Entity{
		Id: uuid.New(), TenantId: tenantID, HouseholdId: householdID,
		RecipeId: oldID, RawName: "flour", Position: 1,
		NormalizationStatus: "unresolved",
		CreatedAt:           now, UpdatedAt: now,
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

func seedPurgeable(t *testing.T, db *gorm.DB) (tenantID, householdID, recipeID uuid.UUID) {
	t.Helper()
	tenantID, householdID, recipeID = uuid.New(), uuid.New(), uuid.New()
	now := time.Now()
	past := now.Add(-31 * 24 * time.Hour)
	db.Create(&recipe.Entity{
		Id: recipeID, TenantId: tenantID, HouseholdId: householdID,
		Title: "Old", Source: "manual", DeletedAt: &past, CreatedAt: now, UpdatedAt: now,
	})
	db.Create(&recipeimage.Entity{
		Id: uuid.New(), TenantId: tenantID, HouseholdId: householdID, RecipeId: recipeID,
		FullKey: "k/full.jpg", ThumbKey: "k/thumb.jpg", ContentType: "image/jpeg",
		Width: 10, Height: 10, FullBytes: 1, ThumbBytes: 1, CreatedAt: now, UpdatedAt: now,
	})
	return
}

func TestDeletedRecipesRestoreWindowPurgesImages(t *testing.T) {
	db := newDB(t)
	tid, hid, _ := seedPurgeable(t, db)
	scope := sr.Scope{TenantId: tid, Kind: sr.ScopeHousehold, ScopeId: hid}

	res, err := DeletedRecipesRestoreWindow{}.Reap(context.Background(), db, scope, 30, false)
	if err != nil {
		t.Fatal(err)
	}
	if res.Deleted != 2 { // 1 image row + 1 recipe row
		t.Fatalf("Deleted = %d, want 2", res.Deleted)
	}
	var images, outbox int64
	db.Model(&recipeimage.Entity{}).Count(&images)
	db.Model(&recipeimage.ObjectDeletionEntity{}).Count(&outbox)
	if images != 0 || outbox != 2 {
		t.Fatalf("images=%d outbox=%d, want 0 and 2", images, outbox)
	}
}

func TestDeletedRecipesRestoreWindowDryRunLeavesImages(t *testing.T) {
	db := newDB(t)
	tid, hid, _ := seedPurgeable(t, db)
	scope := sr.Scope{TenantId: tid, Kind: sr.ScopeHousehold, ScopeId: hid}
	rollback := errors.New("dry run")

	// The shared reaper implements dry-run by rolling back the tx.
	err := db.Transaction(func(tx *gorm.DB) error {
		res, err := DeletedRecipesRestoreWindow{}.Reap(context.Background(), tx, scope, 30, true)
		if err != nil {
			return err
		}
		if res.Deleted != 2 {
			t.Fatalf("dry-run Deleted = %d, want 2 (counts images)", res.Deleted)
		}
		return rollback
	})
	if !errors.Is(err, rollback) {
		t.Fatal(err)
	}
	var images, outbox int64
	db.Model(&recipeimage.Entity{}).Count(&images)
	db.Model(&recipeimage.ObjectDeletionEntity{}).Count(&outbox)
	if images != 1 || outbox != 0 {
		t.Fatalf("images=%d outbox=%d, want 1 and 0", images, outbox)
	}
}

func TestDeletedRecipesRestoreWindowLeavesOtherTenantImages(t *testing.T) {
	db := newDB(t)
	tidA, hidA, _ := seedPurgeable(t, db)
	_, _, recipeB := seedPurgeable(t, db) // tenant B: equally purgeable, different scope
	scope := sr.Scope{TenantId: tidA, Kind: sr.ScopeHousehold, ScopeId: hidA}

	if _, err := (DeletedRecipesRestoreWindow{}).Reap(context.Background(), db, scope, 30, false); err != nil {
		t.Fatal(err)
	}
	var imagesB, recipesB, outbox int64
	db.Model(&recipeimage.Entity{}).Where("recipe_id = ?", recipeB).Count(&imagesB)
	db.Model(&recipe.Entity{}).Where("id = ?", recipeB).Count(&recipesB)
	db.Model(&recipeimage.ObjectDeletionEntity{}).Count(&outbox)
	if imagesB != 1 || recipesB != 1 {
		t.Fatalf("tenant B imagesB=%d recipesB=%d, want 1 and 1", imagesB, recipesB)
	}
	if outbox != 2 {
		t.Fatalf("outbox=%d, want 2 (tenant A only)", outbox)
	}
}
