package oidcprovider

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Entity is the GORM entity for auth.oidc_providers.
type Entity struct {
	Id        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name      string    `gorm:"type:text;not null"`
	IssuerURL string    `gorm:"type:text;not null"`
	ClientID  string    `gorm:"type:text;not null"`
	Enabled   bool      `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (Entity) TableName() string {
	return "oidc_providers"
}

func Migration(db *gorm.DB) error {
	return db.AutoMigrate(&Entity{})
}

// Make converts an Entity to a Model.
func Make(e Entity) (Model, error) {
	return NewBuilder().
		SetId(e.Id).
		SetName(e.Name).
		SetIssuerURL(e.IssuerURL).
		SetClientID(e.ClientID).
		SetEnabled(e.Enabled).
		Build()
}
