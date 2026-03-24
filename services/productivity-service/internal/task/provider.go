package task

import (
	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/model"
	"gorm.io/gorm"
)

func getByID(id uuid.UUID) func(db *gorm.DB) model.Provider[Entity] {
	return func(db *gorm.DB) model.Provider[Entity] {
		var result Entity
		err := db.Where("id = ?", id).First(&result).Error
		if err != nil {
			return model.ErrorProvider[Entity](err)
		}
		return model.FixedProvider(result)
	}
}

func getAll(includeDeleted bool) func(db *gorm.DB) model.Provider[[]Entity] {
	return func(db *gorm.DB) model.Provider[[]Entity] {
		var results []Entity
		q := db
		if !includeDeleted {
			q = q.Where("deleted_at IS NULL")
		}
		err := q.Find(&results).Error
		if err != nil {
			return model.ErrorProvider[[]Entity](err)
		}
		return model.FixedProvider(results)
	}
}

func getByStatus(status string) func(db *gorm.DB) model.Provider[[]Entity] {
	return func(db *gorm.DB) model.Provider[[]Entity] {
		var results []Entity
		err := db.Where("status = ? AND deleted_at IS NULL", status).Find(&results).Error
		if err != nil {
			return model.ErrorProvider[[]Entity](err)
		}
		return model.FixedProvider(results)
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
