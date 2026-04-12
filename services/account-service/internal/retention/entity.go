package retention

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entity struct {
	Id            uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantId      uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_retention_overrides_unique;index:idx_retention_overrides_scope"`
	ScopeKind     string    `gorm:"type:text;not null;uniqueIndex:idx_retention_overrides_unique;index:idx_retention_overrides_scope"`
	ScopeId       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_retention_overrides_unique;index:idx_retention_overrides_scope"`
	Category      string    `gorm:"type:text;not null;uniqueIndex:idx_retention_overrides_unique"`
	RetentionDays int       `gorm:"not null"`
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`
}

func (Entity) TableName() string { return "retention_policy_overrides" }

func Migration(db *gorm.DB) error { return db.AutoMigrate(&Entity{}) }
