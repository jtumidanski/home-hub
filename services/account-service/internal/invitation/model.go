package invitation

import (
	"time"

	"github.com/google/uuid"
)

type Model struct {
	id          uuid.UUID
	tenantID    uuid.UUID
	householdID uuid.UUID
	email       string
	role        string
	status      string
	invitedBy   uuid.UUID
	expiresAt   time.Time
	createdAt   time.Time
	updatedAt   time.Time
}

func (m Model) Id() uuid.UUID          { return m.id }
func (m Model) TenantID() uuid.UUID    { return m.tenantID }
func (m Model) HouseholdID() uuid.UUID { return m.householdID }
func (m Model) Email() string          { return m.email }
func (m Model) Role() string           { return m.role }
func (m Model) Status() string         { return m.status }
func (m Model) InvitedBy() uuid.UUID   { return m.invitedBy }
func (m Model) ExpiresAt() time.Time   { return m.expiresAt }
func (m Model) CreatedAt() time.Time   { return m.createdAt }
func (m Model) UpdatedAt() time.Time   { return m.updatedAt }
