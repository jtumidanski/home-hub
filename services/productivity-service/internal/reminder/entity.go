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

func (m Model) ToEntity() Entity {
	return Entity{
		Id: m.id, TenantId: m.tenantID, HouseholdId: m.householdID,
		Title: m.title, Notes: m.notes, ScheduledFor: m.scheduledFor,
		LastDismissedAt: m.lastDismissedAt, LastSnoozedUntil: m.lastSnoozedUntil,
		CreatedAt: m.createdAt, UpdatedAt: m.updatedAt,
	}
}

func Make(e Entity) (Model, error) {
	return NewBuilder().
		SetId(e.Id).
		SetTenantID(e.TenantId).
		SetHouseholdID(e.HouseholdId).
		SetTitle(e.Title).
		SetNotes(e.Notes).
		SetScheduledFor(e.ScheduledFor).
		SetLastDismissedAt(e.LastDismissedAt).
		SetLastSnoozedUntil(e.LastSnoozedUntil).
		SetCreatedAt(e.CreatedAt).
		SetUpdatedAt(e.UpdatedAt).
		Build()
}
