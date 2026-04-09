package planneditem

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func createPlannedItem(db *gorm.DB, e *Entity) error {
	if e.Id == uuid.Nil {
		e.Id = uuid.New()
	}
	now := time.Now().UTC()
	if e.CreatedAt.IsZero() {
		e.CreatedAt = now
	}
	e.UpdatedAt = now
	return db.Create(e).Error
}

func updatePlannedItem(db *gorm.DB, e *Entity) error {
	e.UpdatedAt = time.Now().UTC()
	return db.Save(e).Error
}

func deletePlannedItem(db *gorm.DB, id uuid.UUID) error {
	return db.Where("id = ?", id).Delete(&Entity{}).Error
}
