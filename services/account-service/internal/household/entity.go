package household

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entity struct {
	Id        uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantId  uuid.UUID `gorm:"type:uuid;not null;index"`
	Name      string    `gorm:"type:text;not null"`
	Timezone  string    `gorm:"type:text;not null"`
	Units     string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (Entity) TableName() string { return "households" }

func Migration(db *gorm.DB) error { return db.AutoMigrate(&Entity{}) }

func Make(e Entity) (Model, error) {
	return NewBuilder().
		SetId(e.Id).
		SetTenantID(e.TenantId).
		SetName(e.Name).
		SetTimezone(e.Timezone).
		SetUnits(e.Units).
		SetCreatedAt(e.CreatedAt).
		SetUpdatedAt(e.UpdatedAt).
		Build()
}
