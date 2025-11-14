package preference

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Entity represents the database model for device preferences.
// This is separate from the domain Model to isolate persistence concerns.
type Entity struct {
	Id              uuid.UUID `gorm:"type:uuid;primaryKey"`
	DeviceId        uuid.UUID `gorm:"type:uuid;not null;uniqueIndex;constraint:OnDelete:CASCADE"`
	Theme           string    `gorm:"type:varchar(20);default:'dark'"`   // "light" or "dark"
	TemperatureUnit string    `gorm:"type:varchar(20);default:'household'"` // "household", "F", or "C"
	// Timestamps
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

// TableName specifies the table name for the device_preferences entity
func (e Entity) TableName() string {
	return "device_preferences"
}

// Make transforms a database Entity into a domain Model
func Make(e Entity) (Model, error) {
	return Model{
		id:              e.Id,
		deviceId:        e.DeviceId,
		theme:           e.Theme,
		temperatureUnit: e.TemperatureUnit,
		createdAt:       e.CreatedAt,
		updatedAt:       e.UpdatedAt,
	}, nil
}

// ToEntity transforms a domain Model into a database Entity
func (m Model) ToEntity() Entity {
	return Entity{
		Id:              m.id,
		DeviceId:        m.deviceId,
		Theme:           m.theme,
		TemperatureUnit: m.temperatureUnit,
		CreatedAt:       m.createdAt,
		UpdatedAt:       m.updatedAt,
	}
}

// Migration runs the auto-migration for the device_preferences table
func Migration() func(db *gorm.DB) error {
	return func(db *gorm.DB) error {
		return db.AutoMigrate(&Entity{})
	}
}
