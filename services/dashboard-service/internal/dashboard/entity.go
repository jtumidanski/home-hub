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
	CreatedAt     time.Time      `gorm:"not null"`
	UpdatedAt     time.Time      `gorm:"not null"`
}

func (Entity) TableName() string { return "dashboards" }

func Migration(db *gorm.DB) error {
	if err := db.AutoMigrate(&Entity{}); err != nil {
		return err
	}
	return db.Exec(`CREATE INDEX IF NOT EXISTS idx_dashboards_household_partial
		ON dashboards (tenant_id, household_id) WHERE user_id IS NULL`).Error
}
