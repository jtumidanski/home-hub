package normalization

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrRawNameRequired = errors.New("raw ingredient name is required")
)

type Builder struct {
	id                    uuid.UUID
	tenantID              uuid.UUID
	householdID           uuid.UUID
	recipeID              uuid.UUID
	rawName               string
	rawQuantity           string
	rawUnit               string
	position              int
	canonicalIngredientID *uuid.UUID
	canonicalUnit         string
	normalizationStatus   Status
	createdAt             time.Time
	updatedAt             time.Time
}

func NewBuilder() *Builder { return &Builder{} }

func (b *Builder) SetId(id uuid.UUID) *Builder                     { b.id = id; return b }
func (b *Builder) SetTenantID(id uuid.UUID) *Builder               { b.tenantID = id; return b }
func (b *Builder) SetHouseholdID(id uuid.UUID) *Builder            { b.householdID = id; return b }
func (b *Builder) SetRecipeID(id uuid.UUID) *Builder               { b.recipeID = id; return b }
func (b *Builder) SetRawName(name string) *Builder                 { b.rawName = name; return b }
func (b *Builder) SetRawQuantity(qty string) *Builder              { b.rawQuantity = qty; return b }
func (b *Builder) SetRawUnit(unit string) *Builder                 { b.rawUnit = unit; return b }
func (b *Builder) SetPosition(pos int) *Builder                    { b.position = pos; return b }
func (b *Builder) SetCanonicalIngredientID(id *uuid.UUID) *Builder { b.canonicalIngredientID = id; return b }
func (b *Builder) SetCanonicalUnit(unit string) *Builder           { b.canonicalUnit = unit; return b }
func (b *Builder) SetNormalizationStatus(s Status) *Builder        { b.normalizationStatus = s; return b }
func (b *Builder) SetCreatedAt(t time.Time) *Builder               { b.createdAt = t; return b }
func (b *Builder) SetUpdatedAt(t time.Time) *Builder               { b.updatedAt = t; return b }

func (b *Builder) Build() (Model, error) {
	if b.rawName == "" {
		return Model{}, ErrRawNameRequired
	}
	return Model{
		id:                    b.id,
		tenantID:              b.tenantID,
		householdID:           b.householdID,
		recipeID:              b.recipeID,
		rawName:               b.rawName,
		rawQuantity:           b.rawQuantity,
		rawUnit:               b.rawUnit,
		position:              b.position,
		canonicalIngredientID: b.canonicalIngredientID,
		canonicalUnit:         b.canonicalUnit,
		normalizationStatus:   b.normalizationStatus,
		createdAt:             b.createdAt,
		updatedAt:             b.updatedAt,
	}, nil
}
