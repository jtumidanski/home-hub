package ingredient

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNameRequired      = errors.New("ingredient name is required")
	ErrInvalidUnitFamily = errors.New("unit family must be count, weight, volume, or empty")
)

type Builder struct {
	id           uuid.UUID
	tenantID     uuid.UUID
	name         string
	displayName  string
	unitFamily   string
	categoryID   *uuid.UUID
	categoryName string
	aliases      []Alias
	aliasCount   int
	usageCount   int
	createdAt    time.Time
	updatedAt    time.Time
}

func NewBuilder() *Builder { return &Builder{} }

func (b *Builder) SetId(id uuid.UUID) *Builder          { b.id = id; return b }
func (b *Builder) SetTenantID(id uuid.UUID) *Builder     { b.tenantID = id; return b }
func (b *Builder) SetName(name string) *Builder          { b.name = name; return b }
func (b *Builder) SetDisplayName(name string) *Builder   { b.displayName = name; return b }
func (b *Builder) SetUnitFamily(uf string) *Builder          { b.unitFamily = uf; return b }
func (b *Builder) SetCategoryID(id *uuid.UUID) *Builder      { b.categoryID = id; return b }
func (b *Builder) SetCategoryName(name string) *Builder      { b.categoryName = name; return b }
func (b *Builder) SetAliases(a []Alias) *Builder             { b.aliases = a; return b }
func (b *Builder) SetAliasCount(c int) *Builder          { b.aliasCount = c; return b }
func (b *Builder) SetUsageCount(c int) *Builder          { b.usageCount = c; return b }
func (b *Builder) SetCreatedAt(t time.Time) *Builder     { b.createdAt = t; return b }
func (b *Builder) SetUpdatedAt(t time.Time) *Builder     { b.updatedAt = t; return b }

func (b *Builder) Build() (Model, error) {
	if b.name == "" {
		return Model{}, ErrNameRequired
	}
	if !ValidUnitFamily(b.unitFamily) {
		return Model{}, ErrInvalidUnitFamily
	}
	return Model{
		id:           b.id,
		tenantID:     b.tenantID,
		name:         b.name,
		displayName:  b.displayName,
		unitFamily:   b.unitFamily,
		categoryID:   b.categoryID,
		categoryName: b.categoryName,
		aliases:      b.aliases,
		aliasCount:   b.aliasCount,
		usageCount:   b.usageCount,
		createdAt:    b.createdAt,
		updatedAt:    b.updatedAt,
	}, nil
}
