package ingredient

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GetByMealId retrieves all ingredients for a meal
func GetByMealId(ctx context.Context, db *gorm.DB, mealId uuid.UUID) ([]Model, error) {
	var entities []Entity
	if err := db.WithContext(ctx).
		Where("meal_id = ?", mealId).
		Order("created_at ASC").
		Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("failed to get ingredients: %w", err)
	}

	models := make([]Model, len(entities))
	for i, entity := range entities {
		model, err := Make(entity)
		if err != nil {
			return nil, fmt.Errorf("failed to convert entity to model: %w", err)
		}
		models[i] = model
	}
	return models, nil
}

// Save persists an ingredient to the database
func Save(ctx context.Context, db *gorm.DB, m Model) error {
	entity := m.ToEntity()
	if err := db.WithContext(ctx).Save(&entity).Error; err != nil {
		return fmt.Errorf("failed to save ingredient: %w", err)
	}
	return nil
}

// SaveBatch persists multiple ingredients to the database
func SaveBatch(ctx context.Context, db *gorm.DB, models []Model) error {
	if len(models) == 0 {
		return nil
	}

	entities := make([]Entity, len(models))
	for i, model := range models {
		entities[i] = model.ToEntity()
	}

	if err := db.WithContext(ctx).Create(&entities).Error; err != nil {
		return fmt.Errorf("failed to save ingredients batch: %w", err)
	}
	return nil
}

// DeleteByMealId deletes all ingredients for a meal
func DeleteByMealId(ctx context.Context, db *gorm.DB, mealId uuid.UUID) error {
	if err := db.WithContext(ctx).Delete(&Entity{}, "meal_id = ?", mealId).Error; err != nil {
		return fmt.Errorf("failed to delete ingredients: %w", err)
	}
	return nil
}
