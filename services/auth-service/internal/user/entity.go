package user

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Entity is the GORM entity for auth.users.
type Entity struct {
	Id                uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email             string    `gorm:"type:text;not null;uniqueIndex"`
	DisplayName       string    `gorm:"type:text"`
	GivenName         string    `gorm:"type:text"`
	FamilyName        string    `gorm:"type:text"`
	AvatarURL         string    `gorm:"type:text"`
	ProviderAvatarURL string    `gorm:"type:text"`
	CreatedAt         time.Time `gorm:"not null"`
	UpdatedAt         time.Time `gorm:"not null"`
}

func (Entity) TableName() string {
	return "users"
}

// Migration returns the GORM auto-migration function for the user entity.
func Migration(db *gorm.DB) error {
	return db.AutoMigrate(&Entity{})
}

// MigrateAvatarData copies avatar_url to provider_avatar_url for existing rows
// where provider_avatar_url is empty and avatar_url is set, then clears avatar_url.
// This is idempotent and safe to run multiple times.
func MigrateAvatarData(db *gorm.DB) error {
	return db.Exec(`
		UPDATE users
		SET provider_avatar_url = avatar_url, avatar_url = ''
		WHERE provider_avatar_url = '' AND avatar_url != '' AND avatar_url NOT LIKE 'dicebear:%'
	`).Error
}

// Make converts an Entity to a Model.
func Make(e Entity) (Model, error) {
	return NewBuilder().
		SetId(e.Id).
		SetEmail(e.Email).
		SetDisplayName(e.DisplayName).
		SetGivenName(e.GivenName).
		SetFamilyName(e.FamilyName).
		SetAvatarURL(e.AvatarURL).
		SetProviderAvatarURL(e.ProviderAvatarURL).
		SetCreatedAt(e.CreatedAt).
		SetUpdatedAt(e.UpdatedAt).
		Build()
}
