package source

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrExternalIDRequired = errors.New("external ID is required")
	ErrNameRequired       = errors.New("source name is required")
)

type Builder struct {
	id           uuid.UUID
	tenantID     uuid.UUID
	householdID  uuid.UUID
	connectionID uuid.UUID
	externalID   string
	name         string
	primary      bool
	visible      bool
	color        string
	syncToken    string
	createdAt    time.Time
	updatedAt    time.Time
}

func NewBuilder() *Builder { return &Builder{} }

func (b *Builder) SetId(id uuid.UUID) *Builder           { b.id = id; return b }
func (b *Builder) SetTenantID(id uuid.UUID) *Builder      { b.tenantID = id; return b }
func (b *Builder) SetHouseholdID(id uuid.UUID) *Builder   { b.householdID = id; return b }
func (b *Builder) SetConnectionID(id uuid.UUID) *Builder  { b.connectionID = id; return b }
func (b *Builder) SetExternalID(eid string) *Builder      { b.externalID = eid; return b }
func (b *Builder) SetName(n string) *Builder              { b.name = n; return b }
func (b *Builder) SetPrimary(p bool) *Builder             { b.primary = p; return b }
func (b *Builder) SetVisible(v bool) *Builder             { b.visible = v; return b }
func (b *Builder) SetColor(c string) *Builder             { b.color = c; return b }
func (b *Builder) SetSyncToken(t string) *Builder         { b.syncToken = t; return b }
func (b *Builder) SetCreatedAt(t time.Time) *Builder      { b.createdAt = t; return b }
func (b *Builder) SetUpdatedAt(t time.Time) *Builder      { b.updatedAt = t; return b }

func (b *Builder) Build() (Model, error) {
	if b.externalID == "" {
		return Model{}, ErrExternalIDRequired
	}
	if b.name == "" {
		return Model{}, ErrNameRequired
	}
	return Model{
		id:           b.id,
		tenantID:     b.tenantID,
		householdID:  b.householdID,
		connectionID: b.connectionID,
		externalID:   b.externalID,
		name:         b.name,
		primary:      b.primary,
		visible:      b.visible,
		color:        b.color,
		syncToken:    b.syncToken,
		createdAt:    b.createdAt,
		updatedAt:    b.updatedAt,
	}, nil
}
