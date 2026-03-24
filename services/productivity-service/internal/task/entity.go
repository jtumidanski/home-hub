package task

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entity struct {
	Id              uuid.UUID  `gorm:"type:uuid;primaryKey"`
	TenantId        uuid.UUID  `gorm:"type:uuid;not null;index"`
	HouseholdId     uuid.UUID  `gorm:"type:uuid;not null;index"`
	Title           string     `gorm:"type:text;not null"`
	Notes           string     `gorm:"type:text"`
	Status          string     `gorm:"type:text;not null;default:pending"`
	DueOn           *time.Time `gorm:"type:date"`
	RolloverEnabled bool       `gorm:"not null;default:false"`
	CompletedAt     *time.Time
	CompletedByUserId *uuid.UUID `gorm:"type:uuid"`
	DeletedAt       *time.Time `gorm:"index"`
	CreatedAt       time.Time  `gorm:"not null"`
	UpdatedAt       time.Time  `gorm:"not null"`
}

func (Entity) TableName() string { return "tasks" }

func Migration(db *gorm.DB) error { return db.AutoMigrate(&Entity{}) }

func Make(e Entity) (Model, error) {
	return Model{
		id: e.Id, tenantID: e.TenantId, householdID: e.HouseholdId,
		title: e.Title, notes: e.Notes, status: e.Status,
		dueOn: e.DueOn, rolloverEnabled: e.RolloverEnabled,
		completedAt: e.CompletedAt, completedByUID: e.CompletedByUserId,
		deletedAt: e.DeletedAt, createdAt: e.CreatedAt, updatedAt: e.UpdatedAt,
	}, nil
}
