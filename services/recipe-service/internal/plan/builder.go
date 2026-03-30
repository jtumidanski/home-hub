package plan

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrStartsOnRequired = errors.New("starts_on is required")
)

type Builder struct {
	id          uuid.UUID
	tenantID    uuid.UUID
	householdID uuid.UUID
	startsOn    time.Time
	name        string
	locked      bool
	createdBy   uuid.UUID
	createdAt   time.Time
	updatedAt   time.Time
}

func NewBuilder() *Builder { return &Builder{} }

func (b *Builder) SetId(id uuid.UUID) *Builder          { b.id = id; return b }
func (b *Builder) SetTenantID(id uuid.UUID) *Builder     { b.tenantID = id; return b }
func (b *Builder) SetHouseholdID(id uuid.UUID) *Builder  { b.householdID = id; return b }
func (b *Builder) SetStartsOn(t time.Time) *Builder      { b.startsOn = t; return b }
func (b *Builder) SetName(name string) *Builder           { b.name = name; return b }
func (b *Builder) SetLocked(locked bool) *Builder         { b.locked = locked; return b }
func (b *Builder) SetCreatedBy(id uuid.UUID) *Builder     { b.createdBy = id; return b }
func (b *Builder) SetCreatedAt(t time.Time) *Builder      { b.createdAt = t; return b }
func (b *Builder) SetUpdatedAt(t time.Time) *Builder      { b.updatedAt = t; return b }

func (b *Builder) Build() (Model, error) {
	if b.startsOn.IsZero() {
		return Model{}, ErrStartsOnRequired
	}
	if b.name == "" {
		b.name = fmt.Sprintf("Week of %s", b.startsOn.Format("January 2, 2006"))
	}
	return Model{
		id:          b.id,
		tenantID:    b.tenantID,
		householdID: b.householdID,
		startsOn:    b.startsOn,
		name:        b.name,
		locked:      b.locked,
		createdBy:   b.createdBy,
		createdAt:   b.createdAt,
		updatedAt:   b.updatedAt,
	}, nil
}
