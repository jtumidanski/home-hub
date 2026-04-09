package locationofinterest

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entity struct {
	Id          uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantId    uuid.UUID `gorm:"type:uuid;not null;index:idx_loi_tenant_household"`
	HouseholdId uuid.UUID `gorm:"type:uuid;not null;index:idx_loi_tenant_household"`
	Label       *string   `gorm:"type:varchar(64)"`
	PlaceName   string    `gorm:"type:text;not null"`
	Latitude    float64   `gorm:"type:double precision;not null"`
	Longitude   float64   `gorm:"type:double precision;not null"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}

func (Entity) TableName() string { return "locations_of_interest" }

func Migration(db *gorm.DB) error {
	return db.AutoMigrate(&Entity{})
}

func (m Model) ToEntity() Entity {
	return Entity{
		Id:          m.id,
		TenantId:    m.tenantID,
		HouseholdId: m.householdID,
		Label:       m.label,
		PlaceName:   m.placeName,
		Latitude:    m.latitude,
		Longitude:   m.longitude,
		CreatedAt:   m.createdAt,
		UpdatedAt:   m.updatedAt,
	}
}

func Make(e Entity) (Model, error) {
	return NewBuilder().
		SetId(e.Id).
		SetTenantID(e.TenantId).
		SetHouseholdID(e.HouseholdId).
		SetLabel(e.Label).
		SetPlaceName(e.PlaceName).
		SetLatitude(e.Latitude).
		SetLongitude(e.Longitude).
		SetCreatedAt(e.CreatedAt).
		SetUpdatedAt(e.UpdatedAt).
		Build()
}
