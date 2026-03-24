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

func Make(e Entity) (Model, error) {
	return Model{
		id: e.Id, tenantID: e.TenantId, householdID: e.HouseholdId,
		userID: e.UserId, role: e.Role,
		createdAt: e.CreatedAt, updatedAt: e.UpdatedAt,
	}, nil
}
