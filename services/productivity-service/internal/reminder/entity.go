package reminder

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entity struct {
	Id               uuid.UUID  `gorm:"type:uuid;primaryKey"`
	TenantId         uuid.UUID  `gorm:"type:uuid;not null;index"`
	HouseholdId      uuid.UUID  `gorm:"type:uuid;not null;index"`
	Title            string     `gorm:"type:text;not null"`
	Notes            string     `gorm:"type:text"`
	ScheduledFor     time.Time  `gorm:"not null;index"`
	LastDismissedAt  *time.Time
	LastSnoozedUntil *time.Time
	CreatedAt        time.Time  `gorm:"not null"`
	UpdatedAt        time.Time  `gorm:"not null"`
}

func (Entity) TableName() string { return "reminders" }

func Migration(db *gorm.DB) error { return db.AutoMigrate(&Entity{}) }

func Make(e Entity) (Model, error) {
	return Model{
		id: e.Id, tenantID: e.TenantId, householdID: e.HouseholdId,
		title: e.Title, notes: e.Notes, scheduledFor: e.ScheduledFor,
		lastDismissedAt: e.LastDismissedAt, lastSnoozedUntil: e.LastSnoozedUntil,
		createdAt: e.CreatedAt, updatedAt: e.UpdatedAt,
	}, nil
}
