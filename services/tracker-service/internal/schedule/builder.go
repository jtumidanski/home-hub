package schedule

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTrackingItemRequired = errors.New("tracking item ID is required")
	ErrEffectiveDateRequired = errors.New("effective date is required")
	ErrInvalidScheduleDay   = errors.New("schedule days must be integers 0-6 (Sun-Sat)")
)

type Builder struct {
	id             uuid.UUID
	trackingItemID uuid.UUID
	schedule       []int
	effectiveDate  time.Time
	createdAt      time.Time
}

func NewBuilder() *Builder { return &Builder{} }

func (b *Builder) SetId(id uuid.UUID) *Builder                  { b.id = id; return b }
func (b *Builder) SetTrackingItemID(id uuid.UUID) *Builder       { b.trackingItemID = id; return b }
func (b *Builder) SetSchedule(s []int) *Builder                  { b.schedule = s; return b }
func (b *Builder) SetEffectiveDate(d time.Time) *Builder         { b.effectiveDate = d; return b }
func (b *Builder) SetCreatedAt(t time.Time) *Builder             { b.createdAt = t; return b }

func (b *Builder) Build() (Model, error) {
	if b.trackingItemID == uuid.Nil {
		return Model{}, ErrTrackingItemRequired
	}
	if b.effectiveDate.IsZero() {
		return Model{}, ErrEffectiveDateRequired
	}
	for _, d := range b.schedule {
		if d < 0 || d > 6 {
			return Model{}, ErrInvalidScheduleDay
		}
	}
	return Model{
		id:             b.id,
		trackingItemID: b.trackingItemID,
		schedule:       b.schedule,
		effectiveDate:  b.effectiveDate,
		createdAt:      b.createdAt,
	}, nil
}
