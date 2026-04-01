package planner

import (
	"time"

	"gorm.io/gorm"
)

func createConfig(db *gorm.DB, e *Entity) error {
	return db.Create(e).Error
}

func updateConfig(db *gorm.DB, e *Entity) error {
	e.UpdatedAt = time.Now().UTC()
	return db.Save(e).Error
}
