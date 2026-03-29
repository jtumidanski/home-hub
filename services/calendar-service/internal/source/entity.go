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
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_sources_connection_external ON calendar_sources (connection_id, external_id)").Error; err != nil {
		return err
	}
	// One-time: clear all sync tokens to force a full re-sync.
	// This corrects stale all-day event end dates and removes orphaned events.
	return runOnce(db, "clear_sync_tokens_v1", func(tx *gorm.DB) error {
		return tx.Model(&Entity{}).Where("sync_token != ''").Update("sync_token", "").Error
	})
}

func runOnce(db *gorm.DB, key string, fn func(*gorm.DB) error) error {
	_ = db.Exec("CREATE TABLE IF NOT EXISTS calendar_migrations (key VARCHAR(255) PRIMARY KEY)").Error
	var count int64
	db.Table("calendar_migrations").Where("key = ?", key).Count(&count)
	if count > 0 {
		return nil
	}
	if err := fn(db); err != nil {
		return err
	}
	return db.Exec("INSERT INTO calendar_migrations (key) VALUES (?)", key).Error
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
