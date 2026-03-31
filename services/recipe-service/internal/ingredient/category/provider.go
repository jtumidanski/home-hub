package category

import (
	"github.com/google/uuid"
	database "github.com/jtumidanski/home-hub/shared/go/database"
	"gorm.io/gorm"
)

func GetByID(id uuid.UUID) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	})
}

func GetByTenantID(tenantID uuid.UUID) func(db *gorm.DB) ([]Entity, error) {
	return func(db *gorm.DB) ([]Entity, error) {
		var entities []Entity
		err := db.Where("tenant_id = ?", tenantID).Order("sort_order ASC").Find(&entities).Error
		return entities, err
	}
}

func GetByName(tenantID uuid.UUID, name string) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("tenant_id = ? AND name = ?", tenantID, name)
	})
}

func CountIngredientsByCategory(db *gorm.DB, categoryID uuid.UUID) (int64, error) {
	var count int64
	err := db.Table("canonical_ingredients").
		Where("category_id = ?", categoryID).
		Count(&count).Error
	return count, err
}

func GetMaxSortOrder(db *gorm.DB, tenantID uuid.UUID) (int, error) {
	var maxOrder *int
	err := db.Model(&Entity{}).
		Where("tenant_id = ?", tenantID).
		Select("MAX(sort_order)").
		Scan(&maxOrder).Error
	if err != nil {
		return 0, err
	}
	if maxOrder == nil {
		return 0, nil
	}
	return *maxOrder, nil
}
