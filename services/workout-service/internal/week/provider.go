package week

import (
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"gorm.io/gorm"
)

func GetByID(id uuid.UUID) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	})
}

func GetByUserAndStart(userID uuid.UUID, weekStart time.Time) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ? AND week_start_date = ?", userID, weekStart)
	})
}

// GetMostRecentPriorWithItems is the source-of-truth lookup for the
// "copy from previous week" endpoint. It returns the most recent week (strictly
// earlier than `before`) that owns at least one planned_items row.
func GetMostRecentPriorWithItems(db *gorm.DB, userID uuid.UUID, before time.Time) (Entity, error) {
	var e Entity
	err := db.
		Joins("INNER JOIN planned_items ON planned_items.week_id = weeks.id").
		Where("weeks.user_id = ? AND weeks.week_start_date < ?", userID, before).
		Order("weeks.week_start_date DESC").
		Group("weeks.id").
		First(&e).Error
	return e, err
}
