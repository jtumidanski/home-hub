package preference

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entity struct {
	Id                uuid.UUID  `gorm:"type:uuid;primaryKey"`
	TenantId          uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_tenant_user"`
	UserId            uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_tenant_user"`
	Theme             string     `gorm:"type:text;not null;default:light"`
	ActiveHouseholdId *uuid.UUID `gorm:"type:uuid"`
	CreatedAt         time.Time  `gorm:"not null"`
	UpdatedAt         time.Time  `gorm:"not null"`
}

func (Entity) TableName() string { return "preferences" }

func Migration(db *gorm.DB) error { return db.AutoMigrate(&Entity{}) }

func (m Model) ToEntity() Entity {
	return Entity{
		Id:                m.id,
		TenantId:          m.tenantID,
		UserId:            m.userID,
		Theme:             m.theme,
		ActiveHouseholdId: m.activeHouseholdID,
		CreatedAt:         m.createdAt,
		UpdatedAt:         m.updatedAt,
	}
}

func Make(e Entity) (Model, error) {
	return NewBuilder().
		SetId(e.Id).
		SetTenantID(e.TenantId).
		SetUserID(e.UserId).
		SetTheme(e.Theme).
		SetActiveHouseholdID(e.ActiveHouseholdId).
		SetCreatedAt(e.CreatedAt).
		SetUpdatedAt(e.UpdatedAt).
		Build()
}
