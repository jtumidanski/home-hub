package theme

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNameRequired     = errors.New("theme name is required")
	ErrNameTooLong      = errors.New("theme name must not exceed 50 characters")
	ErrInvalidSortOrder = errors.New("sort order must be non-negative")
)

type Builder struct {
	id        uuid.UUID
	tenantID  uuid.UUID
	userID    uuid.UUID
	name      string
	sortOrder int
	createdAt time.Time
	updatedAt time.Time
	deletedAt *time.Time
}

func NewBuilder() *Builder { return &Builder{} }

func (b *Builder) SetId(id uuid.UUID) *Builder       { b.id = id; return b }
func (b *Builder) SetTenantID(id uuid.UUID) *Builder { b.tenantID = id; return b }
func (b *Builder) SetUserID(id uuid.UUID) *Builder   { b.userID = id; return b }
func (b *Builder) SetName(n string) *Builder          { b.name = n; return b }
func (b *Builder) SetSortOrder(o int) *Builder        { b.sortOrder = o; return b }
func (b *Builder) SetCreatedAt(t time.Time) *Builder  { b.createdAt = t; return b }
func (b *Builder) SetUpdatedAt(t time.Time) *Builder  { b.updatedAt = t; return b }
func (b *Builder) SetDeletedAt(t *time.Time) *Builder { b.deletedAt = t; return b }

func (b *Builder) Build() (Model, error) {
	if b.name == "" {
		return Model{}, ErrNameRequired
	}
	if len(b.name) > 50 {
		return Model{}, ErrNameTooLong
	}
	if b.sortOrder < 0 {
		return Model{}, ErrInvalidSortOrder
	}
	return Model{
		id:        b.id,
		tenantID:  b.tenantID,
		userID:    b.userID,
		name:      b.name,
		sortOrder: b.sortOrder,
		createdAt: b.createdAt,
		updatedAt: b.updatedAt,
		deletedAt: b.deletedAt,
	}, nil
}
