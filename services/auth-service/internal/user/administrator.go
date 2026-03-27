package user

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func create(db *gorm.DB, email, displayName, givenName, familyName, avatarURL string) (Entity, error) {
	now := time.Now().UTC()
	e := Entity{
		Id:                uuid.New(),
		Email:             email,
		DisplayName:       displayName,
		GivenName:         givenName,
		FamilyName:        familyName,
		ProviderAvatarURL: avatarURL,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := db.Create(&e).Error; err != nil {
		return Entity{}, err
	}
	return e, nil
}

func updateAvatarURL(db *gorm.DB, id uuid.UUID, url string) error {
	return db.Model(&Entity{}).Where("id = ?", id).Update("avatar_url", url).Error
}

func updateProviderAvatarURL(db *gorm.DB, id uuid.UUID, url string) error {
	return db.Model(&Entity{}).Where("id = ?", id).Update("provider_avatar_url", url).Error
}
