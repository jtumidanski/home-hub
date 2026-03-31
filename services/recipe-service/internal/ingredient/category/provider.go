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

// GetAll returns all categories for the tenant in the request context.
// Tenant filtering is automatic via GORM callbacks on db.WithContext(ctx).
func GetAll() database.EntityProvider[[]Entity] {
	return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC")
	})
}

// GetByName returns a category matching the given name for the tenant in context.
// Tenant filtering is automatic via GORM callbacks on db.WithContext(ctx).
func GetByName(name string) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("name = ?", name)
	})
}

// CountIngredientsByCategory queries the canonical_ingredients table (not the category table),
// so it uses an explicit category_id filter. Tenant filtering applies via GORM callbacks.
func CountIngredientsByCategory(categoryID uuid.UUID) func(db *gorm.DB) (int64, error) {
	return func(db *gorm.DB) (int64, error) {
		var count int64
		err := db.Table("canonical_ingredients").
			Where("category_id = ?", categoryID).
			Count(&count).Error
		return count, err
	}
}

// getMaxSortOrder returns the highest sort_order for categories in the tenant from context.
func getMaxSortOrder(db *gorm.DB) (int, error) {
	var maxOrder *int
	err := db.Model(&Entity{}).
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

// countAll returns the total number of categories for the tenant from context.
func countAll(db *gorm.DB) (int64, error) {
	var count int64
	err := db.Model(&Entity{}).Count(&count).Error
	return count, err
}
