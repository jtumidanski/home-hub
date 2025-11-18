package meal

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GetById retrieves a meal by ID
func GetById(ctx context.Context, db *gorm.DB, id uuid.UUID) (Model, error) {
	var entity Entity
	if err := db.WithContext(ctx).Where("id = ?", id).First(&entity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return Model{}, fmt.Errorf("meal not found: %s", id)
		}
		return Model{}, fmt.Errorf("failed to get meal: %w", err)
	}
	return Make(entity)
}

// ListByHousehold retrieves all meals for a household, ordered by creation date descending
func ListByHousehold(ctx context.Context, db *gorm.DB, householdId uuid.UUID) ([]Model, error) {
	var entities []Entity
	if err := db.WithContext(ctx).
		Where("household_id = ?", householdId).
		Order("created_at DESC").
		Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("failed to list meals: %w", err)
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

// Save persists a meal to the database
func Save(ctx context.Context, db *gorm.DB, m Model) error {
	entity := m.ToEntity()
	if err := db.WithContext(ctx).Save(&entity).Error; err != nil {
		return fmt.Errorf("failed to save meal: %w", err)
	}
	return nil
}

// DeleteById deletes a meal by ID
func DeleteById(ctx context.Context, db *gorm.DB, id uuid.UUID) error {
	if err := db.WithContext(ctx).Delete(&Entity{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete meal: %w", err)
	}
	return nil
}
