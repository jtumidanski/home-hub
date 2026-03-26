package source

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entity struct {
	Id           uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantId     uuid.UUID `gorm:"type:uuid;not null;index"`
	HouseholdId  uuid.UUID `gorm:"type:uuid;not null;index"`
	ConnectionId uuid.UUID `gorm:"type:uuid;not null;index:idx_sources_connection"`
	ExternalId   string    `gorm:"type:varchar(255);not null"`
	Name         string    `gorm:"type:varchar(255);not null"`
	Primary      bool      `gorm:"not null;default:false"`
	Visible      bool      `gorm:"not null;default:true"`
	Color        string    `gorm:"type:varchar(7)"`
	SyncToken    string    `gorm:"type:text"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
}

func (Entity) TableName() string { return "calendar_sources" }

func Migration(db *gorm.DB) error {
	if err := db.AutoMigrate(&Entity{}); err != nil {
		return err
	}
	return db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_sources_connection_external ON calendar_sources (connection_id, external_id)").Error
}

func (m Model) ToEntity() Entity {
	return Entity{
		Id:           m.id,
		TenantId:     m.tenantID,
		HouseholdId:  m.householdID,
		ConnectionId: m.connectionID,
		ExternalId:   m.externalID,
		Name:         m.name,
		Primary:      m.primary,
		Visible:      m.visible,
		Color:        m.color,
		SyncToken:    m.syncToken,
		CreatedAt:    m.createdAt,
		UpdatedAt:    m.updatedAt,
	}
}

func Make(e Entity) (Model, error) {
	return NewBuilder().
		SetId(e.Id).
		SetTenantID(e.TenantId).
		SetHouseholdID(e.HouseholdId).
		SetConnectionID(e.ConnectionId).
		SetExternalID(e.ExternalId).
		SetName(e.Name).
		SetPrimary(e.Primary).
		SetVisible(e.Visible).
		SetColor(e.Color).
		SetSyncToken(e.SyncToken).
		SetCreatedAt(e.CreatedAt).
		SetUpdatedAt(e.UpdatedAt).
		Build()
}
