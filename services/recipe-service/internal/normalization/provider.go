package normalization

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func getByRecipeID(db *gorm.DB, recipeID uuid.UUID) ([]Entity, error) {
	var entities []Entity
	err := db.Where("recipe_id = ?", recipeID).Order("position ASC").Find(&entities).Error
	return entities, err
}

func GetByCanonicalIngredientID(db *gorm.DB, canonicalIngredientID uuid.UUID, page, pageSize int) ([]Entity, int64, error) {
	query := db.Model(&Entity{}).Where("canonical_ingredient_id = ?", canonicalIngredientID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	var entities []Entity
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&entities).Error
	return entities, total, err
}

func getByID(db *gorm.DB, id uuid.UUID) (Entity, error) {
	var e Entity
	err := db.Where("id = ?", id).First(&e).Error
	return e, err
}

func ReassignCanonical(db *gorm.DB, fromID, toID uuid.UUID) (int64, error) {
	result := db.Model(&Entity{}).
		Where("canonical_ingredient_id = ?", fromID).
		Updates(map[string]interface{}{
			"canonical_ingredient_id": toID,
			"updated_at":             time.Now().UTC(),
		})
	return result.RowsAffected, result.Error
}
