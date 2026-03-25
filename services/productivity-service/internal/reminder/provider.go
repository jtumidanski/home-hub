package reminder

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

func getAll() database.EntityProvider[[]Entity] {
	return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB {
		return db
	})
}

func countDueNow(db *gorm.DB) (int64, error) {
	var count int64
	err := db.Model(&Entity{}).
		Where("scheduled_for <= CURRENT_TIMESTAMP AND last_dismissed_at IS NULL AND (last_snoozed_until IS NULL OR last_snoozed_until <= CURRENT_TIMESTAMP)").
		Count(&count).Error
	return count, err
}

func countUpcoming(db *gorm.DB) (int64, error) {
	var count int64
	err := db.Model(&Entity{}).
		Where("scheduled_for > CURRENT_TIMESTAMP AND last_dismissed_at IS NULL").
		Count(&count).Error
	return count, err
}

func countSnoozed(db *gorm.DB) (int64, error) {
	var count int64
	err := db.Model(&Entity{}).
		Where("last_snoozed_until > CURRENT_TIMESTAMP AND last_dismissed_at IS NULL").
		Count(&count).Error
	return count, err
}
