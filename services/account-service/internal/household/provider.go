package household

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

// getAll returns all households for the current tenant.
// Tenant filtering is automatic via GORM callbacks when db.WithContext(ctx) is used.
func getAll() database.EntityProvider[[]Entity] {
	return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB {
		return db
	})
}
