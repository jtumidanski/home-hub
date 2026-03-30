package plan

import (
	"time"

	"github.com/google/uuid"
)

type Model struct {
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

func (m Model) Id() uuid.UUID          { return m.id }
func (m Model) TenantID() uuid.UUID    { return m.tenantID }
func (m Model) HouseholdID() uuid.UUID { return m.householdID }
func (m Model) StartsOn() time.Time    { return m.startsOn }
func (m Model) Name() string           { return m.name }
func (m Model) Locked() bool           { return m.locked }
func (m Model) CreatedBy() uuid.UUID   { return m.createdBy }
func (m Model) CreatedAt() time.Time   { return m.createdAt }
func (m Model) UpdatedAt() time.Time   { return m.updatedAt }
