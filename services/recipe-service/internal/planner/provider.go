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

func upsert(db *gorm.DB, recipeID uuid.UUID, classification *string, servingsYield, eatWithinDays, minGapDays, maxConsecutiveDays *int) (Entity, error) {
	var existing Entity
	err := db.Where("recipe_id = ?", recipeID).First(&existing).Error

	now := time.Now().UTC()
	if err == gorm.ErrRecordNotFound {
		e := Entity{
			Id:                 uuid.New(),
			RecipeId:           recipeID,
			Classification:     classification,
			ServingsYield:      servingsYield,
			EatWithinDays:      eatWithinDays,
			MinGapDays:         minGapDays,
			MaxConsecutiveDays: maxConsecutiveDays,
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		if err := db.Create(&e).Error; err != nil {
			return Entity{}, err
		}
		return e, nil
	}
	if err != nil {
		return Entity{}, err
	}

	existing.Classification = classification
	existing.ServingsYield = servingsYield
	existing.EatWithinDays = eatWithinDays
	existing.MinGapDays = minGapDays
	existing.MaxConsecutiveDays = maxConsecutiveDays
	existing.UpdatedAt = now
	if err := db.Save(&existing).Error; err != nil {
		return Entity{}, err
	}
	return existing, nil
}
