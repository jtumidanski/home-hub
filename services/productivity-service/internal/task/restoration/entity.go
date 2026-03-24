package restoration

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entity struct {
	Id              uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantId        uuid.UUID `gorm:"type:uuid;not null;index"`
	HouseholdId     uuid.UUID `gorm:"type:uuid;not null;index"`
	TaskId          uuid.UUID `gorm:"type:uuid;not null"`
	CreatedByUserId uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt       time.Time `gorm:"not null"`
}

func (Entity) TableName() string { return "task_restorations" }

func Migration(db *gorm.DB) error { return db.AutoMigrate(&Entity{}) }

func Make(e Entity) (Model, error) {
	return NewBuilder().
		SetId(e.Id).
		SetTenantID(e.TenantId).
		SetHouseholdID(e.HouseholdId).
		SetTaskID(e.TaskId).
		SetCreatedByUserID(e.CreatedByUserId).
		SetCreatedAt(e.CreatedAt).
		Build()
}
