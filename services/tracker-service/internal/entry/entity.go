package entry

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entity struct {
	Id             uuid.UUID       `gorm:"type:uuid;primaryKey"`
	TenantId       uuid.UUID       `gorm:"type:uuid;not null;index:idx_entry_tenant_user_date"`
	UserId         uuid.UUID       `gorm:"type:uuid;not null;index:idx_entry_tenant_user_date"`
	TrackingItemId uuid.UUID       `gorm:"type:uuid;not null;uniqueIndex:idx_entry_item_date"`
	Date           time.Time       `gorm:"type:date;not null;uniqueIndex:idx_entry_item_date;index:idx_entry_tenant_user_date"`
	Value          json.RawMessage `gorm:"type:jsonb"`
	Skipped        bool            `gorm:"not null;default:false"`
	Note           *string         `gorm:"type:varchar(500)"`
	CreatedAt      time.Time       `gorm:"not null"`
	UpdatedAt      time.Time       `gorm:"not null"`
}

func (Entity) TableName() string { return "tracking_entries" }

func Migration(db *gorm.DB) error {
	return db.AutoMigrate(&Entity{})
}

func (m Model) ToEntity() Entity {
	return Entity{
		Id:             m.id,
		TenantId:       m.tenantID,
		UserId:         m.userID,
		TrackingItemId: m.trackingItemID,
		Date:           m.date,
		Value:          m.value,
		Skipped:        m.skipped,
		Note:           m.note,
		CreatedAt:      m.createdAt,
		UpdatedAt:      m.updatedAt,
	}
}

func Make(e Entity) (Model, error) {
	return NewBuilder().
		SetId(e.Id).
		SetTenantID(e.TenantId).
		SetUserID(e.UserId).
		SetTrackingItemID(e.TrackingItemId).
		SetDate(e.Date).
		SetValue(e.Value).
		SetSkipped(e.Skipped).
		SetNote(e.Note).
		SetCreatedAt(e.CreatedAt).
		SetUpdatedAt(e.UpdatedAt).
		Build()
}
