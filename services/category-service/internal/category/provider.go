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

func GetAll() database.EntityProvider[[]Entity] {
	return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC")
	})
}

func GetByName(name string) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("name = ?", name)
	})
}

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

func countAll(db *gorm.DB) (int64, error) {
	var count int64
	err := db.Model(&Entity{}).Count(&count).Error
	return count, err
}
