package task

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

func getAll(includeDeleted bool) func(db *gorm.DB) model.Provider[[]Entity] {
	return func(db *gorm.DB) model.Provider[[]Entity] {
		return func() ([]Entity, error) {
			var results []Entity
			q := db
			if !includeDeleted {
				q = q.Where("deleted_at IS NULL")
			}
			err := q.Find(&results).Error
			return results, err
		}
	}
}

func getByStatus(status string) func(db *gorm.DB) model.Provider[[]Entity] {
	return func(db *gorm.DB) model.Provider[[]Entity] {
		return func() ([]Entity, error) {
			var results []Entity
			err := db.Where("status = ? AND deleted_at IS NULL", status).Find(&results).Error
			return results, err
		}
	}
}

func countByStatus(db *gorm.DB, status string) (int64, error) {
	var count int64
	err := db.Model(&Entity{}).Where("status = ? AND deleted_at IS NULL", status).Count(&count).Error
	return count, err
}

func countOverdue(db *gorm.DB) (int64, error) {
	var count int64
	err := db.Model(&Entity{}).
		Where("status = ? AND deleted_at IS NULL AND due_on < CURRENT_DATE", "pending").
		Count(&count).Error
	return count, err
}

func countCompletedToday(db *gorm.DB) (int64, error) {
	var count int64
	err := db.Model(&Entity{}).
		Where("status = ? AND deleted_at IS NULL AND completed_at >= CURRENT_DATE", "completed").
		Count(&count).Error
	return count, err
}
