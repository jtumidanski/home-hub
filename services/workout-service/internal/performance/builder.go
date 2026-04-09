package performance

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidStatus     = errors.New("status must be one of: pending, done, skipped, partial")
	ErrInvalidMode       = errors.New("mode must be one of: summary, per_set")
	ErrInvalidWeightUnit = errors.New("weightUnit must be one of: lb, kg")
	ErrInvalidNumeric    = errors.New("numeric actuals must be non-negative")
)

var validStatuses = map[string]bool{
	StatusPending: true, StatusDone: true, StatusSkipped: true, StatusPartial: true,
}
var validModes = map[string]bool{ModeSummary: true, ModePerSet: true}
var validWeightUnits = map[string]bool{"lb": true, "kg": true}

func ValidStatus(s string) bool     { return validStatuses[s] }
func ValidMode(m string) bool       { return validModes[m] }
func ValidWeightUnit(u string) bool { return validWeightUnits[u] }

type Builder struct {
	id                    uuid.UUID
	tenantID              uuid.UUID
	userID                uuid.UUID
	plannedItemID         uuid.UUID
	status                string
	mode                  string
	weightUnit            *string
	actualSets            *int
	actualReps            *int
	actualWeight          *float64
	actualDurationSeconds *int
	actualDistance        *float64
	actualDistanceUnit    *string
	notes                 *string
	createdAt             time.Time
	updatedAt             time.Time
}

func NewBuilder() *Builder { return &Builder{status: StatusPending, mode: ModeSummary} }

func (b *Builder) SetId(id uuid.UUID) *Builder                 { b.id = id; return b }
func (b *Builder) SetTenantID(id uuid.UUID) *Builder           { b.tenantID = id; return b }
func (b *Builder) SetUserID(id uuid.UUID) *Builder             { b.userID = id; return b }
func (b *Builder) SetPlannedItemID(id uuid.UUID) *Builder      { b.plannedItemID = id; return b }
func (b *Builder) SetStatus(s string) *Builder                  { if s != "" { b.status = s }; return b }
func (b *Builder) SetMode(m string) *Builder                    { if m != "" { b.mode = m }; return b }
func (b *Builder) SetWeightUnit(v *string) *Builder             { b.weightUnit = v; return b }
func (b *Builder) SetActualSets(v *int) *Builder                { b.actualSets = v; return b }
func (b *Builder) SetActualReps(v *int) *Builder                { b.actualReps = v; return b }
func (b *Builder) SetActualWeight(v *float64) *Builder          { b.actualWeight = v; return b }
func (b *Builder) SetActualDurationSeconds(v *int) *Builder     { b.actualDurationSeconds = v; return b }
func (b *Builder) SetActualDistance(v *float64) *Builder        { b.actualDistance = v; return b }
func (b *Builder) SetActualDistanceUnit(v *string) *Builder     { b.actualDistanceUnit = v; return b }
func (b *Builder) SetNotes(v *string) *Builder                  { b.notes = v; return b }
func (b *Builder) SetCreatedAt(t time.Time) *Builder            { b.createdAt = t; return b }
func (b *Builder) SetUpdatedAt(t time.Time) *Builder            { b.updatedAt = t; return b }

func (b *Builder) Build() (Model, error) {
	if !validStatuses[b.status] {
		return Model{}, ErrInvalidStatus
	}
	if !validModes[b.mode] {
		return Model{}, ErrInvalidMode
	}
	if b.weightUnit != nil && *b.weightUnit != "" && !validWeightUnits[*b.weightUnit] {
		return Model{}, ErrInvalidWeightUnit
	}
	for _, v := range []*int{b.actualSets, b.actualReps, b.actualDurationSeconds} {
		if v != nil && *v < 0 {
			return Model{}, ErrInvalidNumeric
		}
	}
	for _, v := range []*float64{b.actualWeight, b.actualDistance} {
		if v != nil && *v < 0 {
			return Model{}, ErrInvalidNumeric
		}
	}
	return Model{
		id:                    b.id,
		tenantID:              b.tenantID,
		userID:                b.userID,
		plannedItemID:         b.plannedItemID,
		status:                b.status,
		mode:                  b.mode,
		weightUnit:            b.weightUnit,
		actualSets:            b.actualSets,
		actualReps:            b.actualReps,
		actualWeight:          b.actualWeight,
		actualDurationSeconds: b.actualDurationSeconds,
		actualDistance:        b.actualDistance,
		actualDistanceUnit:    b.actualDistanceUnit,
		notes:                 b.notes,
		createdAt:             b.createdAt,
		updatedAt:             b.updatedAt,
	}, nil
}
