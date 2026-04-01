package item

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

func uncheckAll(db *gorm.DB, listID uuid.UUID) error {
	return db.Model(&Entity{}).
		Where("list_id = ? AND checked = ?", listID, true).
		Updates(map[string]interface{}{
			"checked":    false,
			"updated_at": time.Now().UTC(),
		}).Error
}
