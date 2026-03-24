package household

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

func getByTenantID(tenantID uuid.UUID) func(db *gorm.DB) model.Provider[[]Entity] {
	return func(db *gorm.DB) model.Provider[[]Entity] {
		var results []Entity
		err := db.Where("tenant_id = ?", tenantID).Find(&results).Error
		if err != nil {
			return model.ErrorProvider[[]Entity](err)
		}
		return model.FixedProvider(results)
	}
}
