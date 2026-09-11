package preference

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	database "github.com/jtumidanski/home-hub/shared/go/database"
)

// getByUser returns a preference for a user.
// Tenant filtering is automatic via GORM callbacks when db.WithContext(ctx) is used.
func getByUser(userID uuid.UUID) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ?", userID)
	})
}

func getByID(id uuid.UUID) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	})
}
