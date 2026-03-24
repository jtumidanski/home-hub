package externalidentity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func create(db *gorm.DB, userID uuid.UUID, provider, subject string) (Entity, error) {
	now := time.Now().UTC()
	e := Entity{
		Id:              uuid.New(),
		UserId:          userID,
		Provider:        provider,
		ProviderSubject: subject,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := db.Create(&e).Error; err != nil {
		return Entity{}, err
	}
	return e, nil
}
