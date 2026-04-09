package forecast

import (
	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"gorm.io/gorm"
)

func getByHouseholdAndLocation(householdID uuid.UUID, locationID *uuid.UUID) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		q := db.Where("household_id = ?", householdID)
		if locationID == nil {
			return q.Where("location_id IS NULL")
		}
		return q.Where("location_id = ?", *locationID)
	})
}

func getAll() database.EntityProvider[[]Entity] {
	return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB {
		return db
	})
}
