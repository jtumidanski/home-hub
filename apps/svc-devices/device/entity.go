package device

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Entity represents the database model for devices.
// This is separate from the domain Model to isolate persistence concerns.
type Entity struct {
	Id          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name        string    `gorm:"type:varchar(255);not null"`
	Type        string    `gorm:"type:varchar(50);not null"` // "kiosk", future: "mobile", "tablet"
	HouseholdId uuid.UUID `gorm:"type:uuid;not null;index"`
	// Timestamps
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

// TableName specifies the table name for the devices entity
func (e Entity) TableName() string {
	return "devices"
}

// Make transforms a database Entity into a domain Model
func Make(e Entity) (Model, error) {
	return Model{
		id:          e.Id,
		name:        e.Name,
		deviceType:  e.Type,
		householdId: e.HouseholdId,
		createdAt:   e.CreatedAt,
		updatedAt:   e.UpdatedAt,
	}, nil
}

// ToEntity transforms a domain Model into a database Entity
func (m Model) ToEntity() Entity {
	return Entity{
		Id:          m.id,
		Name:        m.name,
		Type:        m.deviceType,
		HouseholdId: m.householdId,
		CreatedAt:   m.createdAt,
		UpdatedAt:   m.updatedAt,
	}
}

// Migration runs the auto-migration for the devices table
func Migration() func(db *gorm.DB) error {
	return func(db *gorm.DB) error {
		return db.AutoMigrate(&Entity{})
	}
}
