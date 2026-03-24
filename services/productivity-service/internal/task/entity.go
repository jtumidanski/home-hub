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
	return NewBuilder().
		SetId(e.Id).
		SetTenantID(e.TenantId).
		SetHouseholdID(e.HouseholdId).
		SetTitle(e.Title).
		SetNotes(e.Notes).
		SetStatus(e.Status).
		SetDueOn(e.DueOn).
		SetRolloverEnabled(e.RolloverEnabled).
		SetCompletedAt(e.CompletedAt).
		SetCompletedByUID(e.CompletedByUserId).
		SetDeletedAt(e.DeletedAt).
		SetCreatedAt(e.CreatedAt).
		SetUpdatedAt(e.UpdatedAt).
		Build()
}
