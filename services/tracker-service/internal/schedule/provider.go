package schedule

import (
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"gorm.io/gorm"
)

func GetByTrackingItemID(itemID uuid.UUID) database.EntityProvider[[]Entity] {
	return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("tracking_item_id = ?", itemID).Order("effective_date ASC")
	})
}

func GetEffectiveSchedule(itemID uuid.UUID, date time.Time) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("tracking_item_id = ? AND effective_date <= ?", itemID, date).
			Order("effective_date DESC").
			Limit(1)
	})
}

func GetByTrackingItemIDs(itemIDs []uuid.UUID) database.EntityProvider[[]Entity] {
	return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("tracking_item_id IN ?", itemIDs).Order("tracking_item_id, effective_date ASC")
	})
}
