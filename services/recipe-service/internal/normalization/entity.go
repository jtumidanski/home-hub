package normalization

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entity struct {
	Id                    uuid.UUID  `gorm:"type:uuid;primaryKey"`
	TenantId              uuid.UUID  `gorm:"type:uuid;not null;index:idx_recipe_ingredient_tenant_household"`
	HouseholdId           uuid.UUID  `gorm:"type:uuid;not null;index:idx_recipe_ingredient_tenant_household"`
	RecipeId              uuid.UUID  `gorm:"type:uuid;not null;index:idx_recipe_ingredient_recipe"`
	RawName               string     `gorm:"type:varchar(255);not null"`
	RawQuantity           *string    `gorm:"type:varchar(100)"`
	RawUnit               *string    `gorm:"type:varchar(100)"`
	Position              int        `gorm:"type:int;not null"`
	CanonicalIngredientId *uuid.UUID `gorm:"type:uuid;index:idx_recipe_ingredient_canonical"`
	CanonicalUnit         *string    `gorm:"type:varchar(50)"`
	NormalizationStatus   string     `gorm:"type:varchar(30);not null;default:'unresolved'"`
	CreatedAt             time.Time  `gorm:"not null"`
	UpdatedAt             time.Time  `gorm:"not null"`
}

func (Entity) TableName() string { return "recipe_ingredients" }

func Migration(db *gorm.DB) error {
	return db.AutoMigrate(&Entity{})
}

func (m Model) ToEntity() Entity {
	var rawQty *string
	if m.rawQuantity != "" {
		rawQty = &m.rawQuantity
	}
	var rawUnit *string
	if m.rawUnit != "" {
		rawUnit = &m.rawUnit
	}
	var canonicalUnit *string
	if m.canonicalUnit != "" {
		canonicalUnit = &m.canonicalUnit
	}
	return Entity{
		Id: m.id, TenantId: m.tenantID, HouseholdId: m.householdID,
		RecipeId: m.recipeID, RawName: m.rawName,
		RawQuantity: rawQty, RawUnit: rawUnit, Position: m.position,
		CanonicalIngredientId: m.canonicalIngredientID, CanonicalUnit: canonicalUnit,
		NormalizationStatus: string(m.normalizationStatus),
		CreatedAt: m.createdAt, UpdatedAt: m.updatedAt,
	}
}

func Make(e Entity) (Model, error) {
	rawQuantity := ""
	if e.RawQuantity != nil {
		rawQuantity = *e.RawQuantity
	}
	rawUnit := ""
	if e.RawUnit != nil {
		rawUnit = *e.RawUnit
	}
	canonicalUnit := ""
	if e.CanonicalUnit != nil {
		canonicalUnit = *e.CanonicalUnit
	}
	return NewBuilder().
		SetId(e.Id).
		SetTenantID(e.TenantId).
		SetHouseholdID(e.HouseholdId).
		SetRecipeID(e.RecipeId).
		SetRawName(e.RawName).
		SetRawQuantity(rawQuantity).
		SetRawUnit(rawUnit).
		SetPosition(e.Position).
		SetCanonicalIngredientID(e.CanonicalIngredientId).
		SetCanonicalUnit(canonicalUnit).
		SetNormalizationStatus(Status(e.NormalizationStatus)).
		SetCreatedAt(e.CreatedAt).
		SetUpdatedAt(e.UpdatedAt).
		Build()
}
