package wishlist

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entity struct {
	Id               uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantId         uuid.UUID `gorm:"type:uuid;not null;index:idx_wish_list_items_tenant_household_votes"`
	HouseholdId      uuid.UUID `gorm:"type:uuid;not null;index:idx_wish_list_items_tenant_household_votes"`
	Name             string    `gorm:"type:varchar(255);not null"`
	PurchaseLocation *string   `gorm:"type:varchar(255)"`
	Urgency          string    `gorm:"type:varchar(20);not null;default:want"`
	VoteCount        int       `gorm:"not null;default:0;index:idx_wish_list_items_tenant_household_votes"`
	CreatedBy        uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt        time.Time `gorm:"not null;index:idx_wish_list_items_tenant_household_votes"`
	UpdatedAt        time.Time `gorm:"not null"`
}

func (Entity) TableName() string { return "wish_list_items" }

func Migration(db *gorm.DB) error {
	return db.AutoMigrate(&Entity{})
}

func Make(e Entity) (Model, error) {
	b := NewBuilder().
		SetId(e.Id).
		SetTenantID(e.TenantId).
		SetHouseholdID(e.HouseholdId).
		SetName(e.Name).
		SetUrgency(e.Urgency).
		SetVoteCount(e.VoteCount).
		SetCreatedBy(e.CreatedBy).
		SetCreatedAt(e.CreatedAt).
		SetUpdatedAt(e.UpdatedAt)
	if e.PurchaseLocation != nil {
		b.SetPurchaseLocation(e.PurchaseLocation)
	}
	return b.Build()
}
