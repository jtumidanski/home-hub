package schedule

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entity struct {
	Id             uuid.UUID       `gorm:"type:uuid;primaryKey"`
	TrackingItemId uuid.UUID       `gorm:"type:uuid;not null;uniqueIndex:idx_schedule_item_date;index:idx_schedule_item"`
	Schedule       json.RawMessage `gorm:"type:jsonb;not null"`
	EffectiveDate  time.Time       `gorm:"type:date;not null;uniqueIndex:idx_schedule_item_date"`
	CreatedAt      time.Time       `gorm:"not null"`
}

func (Entity) TableName() string { return "schedule_snapshots" }

func Migration(db *gorm.DB) error {
	return db.AutoMigrate(&Entity{})
}

func (m Model) ToEntity() Entity {
	sched, _ := json.Marshal(m.schedule)
	return Entity{
		Id:             m.id,
		TrackingItemId: m.trackingItemID,
		Schedule:       sched,
		EffectiveDate:  m.effectiveDate,
		CreatedAt:      m.createdAt,
	}
}

func Make(e Entity) (Model, error) {
	var sched []int
	if err := json.Unmarshal(e.Schedule, &sched); err != nil {
		sched = []int{}
	}
	return NewBuilder().
		SetId(e.Id).
		SetTrackingItemID(e.TrackingItemId).
		SetSchedule(sched).
		SetEffectiveDate(e.EffectiveDate).
		SetCreatedAt(e.CreatedAt).
		Build()
}
