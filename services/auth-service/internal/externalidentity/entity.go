package externalidentity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Entity is the GORM entity for auth.external_identities.
type Entity struct {
	Id              uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserId          uuid.UUID `gorm:"type:uuid;not null"`
	Provider        string    `gorm:"type:text;not null;uniqueIndex:idx_provider_subject,priority:1"`
	ProviderSubject string    `gorm:"type:text;not null;uniqueIndex:idx_provider_subject,priority:2"`
	CreatedAt       time.Time `gorm:"not null"`
	UpdatedAt       time.Time `gorm:"not null"`
}

func (Entity) TableName() string {
	return "external_identities"
}

func Migration(db *gorm.DB) error {
	return db.AutoMigrate(&Entity{})
}

// Make converts an Entity to a Model.
func Make(e Entity) (Model, error) {
	return NewBuilder().
		SetId(e.Id).
		SetUserId(e.UserId).
		SetProvider(e.Provider).
		SetProviderSubject(e.ProviderSubject).
		SetCreatedAt(e.CreatedAt).
		SetUpdatedAt(e.UpdatedAt).
		Build()
}
