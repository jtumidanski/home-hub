package dashboard

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

var (
	ErrNameRequired      = errors.New("name is required")
	ErrNameTooLong       = errors.New("name exceeds 80 chars")
	ErrTenantRequired    = errors.New("tenant id is required")
	ErrHouseholdRequired = errors.New("household id is required")
)

type Builder struct {
	id            uuid.UUID
	tenantID      uuid.UUID
	householdID   uuid.UUID
	userID        *uuid.UUID
	name          string
	sortOrder     int
	layout        datatypes.JSON
	schemaVersion int
	createdAt     time.Time
	updatedAt     time.Time
}

func NewBuilder() *Builder { return &Builder{schemaVersion: 1} }

func BuilderFromModel(m Model) *Builder {
	return &Builder{
		id:            m.id,
		tenantID:      m.tenantID,
		householdID:   m.householdID,
		userID:        m.userID,
		name:          m.name,
		sortOrder:     m.sortOrder,
		layout:        m.layout,
		schemaVersion: m.schemaVersion,
		createdAt:     m.createdAt,
		updatedAt:     m.updatedAt,
	}
}

func (b *Builder) SetId(id uuid.UUID) *Builder             { b.id = id; return b }
func (b *Builder) SetTenantID(id uuid.UUID) *Builder       { b.tenantID = id; return b }
func (b *Builder) SetHouseholdID(id uuid.UUID) *Builder    { b.householdID = id; return b }
func (b *Builder) SetUserID(id *uuid.UUID) *Builder        { b.userID = id; return b }
func (b *Builder) SetName(n string) *Builder               { b.name = n; return b }
func (b *Builder) SetSortOrder(s int) *Builder             { b.sortOrder = s; return b }
func (b *Builder) SetLayout(l datatypes.JSON) *Builder     { b.layout = l; return b }
func (b *Builder) SetSchemaVersion(v int) *Builder         { b.schemaVersion = v; return b }
func (b *Builder) SetCreatedAt(t time.Time) *Builder       { b.createdAt = t; return b }
func (b *Builder) SetUpdatedAt(t time.Time) *Builder       { b.updatedAt = t; return b }

func (b *Builder) Build() (Model, error) {
	if b.tenantID == uuid.Nil {
		return Model{}, ErrTenantRequired
	}
	if b.householdID == uuid.Nil {
		return Model{}, ErrHouseholdRequired
	}
	trimmed := strings.TrimSpace(b.name)
	if trimmed == "" {
		return Model{}, ErrNameRequired
	}
	if len(trimmed) > 80 {
		return Model{}, ErrNameTooLong
	}
	return Model{
		id:            b.id,
		tenantID:      b.tenantID,
		householdID:   b.householdID,
		userID:        b.userID,
		name:          trimmed,
		sortOrder:     b.sortOrder,
		layout:        b.layout,
		schemaVersion: b.schemaVersion,
		createdAt:     b.createdAt,
		updatedAt:     b.updatedAt,
	}, nil
}
