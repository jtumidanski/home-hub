package householdpreference

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entity struct {
	Id                 uuid.UUID  `gorm:"type:uuid;primaryKey"`
	TenantId           uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_hp_tup"`
	UserId             uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_hp_tup"`
	HouseholdId        uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_hp_tup"`
	DefaultDashboardId *uuid.UUID `gorm:"type:uuid"`
	CreatedAt          time.Time  `gorm:"not null"`
	UpdatedAt          time.Time  `gorm:"not null"`
}

func (Entity) TableName() string { return "household_preferences" }

func Migration(db *gorm.DB) error { return db.AutoMigrate(&Entity{}) }

func (m Model) ToEntity() Entity {
	return Entity{
		Id:                 m.id,
		TenantId:           m.tenantID,
		UserId:             m.userID,
		HouseholdId:        m.householdID,
		DefaultDashboardId: m.defaultDashboardID,
		CreatedAt:          m.createdAt,
		UpdatedAt:          m.updatedAt,
	}
}

func Make(e Entity) (Model, error) {
	return NewBuilder().
		SetId(e.Id).
		SetTenantID(e.TenantId).
		SetUserID(e.UserId).
		SetHouseholdID(e.HouseholdId).
		SetDefaultDashboardID(e.DefaultDashboardId).
		SetCreatedAt(e.CreatedAt).
		SetUpdatedAt(e.UpdatedAt).
		Build()
}
