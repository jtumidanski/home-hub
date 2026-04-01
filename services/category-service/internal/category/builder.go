package category

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNameRequired     = errors.New("category name is required")
	ErrNameTooLong      = errors.New("category name must not exceed 100 characters")
	ErrInvalidSortOrder = errors.New("sort order must be non-negative")
)

type Builder struct {
	id        uuid.UUID
	tenantID  uuid.UUID
	name      string
	sortOrder int
	createdAt time.Time
	updatedAt time.Time
}

func NewBuilder() *Builder { return &Builder{} }

func (b *Builder) SetId(id uuid.UUID) *Builder          { b.id = id; return b }
func (b *Builder) SetTenantID(id uuid.UUID) *Builder     { b.tenantID = id; return b }
func (b *Builder) SetName(name string) *Builder          { b.name = name; return b }
func (b *Builder) SetSortOrder(order int) *Builder       { b.sortOrder = order; return b }
func (b *Builder) SetCreatedAt(t time.Time) *Builder     { b.createdAt = t; return b }
func (b *Builder) SetUpdatedAt(t time.Time) *Builder     { b.updatedAt = t; return b }

func (b *Builder) Build() (Model, error) {
	if b.name == "" {
		return Model{}, ErrNameRequired
	}
	if len(b.name) > 100 {
		return Model{}, ErrNameTooLong
	}
	if b.sortOrder < 0 {
		return Model{}, ErrInvalidSortOrder
	}
	return Model{
		id:        b.id,
		tenantID:  b.tenantID,
		name:      b.name,
		sortOrder: b.sortOrder,
		createdAt: b.createdAt,
		updatedAt: b.updatedAt,
	}, nil
}
