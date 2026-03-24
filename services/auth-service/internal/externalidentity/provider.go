package externalidentity

import (
	"github.com/jtumidanski/home-hub/shared/go/database"
	"gorm.io/gorm"
)

func getByProviderAndSubject(provider, subject string) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("provider = ? AND provider_subject = ?", provider, subject)
	})
}
