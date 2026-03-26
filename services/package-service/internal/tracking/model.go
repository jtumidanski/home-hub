package tracking

import (
	"time"

	"github.com/google/uuid"
)

type Model struct {
	id                 uuid.UUID
	tenantID           uuid.UUID
	householdID        uuid.UUID
	userID             uuid.UUID
	trackingNumber     string
	carrier            string
	label              *string
	notes              *string
	status             string
	private            bool
	estimatedDelivery  *time.Time
	actualDelivery     *time.Time
	lastPolledAt       *time.Time
	lastStatusChangeAt *time.Time
	archivedAt         *time.Time
	createdAt          time.Time
	updatedAt          time.Time
}

func (m Model) Id() uuid.UUID                  { return m.id }
func (m Model) TenantID() uuid.UUID            { return m.tenantID }
func (m Model) HouseholdID() uuid.UUID         { return m.householdID }
func (m Model) UserID() uuid.UUID              { return m.userID }
func (m Model) TrackingNumber() string          { return m.trackingNumber }
func (m Model) Carrier() string                 { return m.carrier }
func (m Model) Label() *string                  { return m.label }
func (m Model) Notes() *string                  { return m.notes }
func (m Model) Status() string                  { return m.status }
func (m Model) Private() bool                   { return m.private }
func (m Model) EstimatedDelivery() *time.Time   { return m.estimatedDelivery }
func (m Model) ActualDelivery() *time.Time      { return m.actualDelivery }
func (m Model) LastPolledAt() *time.Time        { return m.lastPolledAt }
func (m Model) LastStatusChangeAt() *time.Time  { return m.lastStatusChangeAt }
func (m Model) ArchivedAt() *time.Time          { return m.archivedAt }
func (m Model) CreatedAt() time.Time            { return m.createdAt }
func (m Model) UpdatedAt() time.Time            { return m.updatedAt }

func (m Model) IsPrivate() bool {
	return m.private
}

func (m Model) IsArchived() bool {
	return m.status == StatusArchived
}

func (m Model) IsPolling() bool {
	return m.status == StatusPreTransit || m.status == StatusInTransit || m.status == StatusOutForDelivery
}
