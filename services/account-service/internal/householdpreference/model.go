package householdpreference

import (
	"time"

	"github.com/google/uuid"
)

type Model struct {
	id                 uuid.UUID
	tenantID           uuid.UUID
	userID             uuid.UUID
	householdID        uuid.UUID
	defaultDashboardID *uuid.UUID
	createdAt          time.Time
	updatedAt          time.Time
}

func (m Model) Id() uuid.UUID                 { return m.id }
func (m Model) TenantID() uuid.UUID           { return m.tenantID }
func (m Model) UserID() uuid.UUID             { return m.userID }
func (m Model) HouseholdID() uuid.UUID        { return m.householdID }
func (m Model) DefaultDashboardID() *uuid.UUID { return m.defaultDashboardID }
func (m Model) CreatedAt() time.Time          { return m.createdAt }
func (m Model) UpdatedAt() time.Time          { return m.updatedAt }
