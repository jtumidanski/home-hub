package recipe

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jtumidanski/home-hub/services/recipe-service/internal/recipeimage"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
)

// ImageVisibility lets recipeimage check a recipe without importing this
// package: the recipe must exist, not be soft-deleted, and belong to the
// caller's household. Any miss is ErrRecipeNotFound (404, never 403).
func ImageVisibility(db *gorm.DB) recipeimage.RecipeVisibility {
	return func(ctx context.Context, recipeID uuid.UUID) error {
		t, ok := tenantctx.FromContext(ctx)
		if !ok {
			return recipeimage.ErrRecipeNotFound
		}
		e, err := getByID(recipeID)(db.WithContext(ctx))()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return recipeimage.ErrRecipeNotFound
		}
		if err != nil {
			return err
		}
		if e.HouseholdId != t.HouseholdId() {
			return recipeimage.ErrRecipeNotFound
		}
		return nil
	}
}
