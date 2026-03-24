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

func Make(e Entity) (Model, error) {
	return Model{
		id: e.Id, tenantID: e.TenantId, userID: e.UserId,
		theme: e.Theme, activeHouseholdID: e.ActiveHouseholdId,
		createdAt: e.CreatedAt, updatedAt: e.UpdatedAt,
	}, nil
}
