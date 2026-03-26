package oauthstate

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func create(db *gorm.DB, tenantID, householdID, userID uuid.UUID, redirectURI string) (Entity, error) {
	now := time.Now().UTC()
	e := Entity{
		Id:          uuid.New(),
		TenantId:    tenantID,
		HouseholdId: householdID,
		UserId:      userID,
		RedirectUri: redirectURI,
		ExpiresAt:   now.Add(10 * time.Minute),
		CreatedAt:   now,
	}
	if err := db.Create(&e).Error; err != nil {
		return Entity{}, err
	}
	return e, nil
}

func deleteByID(db *gorm.DB, id uuid.UUID) error {
	return db.Where("id = ?", id).Delete(&Entity{}).Error
}

func deleteExpired(db *gorm.DB) error {
	return db.Where("expires_at < ?", time.Now().UTC()).Delete(&Entity{}).Error
}
