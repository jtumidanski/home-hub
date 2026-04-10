package planneditem

import (
	"time"

	"github.com/google/uuid"
)

type Model struct {
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

func (m Model) Id() uuid.UUID                 { return m.id }
func (m Model) TenantID() uuid.UUID           { return m.tenantID }
func (m Model) UserID() uuid.UUID             { return m.userID }
func (m Model) WeekID() uuid.UUID             { return m.weekID }
func (m Model) ExerciseID() uuid.UUID         { return m.exerciseID }
func (m Model) DayOfWeek() int                { return m.dayOfWeek }
func (m Model) Position() int                 { return m.position }
func (m Model) PlannedSets() *int             { return m.plannedSets }
func (m Model) PlannedReps() *int             { return m.plannedReps }
func (m Model) PlannedWeight() *float64       { return m.plannedWeight }
func (m Model) PlannedWeightUnit() *string    { return m.plannedWeightUnit }
func (m Model) PlannedDurationSeconds() *int  { return m.plannedDurationSeconds }
func (m Model) PlannedDistance() *float64     { return m.plannedDistance }
func (m Model) PlannedDistanceUnit() *string  { return m.plannedDistanceUnit }
func (m Model) Notes() *string                { return m.notes }
func (m Model) CreatedAt() time.Time          { return m.createdAt }
func (m Model) UpdatedAt() time.Time          { return m.updatedAt }
