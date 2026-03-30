package planitem

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func createItem(db *gorm.DB, e *Entity) error {
	e.Id = uuid.New()
	now := time.Now().UTC()
	e.CreatedAt = now
	e.UpdatedAt = now
	return db.Create(e).Error
}

func updateItem(db *gorm.DB, e *Entity) error {
	e.UpdatedAt = time.Now().UTC()
	return db.Save(e).Error
}

func deleteItem(db *gorm.DB, id uuid.UUID) error {
	return db.Where("id = ?", id).Delete(&Entity{}).Error
}

func deleteByPlanWeekID(db *gorm.DB, planWeekID uuid.UUID) error {
	return db.Where("plan_week_id = ?", planWeekID).Delete(&Entity{}).Error
}
