package connection

import (
	"github.com/google/uuid"
	database "github.com/jtumidanski/home-hub/shared/go/database"
	"gorm.io/gorm"
)

func getByID(id uuid.UUID) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	})
}

func getByUserAndHousehold(userID, householdID uuid.UUID) database.EntityProvider[[]Entity] {
	return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ? AND household_id = ?", userID, householdID)
	})
}

func getByHousehold(householdID uuid.UUID) database.EntityProvider[[]Entity] {
	return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("household_id = ?", householdID)
	})
}

func getAllConnected() database.EntityProvider[[]Entity] {
	return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", "connected")
	})
}

func getByUserAndProvider(userID uuid.UUID, provider string) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ? AND provider = ?", userID, provider)
	})
}

func countByHousehold(db *gorm.DB, householdID uuid.UUID) (int64, error) {
	var count int64
	err := db.Model(&Entity{}).Where("household_id = ?", householdID).Count(&count).Error
	return count, err
}
