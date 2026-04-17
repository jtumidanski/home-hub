package task

import (
	"time"

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

func countOverdue(db *gorm.DB, date time.Time) (int64, error) {
	dateStr := date.Format("2006-01-02")
	var count int64
	err := db.Model(&Entity{}).
		Where("status = ? AND deleted_at IS NULL AND due_on < ?", "pending", dateStr).
		Count(&count).Error
	return count, err
}

func countCompletedToday(db *gorm.DB, date time.Time) (int64, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	var count int64
	err := db.Model(&Entity{}).
		Where("status = ? AND deleted_at IS NULL AND completed_at >= ?", "completed", startOfDay).
		Count(&count).Error
	return count, err
}
