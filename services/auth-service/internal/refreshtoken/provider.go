package refreshtoken

import (
	"time"

	"github.com/jtumidanski/home-hub/shared/go/database"
	"gorm.io/gorm"
)

func getByHash(hash string) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("token_hash = ? AND revoked = ? AND expires_at > ?", hash, false, time.Now().UTC())
	})
}
