package dismissal

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
	CreatedByUserId uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt       time.Time `gorm:"not null"`
}

func (Entity) TableName() string { return "reminder_dismissals" }

func Migration(db *gorm.DB) error { return db.AutoMigrate(&Entity{}) }

func (m Model) ToEntity() Entity {
	return Entity{
		Id:              m.id,
		TenantId:        m.tenantID,
		HouseholdId:     m.householdID,
		ReminderId:      m.reminderID,
		CreatedByUserId: m.createdByUserID,
		CreatedAt:       m.createdAt,
	}
}

func Make(e Entity) (Model, error) {
	return NewBuilder().
		SetId(e.Id).
		SetTenantID(e.TenantId).
		SetHouseholdID(e.HouseholdId).
		SetReminderID(e.ReminderId).
		SetCreatedByUserID(e.CreatedByUserId).
		SetCreatedAt(e.CreatedAt).
		Build()
}
