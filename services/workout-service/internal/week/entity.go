package week

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Entity is the GORM mapping for `workout.weeks`.
//
// `rest_day_flags` is stored as `jsonb` (rather than a postgres int[]) for
// portability and so that the existing tracker-service jsonb pattern is
// reused. Day-of-week ints are 0..6 with 0 = Monday (ISO).
type Entity struct {
	Id            uuid.UUID       `gorm:"type:uuid;primaryKey"`
	TenantId      uuid.UUID       `gorm:"type:uuid;not null;uniqueIndex:idx_workout_weeks_tenant_user_start,priority:1"`
	UserId        uuid.UUID       `gorm:"type:uuid;not null;uniqueIndex:idx_workout_weeks_tenant_user_start,priority:2"`
	WeekStartDate time.Time       `gorm:"type:date;not null;uniqueIndex:idx_workout_weeks_tenant_user_start,priority:3"`
	RestDayFlags  json.RawMessage `gorm:"type:jsonb;not null;default:'[]'"`
	CreatedAt     time.Time       `gorm:"not null"`
	UpdatedAt     time.Time       `gorm:"not null"`
}

func (Entity) TableName() string { return "weeks" }

func Migration(db *gorm.DB) error {
	return db.AutoMigrate(&Entity{})
}

func (m Model) ToEntity() (Entity, error) {
	flags, err := json.Marshal(m.restDayFlags)
	if err != nil {
		return Entity{}, err
	}
	return Entity{
		Id:            m.id,
		TenantId:      m.tenantID,
		UserId:        m.userID,
		WeekStartDate: m.weekStartDate,
		RestDayFlags:  flags,
		CreatedAt:     m.createdAt,
		UpdatedAt:     m.updatedAt,
	}, nil
}

func Make(e Entity) (Model, error) {
	var flags []int
	if len(e.RestDayFlags) > 0 && string(e.RestDayFlags) != "null" {
		if err := json.Unmarshal(e.RestDayFlags, &flags); err != nil {
			return Model{}, err
		}
	}
	return NewBuilder().
		SetId(e.Id).
		SetTenantID(e.TenantId).
		SetUserID(e.UserId).
		SetWeekStartDate(e.WeekStartDate).
		SetRestDayFlags(flags).
		SetCreatedAt(e.CreatedAt).
		SetUpdatedAt(e.UpdatedAt).
		Build()
}
