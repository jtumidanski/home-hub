package reminder

import (
	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/model"
	"gorm.io/gorm"
)

func getByID(id uuid.UUID) func(db *gorm.DB) model.Provider[Entity] {
	return func(db *gorm.DB) model.Provider[Entity] {
		return func() (Entity, error) {
			var result Entity
			err := db.Where("id = ?", id).First(&result).Error
			return result, err
		}
	}
}

func getAll() func(db *gorm.DB) model.Provider[[]Entity] {
	return func(db *gorm.DB) model.Provider[[]Entity] {
		return func() ([]Entity, error) {
			var results []Entity
			err := db.Find(&results).Error
			return results, err
		}
	}
}

func countDueNow(db *gorm.DB) (int64, error) {
	var count int64
	err := db.Model(&Entity{}).
		Where("scheduled_for <= NOW() AND last_dismissed_at IS NULL AND (last_snoozed_until IS NULL OR last_snoozed_until <= NOW())").
		Count(&count).Error
	return count, err
}

func countUpcoming(db *gorm.DB) (int64, error) {
	var count int64
	err := db.Model(&Entity{}).
		Where("scheduled_for > NOW() AND last_dismissed_at IS NULL").
		Count(&count).Error
	return count, err
}

func countSnoozed(db *gorm.DB) (int64, error) {
	var count int64
	err := db.Model(&Entity{}).
		Where("last_snoozed_until > NOW() AND last_dismissed_at IS NULL").
		Count(&count).Error
	return count, err
}
