package item

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

func GetByListID(listID uuid.UUID) database.EntityProvider[[]Entity] {
	return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("list_id = ?", listID).
			Order("COALESCE(category_sort_order, 2147483647) ASC, position ASC, created_at ASC")
	})
}

func getMaxPosition(db *gorm.DB, listID uuid.UUID) (int, error) {
	var maxPos *int
	err := db.Model(&Entity{}).
		Where("list_id = ?", listID).
		Select("MAX(position)").
		Scan(&maxPos).Error
	if err != nil {
		return 0, err
	}
	if maxPos == nil {
		return 0, nil
	}
	return *maxPos, nil
}
