package planneditem

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidDayOfWeek = errors.New("dayOfWeek must be in [0,6]")
	ErrInvalidPosition  = errors.New("position must be non-negative")
	ErrInvalidNumeric   = errors.New("planned numeric values must be non-negative")
)

type Builder struct {
	id                     uuid.UUID
	tenantID               uuid.UUID
	userID                 uuid.UUID
	weekID                 uuid.UUID
	exerciseID             uuid.UUID
	dayOfWeek              int
	position               int
	plannedSets            *int
	plannedReps            *int
	plannedWeight          *float64
	plannedWeightUnit      *string
	plannedDurationSeconds *int
	plannedDistance        *float64
	plannedDistanceUnit    *string
	notes                  *string
	createdAt              time.Time
	updatedAt              time.Time
}

func NewBuilder() *Builder { return &Builder{} }

func (b *Builder) SetId(id uuid.UUID) *Builder                  { b.id = id; return b }
func (b *Builder) SetTenantID(id uuid.UUID) *Builder            { b.tenantID = id; return b }
func (b *Builder) SetUserID(id uuid.UUID) *Builder              { b.userID = id; return b }
func (b *Builder) SetWeekID(id uuid.UUID) *Builder              { b.weekID = id; return b }
func (b *Builder) SetExerciseID(id uuid.UUID) *Builder          { b.exerciseID = id; return b }
func (b *Builder) SetDayOfWeek(d int) *Builder                   { b.dayOfWeek = d; return b }
func (b *Builder) SetPosition(p int) *Builder                    { b.position = p; return b }
func (b *Builder) SetPlannedSets(v *int) *Builder                { b.plannedSets = v; return b }
func (b *Builder) SetPlannedReps(v *int) *Builder                { b.plannedReps = v; return b }
func (b *Builder) SetPlannedWeight(v *float64) *Builder          { b.plannedWeight = v; return b }
func (b *Builder) SetPlannedWeightUnit(v *string) *Builder       { b.plannedWeightUnit = v; return b }
func (b *Builder) SetPlannedDurationSeconds(v *int) *Builder     { b.plannedDurationSeconds = v; return b }
func (b *Builder) SetPlannedDistance(v *float64) *Builder        { b.plannedDistance = v; return b }
func (b *Builder) SetPlannedDistanceUnit(v *string) *Builder     { b.plannedDistanceUnit = v; return b }
func (b *Builder) SetNotes(v *string) *Builder                   { b.notes = v; return b }
func (b *Builder) SetCreatedAt(t time.Time) *Builder             { b.createdAt = t; return b }
func (b *Builder) SetUpdatedAt(t time.Time) *Builder             { b.updatedAt = t; return b }

func (b *Builder) Build() (Model, error) {
	if b.dayOfWeek < 0 || b.dayOfWeek > 6 {
		return Model{}, ErrInvalidDayOfWeek
	}
	if b.position < 0 {
		return Model{}, ErrInvalidPosition
	}
	for _, v := range []*int{b.plannedSets, b.plannedReps, b.plannedDurationSeconds} {
		if v != nil && *v < 0 {
			return Model{}, ErrInvalidNumeric
		}
	}
	for _, v := range []*float64{b.plannedWeight, b.plannedDistance} {
		if v != nil && *v < 0 {
			return Model{}, ErrInvalidNumeric
		}
	}
	return Model{
		id:                     b.id,
		tenantID:               b.tenantID,
		userID:                 b.userID,
		weekID:                 b.weekID,
		exerciseID:             b.exerciseID,
		dayOfWeek:              b.dayOfWeek,
		position:               b.position,
		plannedSets:            b.plannedSets,
		plannedReps:            b.plannedReps,
		plannedWeight:          b.plannedWeight,
		plannedWeightUnit:      b.plannedWeightUnit,
		plannedDurationSeconds: b.plannedDurationSeconds,
		plannedDistance:        b.plannedDistance,
		plannedDistanceUnit:    b.plannedDistanceUnit,
		notes:                  b.notes,
		createdAt:              b.createdAt,
		updatedAt:              b.updatedAt,
	}, nil
}
