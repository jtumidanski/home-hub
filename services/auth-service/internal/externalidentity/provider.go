package externalidentity

import (
	"gorm.io/gorm"

	"github.com/jtumidanski/home-hub/shared/go/database"
)

func getByProviderAndSubject(provider, subject string) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("provider = ? AND provider_subject = ?", provider, subject)
	})
}
