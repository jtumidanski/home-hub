package wishlist

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
	// Save with explicit fields so we never touch vote_count from this path,
	// even if a stray value is set on the in-memory entity.
	return db.Model(e).
		Where("id = ? AND household_id = ?", e.Id, e.HouseholdId).
		Updates(map[string]interface{}{
			"name":              e.Name,
			"purchase_location": e.PurchaseLocation,
			"urgency":           e.Urgency,
			"updated_at":        e.UpdatedAt,
		}).Error
}

func deleteItem(db *gorm.DB, id uuid.UUID, householdID uuid.UUID) error {
	return db.Where("id = ? AND household_id = ?", id, householdID).Delete(&Entity{}).Error
}

// incrementVote performs an atomic UPDATE … SET vote_count = vote_count + 1.
// Returns gorm.ErrRecordNotFound if no row matches.
func incrementVote(db *gorm.DB, id uuid.UUID, householdID uuid.UUID) error {
	res := db.Model(&Entity{}).
		Where("id = ? AND household_id = ?", id, householdID).
		Updates(map[string]interface{}{
			"vote_count": gorm.Expr("vote_count + 1"),
			"updated_at": time.Now().UTC(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
