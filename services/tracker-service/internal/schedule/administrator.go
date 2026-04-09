package schedule

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreateSnapshot creates a schedule snapshot for the given item and effective
// date, or updates the existing snapshot's schedule when one already exists for
// that exact (item, date) pair. This makes same-day schedule edits idempotent
// rather than colliding on the unique constraint.
func CreateSnapshot(db *gorm.DB, trackingItemID uuid.UUID, days []int, effectiveDate time.Time) (Entity, error) {
	sched, _ := json.Marshal(days)

	var existing Entity
	err := db.Where("tracking_item_id = ? AND effective_date = ?", trackingItemID, effectiveDate).
		Take(&existing).Error
	if err == nil {
		existing.Schedule = sched
		if err := db.Save(&existing).Error; err != nil {
			return Entity{}, err
		}
		return existing, nil
	}

	e := Entity{
		Id:             uuid.New(),
		TrackingItemId: trackingItemID,
		Schedule:       sched,
		EffectiveDate:  effectiveDate,
		CreatedAt:      time.Now().UTC(),
	}
	if err := db.Create(&e).Error; err != nil {
		return Entity{}, err
	}
	return e, nil
}
