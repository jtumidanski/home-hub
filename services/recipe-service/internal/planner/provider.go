package planner

import (
	"github.com/google/uuid"
	database "github.com/jtumidanski/home-hub/shared/go/database"
	"gorm.io/gorm"
)

func getByRecipeID(recipeID uuid.UUID) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("recipe_id = ?", recipeID)
	})
}

func getByRecipeIDs(recipeIDs []uuid.UUID) func(db *gorm.DB) ([]Entity, error) {
	return func(db *gorm.DB) ([]Entity, error) {
		if len(recipeIDs) == 0 {
			return []Entity{}, nil
		}
		var entities []Entity
		err := db.Where("recipe_id IN (?)", recipeIDs).Find(&entities).Error
		return entities, err
	}
}
