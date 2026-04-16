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

func countOverdue(db *gorm.DB, now time.Time) (int64, error) {
	todayDate := now.Format("2006-01-02")
	var count int64
	err := db.Model(&Entity{}).
		Where("status = ? AND deleted_at IS NULL AND due_on < ?", "pending", todayDate).
		Count(&count).Error
	return count, err
}

func countCompletedToday(db *gorm.DB, now time.Time) (int64, error) {
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).UTC()
	var count int64
	err := db.Model(&Entity{}).
		Where("status = ? AND deleted_at IS NULL AND completed_at >= ?", "completed", startOfDay).
		Count(&count).Error
	return count, err
}
