package entry

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func createEntry(db *gorm.DB, e *Entity) error {
	e.Id = uuid.New()
	now := time.Now().UTC()
	e.CreatedAt = now
	e.UpdatedAt = now
	return db.Create(e).Error
}

func updateEntry(db *gorm.DB, e *Entity) error {
	e.UpdatedAt = time.Now().UTC()
	return db.Save(e).Error
}

func deleteEntry(db *gorm.DB, id uuid.UUID) error {
	return db.Where("id = ?", id).Delete(&Entity{}).Error
}
