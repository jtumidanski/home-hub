package category

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entity struct {
	Id        uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantId  uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_ingredient_category_tenant_name"`
	Name      string    `gorm:"type:varchar(100);not null;uniqueIndex:idx_ingredient_category_tenant_name"`
	SortOrder int       `gorm:"not null;default:0"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (Entity) TableName() string { return "ingredient_categories" }

func Migration(db *gorm.DB) error {
	return db.AutoMigrate(&Entity{})
}

func (m Model) ToEntity() Entity {
	return Entity{
		Id:        m.id,
		TenantId:  m.tenantID,
		Name:      m.name,
		SortOrder: m.sortOrder,
		CreatedAt: m.createdAt,
		UpdatedAt: m.updatedAt,
	}
}

func Make(e Entity) (Model, error) {
	return NewBuilder().
		SetId(e.Id).
		SetTenantID(e.TenantId).
		SetName(e.Name).
		SetSortOrder(e.SortOrder).
		SetCreatedAt(e.CreatedAt).
		SetUpdatedAt(e.UpdatedAt).
		Build()
}
