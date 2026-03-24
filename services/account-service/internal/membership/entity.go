package membership

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entity struct {
	Id          uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantId    uuid.UUID `gorm:"type:uuid;not null;index"`
	HouseholdId uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_household_user"`
	UserId      uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_household_user"`
	Role        string    `gorm:"type:text;not null"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}

func (Entity) TableName() string { return "memberships" }

func Migration(db *gorm.DB) error { return db.AutoMigrate(&Entity{}) }

func (m Model) ToEntity() Entity {
	return Entity{
		Id:          m.id,
		TenantId:    m.tenantID,
		HouseholdId: m.householdID,
		UserId:      m.userID,
		Role:        m.role,
		CreatedAt:   m.createdAt,
		UpdatedAt:   m.updatedAt,
	}
}

func Make(e Entity) (Model, error) {
	return NewBuilder().
		SetId(e.Id).
		SetTenantID(e.TenantId).
		SetHouseholdID(e.HouseholdId).
		SetUserID(e.UserId).
		SetRole(e.Role).
		SetCreatedAt(e.CreatedAt).
		SetUpdatedAt(e.UpdatedAt).
		Build()
}
