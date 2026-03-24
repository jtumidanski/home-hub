package refreshtoken

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Entity is the GORM entity for auth.refresh_tokens.
type Entity struct {
	Id        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserId    uuid.UUID `gorm:"type:uuid;not null;index"`
	TokenHash string    `gorm:"type:text;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	Revoked   bool      `gorm:"not null;default:false"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (Entity) TableName() string {
	return "refresh_tokens"
}

func Migration(db *gorm.DB) error {
	return db.AutoMigrate(&Entity{})
}
