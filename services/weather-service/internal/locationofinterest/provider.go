package locationofinterest

import (
	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"gorm.io/gorm"
)

func GetByID(id, householdID uuid.UUID) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ? AND household_id = ?", id, householdID)
	})
}

func ListByHousehold(householdID uuid.UUID) database.EntityProvider[[]Entity] {
	return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("household_id = ?", householdID).Order("created_at ASC")
	})
}

func countByHousehold(db *gorm.DB, householdID uuid.UUID) (int64, error) {
	var count int64
	err := db.Model(&Entity{}).Where("household_id = ?", householdID).Count(&count).Error
	return count, err
}
