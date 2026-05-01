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
	res := db.Model(&Entity{}).Where("id = ?", id).Updates(map[string]any{
		"kiosk_dashboard_seeded": true,
		"updated_at":             now,
	})
	if res.Error != nil {
		return Entity{}, res.Error
	}
	if res.RowsAffected == 0 {
		return Entity{}, gorm.ErrRecordNotFound
	}
	var e Entity
	if err := db.Where("id = ?", id).First(&e).Error; err != nil {
		return Entity{}, err
	}
	return e, nil
}

// updateFields sets default_dashboard_id via a GORM model-scoped UPDATE so the
// tenant callback can inject WHERE tenant_id = ? automatically. The map-form
// of Updates writes every key in the map, including nil values as SQL NULL,
// which preserves the "clear the field on nil" semantics.
func updateFields(db *gorm.DB, id uuid.UUID, defaultDashboardID *uuid.UUID) (Entity, error) {
	now := time.Now().UTC()
	var dd interface{}
	if defaultDashboardID != nil {
		dd = *defaultDashboardID
	} else {
		dd = gorm.Expr("NULL")
	}
	res := db.Model(&Entity{}).Where("id = ?", id).Updates(map[string]any{
		"default_dashboard_id": dd,
		"updated_at":           now,
	})
	if res.Error != nil {
		return Entity{}, res.Error
	}
	if res.RowsAffected == 0 {
		return Entity{}, gorm.ErrRecordNotFound
	}
	var e Entity
	if err := db.Where("id = ?", id).First(&e).Error; err != nil {
		return Entity{}, err
	}
	return e, nil
}
