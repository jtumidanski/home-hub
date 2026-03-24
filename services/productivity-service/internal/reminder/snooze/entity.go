package snooze

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entity struct {
	Id              uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantId        uuid.UUID `gorm:"type:uuid;not null;index"`
	HouseholdId     uuid.UUID `gorm:"type:uuid;not null;index"`
	ReminderId      uuid.UUID `gorm:"type:uuid;not null"`
	DurationMinutes int       `gorm:"not null"`
	SnoozedUntil    time.Time `gorm:"not null"`
	CreatedByUserId uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt       time.Time `gorm:"not null"`
}

func (Entity) TableName() string { return "reminder_snoozes" }

func Migration(db *gorm.DB) error { return db.AutoMigrate(&Entity{}) }
