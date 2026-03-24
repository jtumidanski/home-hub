package refreshtoken

import (
	"time"

	"github.com/jtumidanski/home-hub/shared/go/model"
	"gorm.io/gorm"
)

func getByHash(hash string) func(db *gorm.DB) model.Provider[Entity] {
	return func(db *gorm.DB) model.Provider[Entity] {
		var result Entity
		err := db.Where("token_hash = ? AND revoked = ? AND expires_at > ?", hash, false, time.Now().UTC()).
			First(&result).Error
		if err != nil {
			return model.ErrorProvider[Entity](err)
		}
		return model.FixedProvider(result)
	}
}
