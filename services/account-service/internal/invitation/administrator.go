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

func updateStatus(db *gorm.DB, id uuid.UUID, status string) error {
	return db.Model(&Entity{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now().UTC(),
		}).Error
}
