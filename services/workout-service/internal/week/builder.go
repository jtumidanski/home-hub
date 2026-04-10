package week

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrInvalidRestDay = errors.New("rest day flags must be integers in [0,6]")

type Builder struct {
	id            uuid.UUID
	tenantID      uuid.UUID
	userID        uuid.UUID
	weekStartDate time.Time
	restDayFlags  []int
	createdAt     time.Time
	updatedAt     time.Time
}

func NewBuilder() *Builder { return &Builder{restDayFlags: []int{}} }

func (b *Builder) SetId(id uuid.UUID) *Builder            { b.id = id; return b }
func (b *Builder) SetTenantID(id uuid.UUID) *Builder      { b.tenantID = id; return b }
func (b *Builder) SetUserID(id uuid.UUID) *Builder        { b.userID = id; return b }
func (b *Builder) SetWeekStartDate(t time.Time) *Builder  { b.weekStartDate = t; return b }
func (b *Builder) SetRestDayFlags(f []int) *Builder       { if f == nil { f = []int{} }; b.restDayFlags = f; return b }
func (b *Builder) SetCreatedAt(t time.Time) *Builder      { b.createdAt = t; return b }
func (b *Builder) SetUpdatedAt(t time.Time) *Builder      { b.updatedAt = t; return b }

func ValidateRestDayFlags(flags []int) error {
	for _, d := range flags {
		if d < 0 || d > 6 {
			return ErrInvalidRestDay
		}
	}
	return nil
}

func (b *Builder) Build() (Model, error) {
	if err := ValidateRestDayFlags(b.restDayFlags); err != nil {
		return Model{}, err
	}
	return Model{
		id:            b.id,
		tenantID:      b.tenantID,
		userID:        b.userID,
		weekStartDate: b.weekStartDate,
		restDayFlags:  b.restDayFlags,
		createdAt:     b.createdAt,
		updatedAt:     b.updatedAt,
	}, nil
}

// NormalizeToMonday returns the Monday of the ISO week containing `t`. The
// week-start contract treats Monday as day-of-week 0; this helper is the only
// place that does the wall-clock-to-Monday adjustment so every endpoint sees
// the same canonical date.
func NormalizeToMonday(t time.Time) time.Time {
	t = t.UTC().Truncate(24 * time.Hour)
	wd := int(t.Weekday()) // Sunday=0..Saturday=6 in Go
	// Convert to ISO Mon=0..Sun=6
	iso := (wd + 6) % 7
	return t.AddDate(0, 0, -iso)
}
