package restoration

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func create(db *gorm.DB, tenantID, householdID, taskID, userID uuid.UUID) (Entity, error) {
	e := Entity{
		Id:              uuid.New(),
		TenantId:        tenantID,
		HouseholdId:     householdID,
		TaskId:          taskID,
		CreatedByUserId: userID,
		CreatedAt:       time.Now().UTC(),
	}
	if err := db.Create(&e).Error; err != nil {
		return Entity{}, err
	}
	return e, nil
}
