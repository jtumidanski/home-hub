package task

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

func getAll(includeDeleted bool) database.EntityProvider[[]Entity] {
	return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB {
		if !includeDeleted {
			db = db.Where("deleted_at IS NULL")
		}
		return db.Order("status ASC").Order("due_on ASC NULLS LAST")
	})
}

func getByStatus(status string) database.EntityProvider[[]Entity] {
	return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ? AND deleted_at IS NULL", status)
	})
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
