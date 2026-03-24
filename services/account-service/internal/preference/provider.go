package preference

import (
	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/model"
	"gorm.io/gorm"
)

// getByUser returns a preference for a user.
// Tenant filtering is automatic via GORM callbacks when db.WithContext(ctx) is used.
func getByUser(userID uuid.UUID) func(db *gorm.DB) model.Provider[Entity] {
	return func(db *gorm.DB) model.Provider[Entity] {
		return func() (Entity, error) {
			var result Entity
			err := db.Where("user_id = ?", userID).First(&result).Error
			return result, err
		}
	}
}

func getByID(id uuid.UUID) func(db *gorm.DB) model.Provider[Entity] {
	return func(db *gorm.DB) model.Provider[Entity] {
		return func() (Entity, error) {
			var result Entity
			err := db.Where("id = ?", id).First(&result).Error
			return result, err
		}
	}
}
