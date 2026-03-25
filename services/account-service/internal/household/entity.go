package household

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entity struct {
	Id           uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantId     uuid.UUID `gorm:"type:uuid;not null;index"`
	Name         string    `gorm:"type:text;not null"`
	Timezone     string    `gorm:"type:text;not null"`
	Units        string    `gorm:"type:text;not null"`
	Latitude     *float64  `gorm:"type:double precision"`
	Longitude    *float64  `gorm:"type:double precision"`
	LocationName *string   `gorm:"type:text"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
}

func (Entity) TableName() string { return "households" }

func Migration(db *gorm.DB) error { return db.AutoMigrate(&Entity{}) }

func (m Model) ToEntity() Entity {
	return Entity{
		Id:           m.id,
		TenantId:     m.tenantID,
		Name:         m.name,
		Timezone:     m.timezone,
		Units:        m.units,
		Latitude:     m.latitude,
		Longitude:    m.longitude,
		LocationName: m.locationName,
		CreatedAt:    m.createdAt,
		UpdatedAt:    m.updatedAt,
	}
}

func Make(e Entity) (Model, error) {
	return NewBuilder().
		SetId(e.Id).
		SetTenantID(e.TenantId).
		SetName(e.Name).
		SetTimezone(e.Timezone).
		SetUnits(e.Units).
		SetLatitude(e.Latitude).
		SetLongitude(e.Longitude).
		SetLocationName(e.LocationName).
		SetCreatedAt(e.CreatedAt).
		SetUpdatedAt(e.UpdatedAt).
		Build()
}
