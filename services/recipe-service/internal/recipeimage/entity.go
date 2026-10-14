package recipeimage

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entity struct {
	Id          uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantId    uuid.UUID `gorm:"type:uuid;not null;index:idx_recipe_image_tenant_household"`
	HouseholdId uuid.UUID `gorm:"type:uuid;not null;index:idx_recipe_image_tenant_household"`
	RecipeId    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_recipe_image_recipe"`
	FullKey     string    `gorm:"type:varchar(512);not null"`
	ThumbKey    string    `gorm:"type:varchar(512);not null"`
	ContentType string    `gorm:"type:varchar(50);not null"`
	Width       int       `gorm:"type:int;not null"`
	Height      int       `gorm:"type:int;not null"`
	FullBytes   int       `gorm:"type:int;not null"`
	ThumbBytes  int       `gorm:"type:int;not null"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}

func (Entity) TableName() string { return "recipe_images" }

// ObjectDeletionEntity is the transactional outbox for object deletions. Rows
// are written in the same transaction that removes an image row and drained
// after commit by ObjectReaper.
type ObjectDeletionEntity struct {
	Id          uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantId    uuid.UUID `gorm:"type:uuid;not null"`
	HouseholdId uuid.UUID `gorm:"type:uuid;not null"`
	ObjectKey   string    `gorm:"type:varchar(512);not null"`
	Attempts    int       `gorm:"type:int;not null;default:0"`
	LastError   *string   `gorm:"type:text"`
	CreatedAt   time.Time `gorm:"not null;index:idx_recipe_image_deletion_created"`
}

func (ObjectDeletionEntity) TableName() string { return "recipe_image_object_deletions" }

func Migration(db *gorm.DB) error {
	return db.AutoMigrate(&Entity{}, &ObjectDeletionEntity{})
}

func (m Model) ToEntity() Entity {
	return Entity{
		Id: m.id, TenantId: m.tenantID, HouseholdId: m.householdID, RecipeId: m.recipeID,
		FullKey: m.fullKey, ThumbKey: m.thumbKey, ContentType: m.contentType,
		Width: m.width, Height: m.height, FullBytes: m.fullBytes, ThumbBytes: m.thumbBytes,
		CreatedAt: m.createdAt, UpdatedAt: m.updatedAt,
	}
}

func Make(e Entity) (Model, error) {
	return NewBuilder().
		SetId(e.Id).
		SetTenantID(e.TenantId).
		SetHouseholdID(e.HouseholdId).
		SetRecipeID(e.RecipeId).
		SetFullKey(e.FullKey).
		SetThumbKey(e.ThumbKey).
		SetContentType(e.ContentType).
		SetWidth(e.Width).
		SetHeight(e.Height).
		SetFullBytes(e.FullBytes).
		SetThumbBytes(e.ThumbBytes).
		SetCreatedAt(e.CreatedAt).
		SetUpdatedAt(e.UpdatedAt).
		Build()
}
