package tenant

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	database "github.com/jtumidanski/home-hub/shared/go/database"
)

func getByID(id uuid.UUID) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	})
}
