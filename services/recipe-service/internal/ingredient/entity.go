package ingredient

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entity struct {
	Id          uuid.UUID     `gorm:"type:uuid;primaryKey"`
	TenantId    uuid.UUID     `gorm:"type:uuid;not null;uniqueIndex:idx_canonical_ingredient_tenant_name"`
	Name        string        `gorm:"type:varchar(255);not null;uniqueIndex:idx_canonical_ingredient_tenant_name"`
	DisplayName *string       `gorm:"type:varchar(255)"`
	UnitFamily  *string       `gorm:"type:varchar(20)"`
	CategoryId  *uuid.UUID    `gorm:"type:uuid;index:idx_canonical_ingredient_category"`
	Aliases     []AliasEntity `gorm:"foreignKey:CanonicalIngredientId;constraint:OnDelete:CASCADE"`
	CreatedAt   time.Time     `gorm:"not null"`
	UpdatedAt   time.Time     `gorm:"not null"`
}

func (Entity) TableName() string { return "canonical_ingredients" }

type AliasEntity struct {
	Id                    uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantId              uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_alias_tenant_name"`
	CanonicalIngredientId uuid.UUID `gorm:"type:uuid;not null;index:idx_alias_canonical_ingredient"`
	Name                  string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_alias_tenant_name"`
	CreatedAt             time.Time `gorm:"not null"`
}

func (AliasEntity) TableName() string { return "canonical_ingredient_aliases" }

func Migration(db *gorm.DB) error {
	if err := db.AutoMigrate(&Entity{}, &AliasEntity{}); err != nil {
		return err
	}
	// Drop the FK constraint to ingredient_categories since categories are now managed by category-service.
	// category_id remains as an opaque UUID reference.
	if db.Migrator().HasConstraint(&Entity{}, "fk_canonical_ingredients_category") {
		_ = db.Exec("ALTER TABLE canonical_ingredients DROP CONSTRAINT fk_canonical_ingredients_category").Error
	}
	return nil
}

func Make(e Entity) (Model, error) {
	displayName := ""
	if e.DisplayName != nil {
		displayName = *e.DisplayName
	}
	unitFamily := ""
	if e.UnitFamily != nil {
		unitFamily = *e.UnitFamily
	}
	aliases := make([]Alias, len(e.Aliases))
	for i, a := range e.Aliases {
		aliases[i] = Alias{id: a.Id, name: a.Name}
	}
	return NewBuilder().
		SetId(e.Id).
		SetTenantID(e.TenantId).
		SetName(e.Name).
		SetDisplayName(displayName).
		SetUnitFamily(unitFamily).
		SetCategoryID(e.CategoryId).
		SetAliases(aliases).
		SetAliasCount(len(e.Aliases)).
		SetCreatedAt(e.CreatedAt).
		SetUpdatedAt(e.UpdatedAt).
		Build()
}
