package recipeimage

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

var (
	// ErrRecipeNotFound is returned by RecipeVisibility when the recipe does
	// not exist, is soft-deleted, or belongs to another household.
	ErrRecipeNotFound = errors.New("recipe not found")
	ErrImageNotFound  = errors.New("recipe image not found")
	// ErrObjectMissing means a DB row points at an object that is gone.
	ErrObjectMissing = errors.New("recipe image object missing")
	ErrStorage       = errors.New("object storage unavailable")
	ErrNoTenant      = errors.New("tenant missing from context")
)

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "duplicate key value") || // postgres 23505
		strings.Contains(msg, "UNIQUE constraint failed") // sqlite
}
