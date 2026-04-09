package planneditem

import (
	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"gorm.io/gorm"
)

func GetByID(id uuid.UUID) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	})
}

func GetByWeek(weekID uuid.UUID) database.EntityProvider[[]Entity] {
	return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("week_id = ?", weekID).Order("day_of_week ASC, position ASC")
	})
}

// MaxPositionForDay returns the highest existing position on the given day,
// or -1 when the day is empty. Used by the auto-assign branch in the create
// path so omitted positions append to the day's tail rather than colliding.
func MaxPositionForDay(db *gorm.DB, weekID uuid.UUID, dayOfWeek int) (int, error) {
	var max *int
	err := db.Model(&Entity{}).
		Where("week_id = ? AND day_of_week = ?", weekID, dayOfWeek).
		Select("MAX(position)").
		Scan(&max).Error
	if err != nil {
		return 0, err
	}
	if max == nil {
		return -1, nil
	}
	return *max, nil
}
