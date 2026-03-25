package reminder

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTitleRequired        = errors.New("reminder title is required")
	ErrScheduledForRequired = errors.New("reminder scheduledFor is required")
)

type Builder struct {
	id               uuid.UUID
	tenantID         uuid.UUID
	householdID      uuid.UUID
	title            string
	notes            string
	scheduledFor     time.Time
	lastDismissedAt  *time.Time
	lastSnoozedUntil *time.Time
	createdAt        time.Time
	updatedAt        time.Time
}

func NewBuilder() *Builder {
	return &Builder{}
}

func (b *Builder) SetId(id uuid.UUID) *Builder                   { b.id = id; return b }
func (b *Builder) SetTenantID(id uuid.UUID) *Builder              { b.tenantID = id; return b }
func (b *Builder) SetHouseholdID(id uuid.UUID) *Builder           { b.householdID = id; return b }
func (b *Builder) SetTitle(title string) *Builder                  { b.title = title; return b }
func (b *Builder) SetNotes(notes string) *Builder                  { b.notes = notes; return b }
func (b *Builder) SetScheduledFor(t time.Time) *Builder            { b.scheduledFor = t; return b }
func (b *Builder) SetLastDismissedAt(t *time.Time) *Builder        { b.lastDismissedAt = t; return b }
func (b *Builder) SetLastSnoozedUntil(t *time.Time) *Builder       { b.lastSnoozedUntil = t; return b }
func (b *Builder) SetCreatedAt(t time.Time) *Builder               { b.createdAt = t; return b }
func (b *Builder) SetUpdatedAt(t time.Time) *Builder               { b.updatedAt = t; return b }

func (b *Builder) Build() (Model, error) {
	if b.title == "" {
		return Model{}, ErrTitleRequired
	}
	if b.scheduledFor.IsZero() {
		return Model{}, ErrScheduledForRequired
	}
	return Model{
		id:               b.id,
		tenantID:         b.tenantID,
		householdID:      b.householdID,
		title:            b.title,
		notes:            b.notes,
		scheduledFor:     b.scheduledFor,
		lastDismissedAt:  b.lastDismissedAt,
		lastSnoozedUntil: b.lastSnoozedUntil,
		createdAt:        b.createdAt,
		updatedAt:        b.updatedAt,
	}, nil
}
