package task

import (
	"time"

	"github.com/google/uuid"
)

type Model struct {
	id              uuid.UUID
	tenantID        uuid.UUID
	householdID     uuid.UUID
	title           string
	notes           string
	status          string
	dueOn           *time.Time
	rolloverEnabled bool
	ownerUserID     *uuid.UUID
	completedAt     *time.Time
	completedByUID  *uuid.UUID
	deletedAt       *time.Time
	createdAt       time.Time
	updatedAt       time.Time
}

func (m Model) Id() uuid.UUID            { return m.id }
func (m Model) TenantID() uuid.UUID      { return m.tenantID }
func (m Model) HouseholdID() uuid.UUID   { return m.householdID }
func (m Model) Title() string            { return m.title }
func (m Model) Notes() string            { return m.notes }
func (m Model) Status() string           { return m.status }
func (m Model) DueOn() *time.Time        { return m.dueOn }
func (m Model) RolloverEnabled() bool    { return m.rolloverEnabled }
func (m Model) OwnerUserID() *uuid.UUID    { return m.ownerUserID }
func (m Model) CompletedAt() *time.Time  { return m.completedAt }
func (m Model) CompletedByUID() *uuid.UUID { return m.completedByUID }
func (m Model) DeletedAt() *time.Time    { return m.deletedAt }
func (m Model) CreatedAt() time.Time     { return m.createdAt }
func (m Model) UpdatedAt() time.Time     { return m.updatedAt }
func (m Model) IsDeleted() bool          { return m.deletedAt != nil }
func (m Model) IsCompleted() bool        { return m.status == "completed" }

