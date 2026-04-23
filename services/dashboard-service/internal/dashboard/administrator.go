package dashboard

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func insert(db *gorm.DB, e Entity) (Entity, error) {
	if e.Id == uuid.Nil {
		e.Id = uuid.New()
	}
	now := time.Now().UTC()
	e.CreatedAt = now
	e.UpdatedAt = now
	if err := db.Create(&e).Error; err != nil {
		return Entity{}, err
	}
	return e, nil
}

func updateFields(db *gorm.DB, id uuid.UUID, fields map[string]any) (Entity, error) {
	fields["updated_at"] = time.Now().UTC()
	if err := db.Model(&Entity{}).Where("id = ?", id).Updates(fields).Error; err != nil {
		return Entity{}, err
	}
	var out Entity
	if err := db.First(&out, "id = ?", id).Error; err != nil {
		return Entity{}, err
	}
	return out, nil
}

func deleteByID(db *gorm.DB, id uuid.UUID) error {
	return db.Delete(&Entity{}, "id = ?", id).Error
}

func updateSortOrders(db *gorm.DB, updates map[uuid.UUID]int) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for id, order := range updates {
			if err := tx.Model(&Entity{}).Where("id = ?", id).
				Updates(map[string]any{"sort_order": order, "updated_at": time.Now().UTC()}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// layoutAsJSON casts a validated Layout into datatypes.JSON safely.
func layoutAsJSON(raw []byte) datatypes.JSON { return datatypes.JSON(raw) }
