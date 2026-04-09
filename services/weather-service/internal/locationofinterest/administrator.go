package locationofinterest

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func createLocation(db *gorm.DB, e *Entity) error {
	e.Id = uuid.New()
	now := time.Now().UTC()
	e.CreatedAt = now
	e.UpdatedAt = now
	return db.Create(e).Error
}

func updateLocation(db *gorm.DB, e *Entity) error {
	e.UpdatedAt = time.Now().UTC()
	return db.Save(e).Error
}

func deleteLocation(db *gorm.DB, id, householdID uuid.UUID) error {
	return db.Where("id = ? AND household_id = ?", id, householdID).Delete(&Entity{}).Error
}
