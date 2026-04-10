package region

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func createRegion(db *gorm.DB, e *Entity) error {
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

func updateRegion(db *gorm.DB, e *Entity) error {
	e.UpdatedAt = time.Now().UTC()
	return db.Save(e).Error
}

func softDeleteRegion(db *gorm.DB, e *Entity) error {
	now := time.Now().UTC()
	e.DeletedAt = &now
	e.UpdatedAt = now
	return db.Save(e).Error
}
