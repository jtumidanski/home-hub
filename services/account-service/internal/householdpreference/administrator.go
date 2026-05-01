package householdpreference

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func insert(db *gorm.DB, tenantID, userID, householdID uuid.UUID) (Entity, error) {
	now := time.Now().UTC()
	e := Entity{
		Id:          uuid.New(),
		TenantId:    tenantID,
		UserId:      userID,
		HouseholdId: householdID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.Create(&e).Error; err != nil {
		return Entity{}, err
	}
	return e, nil
}

// markKioskSeeded sets kiosk_dashboard_seeded = TRUE for the row with the
// given id. The flag is write-once-true; this is intentionally idempotent.
func markKioskSeeded(db *gorm.DB, id uuid.UUID) (Entity, error) {
	now := time.Now().UTC()
	if err := db.Exec(
		"UPDATE household_preferences SET kiosk_dashboard_seeded = TRUE, updated_at = ? WHERE id = ?",
		now, id,
	).Error; err != nil {
		return Entity{}, err
	}
	var e Entity
	if err := db.Where("id = ?", id).First(&e).Error; err != nil {
		return Entity{}, err
	}
	return e, nil
}

// updateFields sets default_dashboard_id via a raw UPDATE statement so a nil
// pointer reliably writes NULL across dialects. GORM's Updates(map) drops nil
// entries silently, which is the wrong behavior for "clear the field".
func updateFields(db *gorm.DB, id uuid.UUID, defaultDashboardID *uuid.UUID) (Entity, error) {
	now := time.Now().UTC()
	var dd interface{}
	if defaultDashboardID != nil {
		dd = *defaultDashboardID
	} else {
		dd = nil
	}
	if err := db.Exec(
		"UPDATE household_preferences SET default_dashboard_id = ?, updated_at = ? WHERE id = ?",
		dd, now, id,
	).Error; err != nil {
		return Entity{}, err
	}
	var e Entity
	if err := db.Where("id = ?", id).First(&e).Error; err != nil {
		return Entity{}, err
	}
	return e, nil
}
