package invitation

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func create(db *gorm.DB, tenantID, householdID uuid.UUID, email, role string, invitedBy uuid.UUID) (Entity, error) {
	now := time.Now().UTC()
	e := Entity{
		Id:          uuid.New(),
		TenantId:    tenantID,
		HouseholdId: householdID,
		Email:       email,
		Role:        role,
		Status:      "pending",
		InvitedBy:   invitedBy,
		ExpiresAt:   now.Add(7 * 24 * time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.Create(&e).Error; err != nil {
		return Entity{}, err
	}
	return e, nil
}

func updateStatus(db *gorm.DB, id uuid.UUID, status string) (Entity, error) {
	var e Entity
	if err := db.Where("id = ?", id).First(&e).Error; err != nil {
		return Entity{}, err
	}
	e.Status = status
	e.UpdatedAt = time.Now().UTC()
	if err := db.Save(&e).Error; err != nil {
		return Entity{}, err
	}
	return e, nil
}
