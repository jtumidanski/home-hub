package membership

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func create(db *gorm.DB, tenantID, householdID, userID uuid.UUID, role string) (Entity, error) {
	now := time.Now().UTC()
	e := Entity{
		Id:          uuid.New(),
		TenantId:    tenantID,
		HouseholdId: householdID,
		UserId:      userID,
		Role:        role,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.Create(&e).Error; err != nil {
		return Entity{}, err
	}
	return e, nil
}

func updateRole(db *gorm.DB, id uuid.UUID, role string) (Entity, error) {
	var e Entity
	if err := db.Where("id = ?", id).First(&e).Error; err != nil {
		return Entity{}, err
	}
	e.Role = role
	e.UpdatedAt = time.Now().UTC()
	if err := db.Save(&e).Error; err != nil {
		return Entity{}, err
	}
	return e, nil
}

func deleteByID(db *gorm.DB, id uuid.UUID) error {
	return db.Where("id = ?", id).Delete(&Entity{}).Error
}
