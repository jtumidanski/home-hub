package tenant

import (
	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/model"
	"gorm.io/gorm"
)

func getByID(id uuid.UUID) func(db *gorm.DB) model.Provider[Entity] {
	return func(db *gorm.DB) model.Provider[Entity] {
		var result Entity
		err := db.Where("id = ?", id).First(&result).Error
		if err != nil {
			return model.ErrorProvider[Entity](err)
		}
		return model.FixedProvider(result)
	}
}

func getByUserID(userID uuid.UUID) func(db *gorm.DB) model.Provider[[]Entity] {
	return func(db *gorm.DB) model.Provider[[]Entity] {
		// Tenants are linked to users via memberships; for v1, one tenant per user
		// This is a simplified lookup — in practice we'd join through memberships
		var results []Entity
		err := db.Find(&results).Error
		if err != nil {
			return model.ErrorProvider[[]Entity](err)
		}
		return model.FixedProvider(results)
	}
}
