package tenant

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entity struct {
	Id        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name      string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (Entity) TableName() string { return "tenants" }

func Migration(db *gorm.DB) error { return db.AutoMigrate(&Entity{}) }

func Make(e Entity) (Model, error) {
	return Model{
		id: e.Id, name: e.Name,
		createdAt: e.CreatedAt, updatedAt: e.UpdatedAt,
	}, nil
}
