package preference

import (
	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/model"
	"gorm.io/gorm"
)

func getByTenantAndUser(tenantID, userID uuid.UUID) func(db *gorm.DB) model.Provider[Entity] {
	return func(db *gorm.DB) model.Provider[Entity] {
		var result Entity
		err := db.Where("tenant_id = ? AND user_id = ?", tenantID, userID).First(&result).Error
		if err != nil {
			return model.ErrorProvider[Entity](err)
		}
		return model.FixedProvider(result)
	}
}

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
