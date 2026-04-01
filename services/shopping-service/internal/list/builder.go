package list

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNameRequired = errors.New("shopping list name is required")
	ErrNameTooLong  = errors.New("shopping list name must not exceed 255 characters")
)

type Builder struct {
	id          uuid.UUID
	tenantID    uuid.UUID
	householdID uuid.UUID
	name        string
	status      string
	archivedAt  *time.Time
	createdBy   uuid.UUID
	createdAt   time.Time
	updatedAt   time.Time
}

func NewBuilder() *Builder { return &Builder{status: "active"} }

func (b *Builder) SetId(id uuid.UUID) *Builder          { b.id = id; return b }
func (b *Builder) SetTenantID(id uuid.UUID) *Builder     { b.tenantID = id; return b }
func (b *Builder) SetHouseholdID(id uuid.UUID) *Builder  { b.householdID = id; return b }
func (b *Builder) SetName(name string) *Builder          { b.name = name; return b }
func (b *Builder) SetStatus(status string) *Builder      { b.status = status; return b }
func (b *Builder) SetArchivedAt(t *time.Time) *Builder   { b.archivedAt = t; return b }
func (b *Builder) SetCreatedBy(id uuid.UUID) *Builder    { b.createdBy = id; return b }
func (b *Builder) SetCreatedAt(t time.Time) *Builder     { b.createdAt = t; return b }
func (b *Builder) SetUpdatedAt(t time.Time) *Builder     { b.updatedAt = t; return b }

func (b *Builder) Build() (Model, error) {
	if b.name == "" {
		return Model{}, ErrNameRequired
	}
	if len(b.name) > 255 {
		return Model{}, ErrNameTooLong
	}
	return Model{
		id:          b.id,
		tenantID:    b.tenantID,
		householdID: b.householdID,
		name:        b.name,
		status:      b.status,
		archivedAt:  b.archivedAt,
		createdBy:   b.createdBy,
		createdAt:   b.createdAt,
		updatedAt:   b.updatedAt,
	}, nil
}
