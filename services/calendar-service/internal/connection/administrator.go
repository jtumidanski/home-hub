package connection

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func create(db *gorm.DB, tenantID, householdID, userID uuid.UUID, provider, email, accessToken, refreshToken, userDisplayName, userColor string, tokenExpiry time.Time) (Entity, error) {
	now := time.Now().UTC()
	e := Entity{
		Id:              uuid.New(),
		TenantId:        tenantID,
		HouseholdId:     householdID,
		UserId:          userID,
		Provider:        provider,
		Status:          "connected",
		Email:           email,
		AccessToken:     accessToken,
		RefreshToken:    refreshToken,
		TokenExpiry:     tokenExpiry,
		UserDisplayName: userDisplayName,
		UserColor:       userColor,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := db.Create(&e).Error; err != nil {
		return Entity{}, err
	}
	return e, nil
}

func updateStatus(db *gorm.DB, id uuid.UUID, status string) error {
	return db.Model(&Entity{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     status,
		"updated_at": time.Now().UTC(),
	}).Error
}

func updateTokens(db *gorm.DB, id uuid.UUID, accessToken string, tokenExpiry time.Time) error {
	return db.Model(&Entity{}).Where("id = ?", id).Updates(map[string]interface{}{
		"access_token": accessToken,
		"token_expiry": tokenExpiry,
		"updated_at":   time.Now().UTC(),
	}).Error
}

func updateSyncInfo(db *gorm.DB, id uuid.UUID, eventCount int) error {
	now := time.Now().UTC()
	return db.Model(&Entity{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_sync_at":          now,
		"last_sync_event_count": eventCount,
		"updated_at":            now,
	}).Error
}

func deleteByID(db *gorm.DB, id uuid.UUID) error {
	return db.Where("id = ?", id).Delete(&Entity{}).Error
}
