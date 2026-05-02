package dashboard

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Entity struct {
	Id            uuid.UUID      `gorm:"type:uuid;primaryKey"`
	TenantId      uuid.UUID      `gorm:"type:uuid;not null;index:idx_dashboards_scope"`
	HouseholdId   uuid.UUID      `gorm:"type:uuid;not null;index:idx_dashboards_scope"`
	UserId        *uuid.UUID     `gorm:"type:uuid;index:idx_dashboards_scope"`
	Name          string         `gorm:"type:varchar(80);not null"`
	SortOrder     int            `gorm:"not null;default:0"`
	Layout        datatypes.JSON `gorm:"type:jsonb;not null"`
	SchemaVersion int            `gorm:"not null;default:1"`
	SeedKey       *string        `gorm:"column:seed_key;type:varchar(40)"`
	CreatedAt     time.Time      `gorm:"not null"`
	UpdatedAt     time.Time      `gorm:"not null"`
}

func (Entity) TableName() string { return "dashboards" }

func Migration(db *gorm.DB) error {
	if err := db.AutoMigrate(&Entity{}); err != nil {
		return err
	}
	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_dashboards_household_partial
		ON dashboards (tenant_id, household_id) WHERE user_id IS NULL`).Error; err != nil {
		return err
	}
	if err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_dashboards_seed_key
		ON dashboards (tenant_id, household_id, seed_key) WHERE seed_key IS NOT NULL`).Error; err != nil {
		return err
	}
	// Brownfield backfill — claim existing seeded "Home" rows so the
	// updated client's idempotent home-seed call is a no-op for them.
	return db.Exec(`UPDATE dashboards
		SET seed_key = 'home'
		WHERE seed_key IS NULL
		  AND user_id IS NULL
		  AND sort_order = 0
		  AND name = 'Home'`).Error
}
