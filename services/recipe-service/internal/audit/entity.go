package audit

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entity struct {
	Id         uuid.UUID  `gorm:"type:uuid;primaryKey"`
	TenantId   uuid.UUID  `gorm:"type:uuid;not null;index:idx_audit_tenant_entity,priority:1;index:idx_audit_tenant_action,priority:1"`
	EntityType string     `gorm:"type:varchar(50);not null;index:idx_audit_tenant_entity,priority:2"`
	EntityId   uuid.UUID  `gorm:"type:uuid;not null;index:idx_audit_tenant_entity,priority:3"`
	Action     string     `gorm:"type:varchar(50);not null;index:idx_audit_tenant_action,priority:2"`
	ActorId    uuid.UUID  `gorm:"type:uuid;not null"`
	Metadata   *string    `gorm:"type:jsonb"`
	CreatedAt  time.Time  `gorm:"not null;index:idx_audit_created_at"`
}

func (Entity) TableName() string { return "recipe_audit_events" }

func Migration(db *gorm.DB) error {
	return db.AutoMigrate(&Entity{})
}
