package event

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	database "github.com/jtumidanski/home-hub/shared/go/database"
)

func getByID(id uuid.UUID) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	})
}

func getVisibleByHouseholdAndTimeRange(householdID uuid.UUID, start, end time.Time) database.EntityProvider[[]Entity] {
	return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("household_id = ? AND start_time < ? AND end_time > ?", householdID, end, start).
			Where("source_id IN (SELECT id FROM calendar_sources WHERE visible = true)").
			Order("all_day DESC, start_time ASC")
	})
}
