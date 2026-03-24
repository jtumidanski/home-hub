package household

import (
	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/model"
	"gorm.io/gorm"
)

func getByID(id uuid.UUID) func(db *gorm.DB) model.Provider[Entity] {
	return func(db *gorm.DB) model.Provider[Entity] {
		return func() (Entity, error) {
			var result Entity
			err := db.Where("id = ?", id).First(&result).Error
			return result, err
		}
	}
}

// getAll returns all households for the current tenant.
// Tenant filtering is automatic via GORM callbacks when db.WithContext(ctx) is used.
func getAll() func(db *gorm.DB) model.Provider[[]Entity] {
	return func(db *gorm.DB) model.Provider[[]Entity] {
		return func() ([]Entity, error) {
			var results []Entity
			err := db.Find(&results).Error
			return results, err
		}
	}
}
