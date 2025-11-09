package user

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Entity represents the database model for users.
// This is separate from the domain Model to isolate persistence concerns.
type Entity struct {
	Id          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Email       string     `gorm:"type:varchar(255);not null;uniqueIndex:idx_users_email"`
	DisplayName string     `gorm:"type:varchar(255);not null"`
	HouseholdId *uuid.UUID `gorm:"type:uuid;index:idx_users_household_id"`
	CreatedAt   time.Time  `gorm:"not null"`
	UpdatedAt   time.Time  `gorm:"not null"`
}

// TableName specifies the table name for the users entity
func (e Entity) TableName() string {
	return "users"
}

// Make transforms a database Entity into a domain Model
func Make(e Entity) (Model, error) {
	return Model{
		id:          e.Id,
		email:       e.Email,
		displayName: e.DisplayName,
		householdId: e.HouseholdId,
		createdAt:   e.CreatedAt,
		updatedAt:   e.UpdatedAt,
	}, nil
}

// ToEntity transforms a domain Model into a database Entity
func (m Model) ToEntity() Entity {
	return Entity{
		Id:          m.id,
		Email:       m.email,
		DisplayName: m.displayName,
		HouseholdId: m.householdId,
		CreatedAt:   m.createdAt,
		UpdatedAt:   m.updatedAt,
	}
}

// Migration runs the auto-migration for the users table
func Migration() func(db *gorm.DB) error {
	return func(db *gorm.DB) error {
		return db.AutoMigrate(&Entity{})
	}
}
