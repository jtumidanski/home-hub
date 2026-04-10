package theme

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entity struct {
	Id        uuid.UUID  `gorm:"type:uuid;primaryKey"`
	TenantId  uuid.UUID  `gorm:"type:uuid;not null;index:idx_workout_theme_tenant_user_deleted,priority:1"`
	UserId    uuid.UUID  `gorm:"type:uuid;not null;index:idx_workout_theme_tenant_user_deleted,priority:2"`
	Name      string     `gorm:"type:varchar(50);not null"`
	SortOrder int        `gorm:"not null;default:0"`
	CreatedAt time.Time  `gorm:"not null"`
	UpdatedAt time.Time  `gorm:"not null"`
	DeletedAt *time.Time `gorm:"index:idx_workout_theme_tenant_user_deleted,priority:3"`
}

func (Entity) TableName() string { return "themes" }

// Migration runs AutoMigrate plus the partial-unique-index DDL that GORM cannot
// generate natively (the `WHERE deleted_at IS NULL` clause).
func Migration(db *gorm.DB) error {
	if err := db.AutoMigrate(&Entity{}); err != nil {
		return err
	}
	return db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_workout_themes_tenant_user_name_active
		ON themes (tenant_id, user_id, name) WHERE deleted_at IS NULL`).Error
}

func (m Model) ToEntity() Entity {
	return Entity{
		Id:        m.id,
		TenantId:  m.tenantID,
		UserId:    m.userID,
		Name:      m.name,
		SortOrder: m.sortOrder,
		CreatedAt: m.createdAt,
		UpdatedAt: m.updatedAt,
		DeletedAt: m.deletedAt,
	}
}

func Make(e Entity) (Model, error) {
	return NewBuilder().
		SetId(e.Id).
		SetTenantID(e.TenantId).
		SetUserID(e.UserId).
		SetName(e.Name).
		SetSortOrder(e.SortOrder).
		SetCreatedAt(e.CreatedAt).
		SetUpdatedAt(e.UpdatedAt).
		SetDeletedAt(e.DeletedAt).
		Build()
}
