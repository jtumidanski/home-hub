package planner

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func getByRecipeID(db *gorm.DB, recipeID uuid.UUID) (Entity, error) {
	var e Entity
	err := db.Where("recipe_id = ?", recipeID).First(&e).Error
	return e, err
}

func createConfig(db *gorm.DB, e *Entity) error {
	return db.Create(e).Error
}

func updateConfig(db *gorm.DB, e *Entity) error {
	e.UpdatedAt = time.Now().UTC()
	return db.Save(e).Error
}
