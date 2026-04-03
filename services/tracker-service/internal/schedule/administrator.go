package schedule

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func CreateSnapshot(db *gorm.DB, trackingItemID uuid.UUID, days []int, effectiveDate time.Time) (Entity, error) {
	sched, _ := json.Marshal(days)
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
