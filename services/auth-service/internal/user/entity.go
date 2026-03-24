package user

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Entity is the GORM entity for auth.users.
type Entity struct {
	Id          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email       string    `gorm:"type:text;not null;uniqueIndex"`
	DisplayName string    `gorm:"type:text"`
	GivenName   string    `gorm:"type:text"`
	FamilyName  string    `gorm:"type:text"`
	AvatarURL   string    `gorm:"type:text"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}

func (Entity) TableName() string {
	return "users"
}

// Migration returns the GORM auto-migration function for the user entity.
func Migration(db *gorm.DB) error {
	return db.AutoMigrate(&Entity{})
}

// Make converts an Entity to a Model.
func Make(e Entity) (Model, error) {
	return Model{
		id:          e.Id,
		email:       e.Email,
		displayName: e.DisplayName,
		givenName:   e.GivenName,
		familyName:  e.FamilyName,
		avatarURL:   e.AvatarURL,
		createdAt:   e.CreatedAt,
		updatedAt:   e.UpdatedAt,
	}, nil
}
