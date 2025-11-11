package role

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Entity represents the database model for user roles.
// This is a junction table between users and their assigned roles.
type Entity struct {
	UserId uuid.UUID `gorm:"type:uuid;primaryKey;index:idx_user_roles_user_id"`
	Role   string    `gorm:"type:varchar(100);primaryKey;not null"`
}

// TableName specifies the table name for the user_roles entity
func (e Entity) TableName() string {
	return "user_roles"
}

// Make transforms a database Entity into a domain Model
func Make(e Entity) (Model, error) {
	return Model{
		userId: e.UserId,
		role:   e.Role,
	}, nil
}

// ToEntity transforms a domain Model into a database Entity
func (m Model) ToEntity() Entity {
	return Entity{
		UserId: m.userId,
		Role:   m.role,
	}
}

// Migration runs the auto-migration for the user_roles table
func Migration() func(db *gorm.DB) error {
	return func(db *gorm.DB) error {
		return db.AutoMigrate(&Entity{})
	}
}
