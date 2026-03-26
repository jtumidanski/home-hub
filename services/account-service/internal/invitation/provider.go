package invitation

import (
	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"gorm.io/gorm"
)

func getByID(id uuid.UUID) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	})
}

// getByHouseholdPending returns pending, non-expired invitations for a household.
func getByHouseholdPending(householdID uuid.UUID) database.EntityProvider[[]Entity] {
	return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("household_id = ? AND status = 'pending' AND expires_at > NOW()", householdID)
	})
}

// getByEmailPending returns pending, non-expired invitations for an email.
// This is intended to be used with WithoutTenantFilter since the user may not have a tenant yet.
func getByEmailPending(email string) database.EntityProvider[[]Entity] {
	return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("email = ? AND status = 'pending' AND expires_at > NOW()", email)
	})
}

// getByHouseholdAndEmailPending checks if a pending invitation already exists.
func getByHouseholdAndEmailPending(householdID uuid.UUID, email string) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("household_id = ? AND email = ? AND status = 'pending' AND expires_at > NOW()", householdID, email)
	})
}

// countByEmailPending returns the count of pending, non-expired invitations for an email.
func countByEmailPending(db *gorm.DB, email string) (int64, error) {
	var count int64
	err := db.Model(&Entity{}).
		Where("email = ? AND status = 'pending' AND expires_at > NOW()", email).
		Count(&count).Error
	return count, err
}
