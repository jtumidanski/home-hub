package performance

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusPending = "pending"
	StatusDone    = "done"
	StatusSkipped = "skipped"
	StatusPartial = "partial"

	ModeSummary = "summary"
	ModePerSet  = "per_set"
)

type Model struct {
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

func (m Model) Id() uuid.UUID                 { return m.id }
func (m Model) TenantID() uuid.UUID           { return m.tenantID }
func (m Model) UserID() uuid.UUID             { return m.userID }
func (m Model) PlannedItemID() uuid.UUID      { return m.plannedItemID }
func (m Model) Status() string                { return m.status }
func (m Model) Mode() string                  { return m.mode }
func (m Model) WeightUnit() *string           { return m.weightUnit }
func (m Model) ActualSets() *int              { return m.actualSets }
func (m Model) ActualReps() *int              { return m.actualReps }
func (m Model) ActualWeight() *float64        { return m.actualWeight }
func (m Model) ActualDurationSeconds() *int   { return m.actualDurationSeconds }
func (m Model) ActualDistance() *float64      { return m.actualDistance }
func (m Model) ActualDistanceUnit() *string   { return m.actualDistanceUnit }
func (m Model) Notes() *string                { return m.notes }
func (m Model) CreatedAt() time.Time          { return m.createdAt }
func (m Model) UpdatedAt() time.Time          { return m.updatedAt }

type SetModel struct {
	id            uuid.UUID
	tenantID      uuid.UUID
	userID        uuid.UUID
	performanceID uuid.UUID
	setNumber     int
	reps          int
	weight        float64
	createdAt     time.Time
}

func (m SetModel) Id() uuid.UUID            { return m.id }
func (m SetModel) PerformanceID() uuid.UUID { return m.performanceID }
func (m SetModel) SetNumber() int           { return m.setNumber }
func (m SetModel) Reps() int                { return m.reps }
func (m SetModel) Weight() float64          { return m.weight }
func (m SetModel) CreatedAt() time.Time     { return m.createdAt }
