package trackingitem

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entity struct {
	Id          uuid.UUID       `gorm:"type:uuid;primaryKey"`
	TenantId    uuid.UUID       `gorm:"type:uuid;not null;uniqueIndex:idx_tracking_item_tenant_user_name,where:deleted_at IS NULL"`
	UserId      uuid.UUID       `gorm:"type:uuid;not null;uniqueIndex:idx_tracking_item_tenant_user_name,where:deleted_at IS NULL;index:idx_tracking_item_user"`
	Name        string          `gorm:"type:varchar(100);not null;uniqueIndex:idx_tracking_item_tenant_user_name,where:deleted_at IS NULL"`
	ScaleType   string          `gorm:"type:varchar(20);not null"`
	ScaleConfig json.RawMessage `gorm:"type:jsonb"`
	Color       string          `gorm:"type:varchar(20);not null"`
	SortOrder   int             `gorm:"not null;default:0"`
	CreatedAt   time.Time       `gorm:"not null"`
	UpdatedAt   time.Time       `gorm:"not null"`
	DeletedAt   *time.Time      `gorm:"index"`
}

func (Entity) TableName() string { return "tracker.tracking_items" }

func Migration(db *gorm.DB) error {
	return db.AutoMigrate(&Entity{})
}

func (m Model) ToEntity() Entity {
	return Entity{
		Id:          m.id,
		TenantId:    m.tenantID,
		UserId:      m.userID,
		Name:        m.name,
		ScaleType:   m.scaleType,
		ScaleConfig: m.scaleConfig,
		Color:       m.color,
		SortOrder:   m.sortOrder,
		CreatedAt:   m.createdAt,
		UpdatedAt:   m.updatedAt,
		DeletedAt:   m.deletedAt,
	}
}

func Make(e Entity) (Model, error) {
	return NewBuilder().
		SetId(e.Id).
		SetTenantID(e.TenantId).
		SetUserID(e.UserId).
		SetName(e.Name).
		SetScaleType(e.ScaleType).
		SetScaleConfig(e.ScaleConfig).
		SetColor(e.Color).
		SetSortOrder(e.SortOrder).
		SetCreatedAt(e.CreatedAt).
		SetUpdatedAt(e.UpdatedAt).
		SetDeletedAt(e.DeletedAt).
		Build()
}
