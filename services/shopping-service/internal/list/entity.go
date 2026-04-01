package list

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entity struct {
	Id          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	TenantId    uuid.UUID  `gorm:"type:uuid;not null;index:idx_shopping_list_tenant_household_status"`
	HouseholdId uuid.UUID  `gorm:"type:uuid;not null;index:idx_shopping_list_tenant_household_status"`
	Name        string     `gorm:"type:varchar(255);not null"`
	Status      string     `gorm:"type:varchar(20);not null;default:active;index:idx_shopping_list_tenant_household_status"`
	ArchivedAt  *time.Time `gorm:""`
	CreatedBy   uuid.UUID  `gorm:"type:uuid;not null"`
	CreatedAt   time.Time  `gorm:"not null"`
	UpdatedAt   time.Time  `gorm:"not null"`
}

func (Entity) TableName() string { return "shopping_lists" }

func Migration(db *gorm.DB) error {
	return db.AutoMigrate(&Entity{})
}

func Make(e Entity) (Model, error) {
	b := NewBuilder().
		SetId(e.Id).
		SetTenantID(e.TenantId).
		SetHouseholdID(e.HouseholdId).
		SetName(e.Name).
		SetStatus(e.Status).
		SetCreatedBy(e.CreatedBy).
		SetCreatedAt(e.CreatedAt).
		SetUpdatedAt(e.UpdatedAt)
	if e.ArchivedAt != nil {
		b.SetArchivedAt(e.ArchivedAt)
	}
	return b.Build()
}
