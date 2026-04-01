package normalization

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func bulkCreate(db *gorm.DB, entities []Entity) error {
	if len(entities) == 0 {
		return nil
	}
	return db.Create(&entities).Error
}

func bulkUpdate(db *gorm.DB, entities []Entity) error {
	for i := range entities {
		entities[i].UpdatedAt = time.Now().UTC()
		if err := db.Save(&entities[i]).Error; err != nil {
			return err
		}
	}
	return nil
}

func deleteByRecipeID(db *gorm.DB, recipeID uuid.UUID) error {
	return db.Where("recipe_id = ?", recipeID).Delete(&Entity{}).Error
}

func updateOne(db *gorm.DB, e *Entity) error {
	e.UpdatedAt = time.Now().UTC()
	return db.Save(e).Error
}
