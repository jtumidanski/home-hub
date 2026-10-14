package recipeimage

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	database "github.com/jtumidanski/home-hub/shared/go/database"
)

// Tenant scoping comes from the GORM tenant callback; household is filtered
// explicitly on every query.

func getByRecipeID(recipeID, householdID uuid.UUID) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("recipe_id = ? AND household_id = ?", recipeID, householdID)
	})
}

func getByRecipeIDs(recipeIDs []uuid.UUID, householdID uuid.UUID) func(db *gorm.DB) ([]Entity, error) {
	return func(db *gorm.DB) ([]Entity, error) {
		if len(recipeIDs) == 0 {
			return []Entity{}, nil
		}
		var entities []Entity
		err := db.Where("recipe_id IN (?) AND household_id = ?", recipeIDs, householdID).Find(&entities).Error
		return entities, err
	}
}

func getPendingDeletions(limit int) func(db *gorm.DB) ([]ObjectDeletionEntity, error) {
	return func(db *gorm.DB) ([]ObjectDeletionEntity, error) {
		var rows []ObjectDeletionEntity
		err := db.Order("created_at ASC").Limit(limit).Find(&rows).Error
		return rows, err
	}
}

func getDeletionsByIDs(ids []uuid.UUID) func(db *gorm.DB) ([]ObjectDeletionEntity, error) {
	return func(db *gorm.DB) ([]ObjectDeletionEntity, error) {
		if len(ids) == 0 {
			return []ObjectDeletionEntity{}, nil
		}
		var rows []ObjectDeletionEntity
		err := db.Where("id IN (?)", ids).Find(&rows).Error
		return rows, err
	}
}
