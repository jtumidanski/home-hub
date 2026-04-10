package week

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func createWeek(db *gorm.DB, e *Entity) error {
	if e.Id == uuid.Nil {
		e.Id = uuid.New()
	}
	now := time.Now().UTC()
	if e.CreatedAt.IsZero() {
		e.CreatedAt = now
	}
	e.UpdatedAt = now
	if len(e.RestDayFlags) == 0 {
		e.RestDayFlags = json.RawMessage("[]")
	}
	return db.Create(e).Error
}

func updateWeek(db *gorm.DB, e *Entity) error {
	e.UpdatedAt = time.Now().UTC()
	return db.Save(e).Error
}
