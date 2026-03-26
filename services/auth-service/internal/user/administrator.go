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
