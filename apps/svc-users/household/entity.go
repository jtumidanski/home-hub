package household

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Entity represents the database model for households.
// This is separate from the domain Model to isolate persistence concerns.
type Entity struct {
	Id        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name      string    `gorm:"type:varchar(255);not null"`
	// Location coordinates for weather APIs
	Latitude  *float64  `gorm:"type:double precision"`
	Longitude *float64  `gorm:"type:double precision"`
	Timezone  *string   `gorm:"type:varchar(100)"`
	// Timestamps
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

// TableName specifies the table name for the households entity
func (e Entity) TableName() string {
	return "households"
}

// Make transforms a database Entity into a domain Model
func Make(e Entity) (Model, error) {
	return Model{
		id:        e.Id,
		name:      e.Name,
		latitude:  e.Latitude,
		longitude: e.Longitude,
		timezone:  e.Timezone,
		createdAt: e.CreatedAt,
		updatedAt: e.UpdatedAt,
	}, nil
}

// ToEntity transforms a domain Model into a database Entity
func (m Model) ToEntity() Entity {
	return Entity{
		Id:        m.id,
		Name:      m.name,
		Latitude:  m.latitude,
		Longitude: m.longitude,
		Timezone:  m.timezone,
		CreatedAt: m.createdAt,
		UpdatedAt: m.updatedAt,
	}
}

// Migration runs the auto-migration for the households table
func Migration() func(db *gorm.DB) error {
	return func(db *gorm.DB) error {
		return db.AutoMigrate(&Entity{})
	}
}
