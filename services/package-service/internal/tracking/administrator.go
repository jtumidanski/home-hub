package tracking

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func create(db *gorm.DB, e *Entity) error {
	return db.Create(e).Error
}

func update(db *gorm.DB, e *Entity) error {
	return db.Save(e).Error
}

func deleteByID(db *gorm.DB, id uuid.UUID) error {
	return db.Where("id = ?", id).Delete(&Entity{}).Error
}
