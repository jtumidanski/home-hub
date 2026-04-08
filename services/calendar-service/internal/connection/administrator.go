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
		WriteAccess:     true,
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

func updateTokensAndWriteAccess(db *gorm.DB, id uuid.UUID, accessToken, refreshToken string, tokenExpiry time.Time, writeAccess bool) error {
	return db.Model(&Entity{}).Where("id = ?", id).Updates(map[string]interface{}{
		"access_token":         accessToken,
		"refresh_token":        refreshToken,
		"token_expiry":         tokenExpiry,
		"write_access":         writeAccess,
		"status":               "connected",
		"error_code":           gorm.Expr("NULL"),
		"error_message":        gorm.Expr("NULL"),
		"last_error_at":        gorm.Expr("NULL"),
		"consecutive_failures": 0,
		"updated_at":           time.Now().UTC(),
	}).Error
}

func updateSyncAttempt(db *gorm.DB, id uuid.UUID, at time.Time) error {
	return db.Model(&Entity{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_sync_attempt_at": at,
		"updated_at":           at,
	}).Error
}

func updateSyncSuccess(db *gorm.DB, id uuid.UUID, eventCount int, at time.Time) error {
	return db.Model(&Entity{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":                "connected",
		"last_sync_at":          at,
		"last_sync_attempt_at":  at,
		"last_sync_event_count": eventCount,
		"error_code":            gorm.Expr("NULL"),
		"error_message":         gorm.Expr("NULL"),
		"last_error_at":         gorm.Expr("NULL"),
		"consecutive_failures":  0,
		"updated_at":            at,
	}).Error
}

// updateSyncFailure performs the entire failure state transition in a single
// atomic UPDATE so the counter increment, threshold check, and status
// transition cannot race with a concurrent writer.
func updateSyncFailure(db *gorm.DB, id uuid.UUID, code, message string, at time.Time, hard bool) error {
	updates := map[string]interface{}{
		"error_code":           code,
		"error_message":        message,
		"last_error_at":        at,
		"last_sync_attempt_at": at,
		"updated_at":           at,
	}
	if hard {
		updates["status"] = "disconnected"
		// Equivalent to GREATEST(consecutive_failures + 1, threshold) but
		// dialect-agnostic so tests on sqlite behave the same as Postgres prod.
		updates["consecutive_failures"] = gorm.Expr(
			"CASE WHEN consecutive_failures + 1 < ? THEN ? ELSE consecutive_failures + 1 END",
			FailureEscalationThreshold, FailureEscalationThreshold,
		)
	} else {
		updates["status"] = gorm.Expr(
			"CASE WHEN consecutive_failures + 1 >= ? THEN 'error' ELSE status END",
			FailureEscalationThreshold,
		)
		updates["consecutive_failures"] = gorm.Expr("consecutive_failures + 1")
	}
	return db.Model(&Entity{}).Where("id = ?", id).Updates(updates).Error
}

func clearErrorState(db *gorm.DB, id uuid.UUID) error {
	return db.Model(&Entity{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":               "connected",
		"error_code":           gorm.Expr("NULL"),
		"error_message":        gorm.Expr("NULL"),
		"last_error_at":        gorm.Expr("NULL"),
		"consecutive_failures": 0,
		"updated_at":           time.Now().UTC(),
	}).Error
}

func deleteByID(db *gorm.DB, id uuid.UUID) error {
	return db.Where("id = ?", id).Delete(&Entity{}).Error
}
