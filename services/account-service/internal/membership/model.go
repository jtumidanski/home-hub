package membership

import (
	"time"

	"github.com/google/uuid"
)

type Model struct {
	id          uuid.UUID
	tenantID    uuid.UUID
	householdID uuid.UUID
	userID      uuid.UUID
	role        string
	createdAt   time.Time
	updatedAt   time.Time
}

func (m Model) Id() uuid.UUID          { return m.id }
func (m Model) TenantID() uuid.UUID    { return m.tenantID }
func (m Model) HouseholdID() uuid.UUID { return m.householdID }
func (m Model) UserID() uuid.UUID      { return m.userID }
func (m Model) Role() string           { return m.role }
func (m Model) CreatedAt() time.Time   { return m.createdAt }
func (m Model) UpdatedAt() time.Time   { return m.updatedAt }
