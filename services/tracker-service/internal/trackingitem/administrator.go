package trackingitem

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func createTrackingItem(db *gorm.DB, e *Entity) error {
	e.Id = uuid.New()
	now := time.Now().UTC()
	e.CreatedAt = now
	e.UpdatedAt = now
	return db.Create(e).Error
}

func updateTrackingItem(db *gorm.DB, e *Entity) error {
	e.UpdatedAt = time.Now().UTC()
	return db.Save(e).Error
}

func softDeleteTrackingItem(db *gorm.DB, e *Entity) error {
	now := time.Now().UTC()
	e.DeletedAt = &now
	e.UpdatedAt = now
	return db.Save(e).Error
}
