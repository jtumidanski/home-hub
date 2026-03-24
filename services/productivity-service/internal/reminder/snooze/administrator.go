package snooze

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func create(db *gorm.DB, tenantID, householdID, reminderID, userID uuid.UUID, durationMinutes int, snoozedUntil time.Time) (Entity, error) {
	e := Entity{
		Id:              uuid.New(),
		TenantId:        tenantID,
		HouseholdId:     householdID,
		ReminderId:      reminderID,
		DurationMinutes: durationMinutes,
		SnoozedUntil:    snoozedUntil,
		CreatedByUserId: userID,
		CreatedAt:       time.Now().UTC(),
	}
	if err := db.Create(&e).Error; err != nil {
		return Entity{}, err
	}
	return e, nil
}
