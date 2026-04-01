package ingredient

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func createEntity(db *gorm.DB, e *Entity) error {
	return db.Create(e).Error
}

func saveEntity(db *gorm.DB, e *Entity) error {
	e.UpdatedAt = time.Now().UTC()
	return db.Save(e).Error
}

func deleteEntity(db *gorm.DB, id uuid.UUID) error {
	return db.Where("id = ?", id).Delete(&Entity{}).Error
}

func createAlias(db *gorm.DB, alias *AliasEntity) error {
	return db.Create(alias).Error
}

func deleteAlias(db *gorm.DB, id uuid.UUID) error {
	return db.Where("id = ?", id).Delete(&AliasEntity{}).Error
}

func nullifyReferences(db *gorm.DB, canonicalIngredientID uuid.UUID) error {
	return db.Table("recipe_ingredients").
		Where("canonical_ingredient_id = ?", canonicalIngredientID).
		Updates(map[string]interface{}{
			"canonical_ingredient_id": nil,
			"normalization_status":    "unresolved",
			"updated_at":             time.Now().UTC(),
		}).Error
}

func bulkUpdateCategory(db *gorm.DB, ingredientIDs []uuid.UUID, tenantID, categoryID uuid.UUID) error {
	return db.Transaction(func(tx *gorm.DB) error {
		return tx.Table("canonical_ingredients").
			Where("id IN ? AND tenant_id = ?", ingredientIDs, tenantID).
			Updates(map[string]interface{}{
				"category_id": categoryID,
				"updated_at":  time.Now().UTC(),
			}).Error
	})
}
