package exercise

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func createExercise(db *gorm.DB, e *Entity) error {
	if e.Id == uuid.Nil {
		e.Id = uuid.New()
	}
	now := time.Now().UTC()
	if e.CreatedAt.IsZero() {
		e.CreatedAt = now
	}
	e.UpdatedAt = now
	if len(e.SecondaryRegionIds) == 0 {
		e.SecondaryRegionIds = []byte("[]")
	}
	return db.Create(e).Error
}

func updateExercise(db *gorm.DB, e *Entity) error {
	e.UpdatedAt = time.Now().UTC()
	return db.Save(e).Error
}

func softDeleteExercise(db *gorm.DB, e *Entity) error {
	now := time.Now().UTC()
	e.DeletedAt = &now
	e.UpdatedAt = now
	return db.Save(e).Error
}
