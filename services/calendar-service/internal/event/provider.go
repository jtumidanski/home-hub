package event

import (
	"time"

	"github.com/google/uuid"
	database "github.com/jtumidanski/home-hub/shared/go/database"
	"gorm.io/gorm"
)

func getByID(id uuid.UUID) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	})
}

func getByHouseholdAndTimeRange(householdID uuid.UUID, start, end time.Time) database.EntityProvider[[]Entity] {
	return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("household_id = ? AND start_time < ? AND end_time > ?", householdID, end, start).
			Order("all_day DESC, start_time ASC")
	})
}

func getVisibleByHouseholdAndTimeRange(householdID uuid.UUID, start, end time.Time) database.EntityProvider[[]Entity] {
	return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("household_id = ? AND start_time < ? AND end_time > ?", householdID, end, start).
			Where("source_id IN (SELECT id FROM calendar_sources WHERE visible = true)").
			Order("all_day DESC, start_time ASC")
	})
}
