package membership

import (
	"github.com/google/uuid"
	database "github.com/jtumidanski/home-hub/shared/go/database"
	"gorm.io/gorm"
)

func getByID(id uuid.UUID) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	})
}

// getByUser returns memberships for a user.
// Tenant filtering is automatic via GORM callbacks when db.WithContext(ctx) is used.
func getByUser(userID uuid.UUID) database.EntityProvider[[]Entity] {
	return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ?", userID)
	})
}

func getByHouseholdAndUser(householdID, userID uuid.UUID) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("household_id = ? AND user_id = ?", householdID, userID)
	})
}

func getByHousehold(householdID uuid.UUID) database.EntityProvider[[]Entity] {
	return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("household_id = ?", householdID)
	})
}

func countOwnersByHousehold(db *gorm.DB, householdID uuid.UUID) (int64, error) {
	var count int64
	err := db.Model(&Entity{}).
		Where("household_id = ? AND role = 'owner'", householdID).
		Count(&count).Error
	return count, err
}
