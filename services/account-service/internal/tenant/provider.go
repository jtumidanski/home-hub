package tenant

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
