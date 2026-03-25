package dismissal

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func create(db *gorm.DB, tenantID, householdID, reminderID, userID uuid.UUID) (Entity, error) {
	e := Entity{
		Id:              uuid.New(),
		TenantId:        tenantID,
		HouseholdId:     householdID,
		ReminderId:      reminderID,
		CreatedByUserId: userID,
		CreatedAt:       time.Now().UTC(),
	}
	if err := db.Create(&e).Error; err != nil {
		return Entity{}, err
	}
	return e, nil
}
