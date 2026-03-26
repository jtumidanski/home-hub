package oauthstate

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entity struct {
	Id          uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantId    uuid.UUID `gorm:"type:uuid;not null"`
	HouseholdId uuid.UUID `gorm:"type:uuid;not null"`
	UserId      uuid.UUID `gorm:"type:uuid;not null"`
	RedirectUri string    `gorm:"type:varchar(500);not null"`
	ExpiresAt   time.Time `gorm:"not null"`
	CreatedAt   time.Time `gorm:"not null"`
}

func (Entity) TableName() string { return "calendar_oauth_states" }

func Migration(db *gorm.DB) error { return db.AutoMigrate(&Entity{}) }

func (m Model) ToEntity() Entity {
	return Entity{
		Id:          m.id,
		TenantId:    m.tenantID,
		HouseholdId: m.householdID,
		UserId:      m.userID,
		RedirectUri: m.redirectURI,
		ExpiresAt:   m.expiresAt,
		CreatedAt:   m.createdAt,
	}
}

func Make(e Entity) (Model, error) {
	return NewBuilder().
		SetId(e.Id).
		SetTenantID(e.TenantId).
		SetHouseholdID(e.HouseholdId).
		SetUserID(e.UserId).
		SetRedirectURI(e.RedirectUri).
		SetExpiresAt(e.ExpiresAt).
		SetCreatedAt(e.CreatedAt).
		Build()
}
