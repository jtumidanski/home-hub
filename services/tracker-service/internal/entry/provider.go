package entry

import (
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"gorm.io/gorm"
)

func GetByItemAndDate(itemID uuid.UUID, date time.Time) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("tracking_item_id = ? AND date = ?", itemID, date)
	})
}

func GetByUserAndMonth(userID uuid.UUID, start, end time.Time) database.EntityProvider[[]Entity] {
	return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ? AND date >= ? AND date <= ?", userID, start, end)
	})
}

func GetByItemAndDateRange(itemID uuid.UUID, start, end time.Time) database.EntityProvider[[]Entity] {
	return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("tracking_item_id = ? AND date >= ? AND date <= ?", itemID, start, end)
	})
}
