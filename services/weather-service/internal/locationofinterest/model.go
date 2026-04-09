package locationofinterest

import (
	"time"

	"github.com/google/uuid"
)

type Model struct {
	id          uuid.UUID
	tenantID    uuid.UUID
	householdID uuid.UUID
	label       *string
	placeName   string
	latitude    float64
	longitude   float64
	createdAt   time.Time
	updatedAt   time.Time
}

func (m Model) Id() uuid.UUID          { return m.id }
func (m Model) TenantID() uuid.UUID    { return m.tenantID }
func (m Model) HouseholdID() uuid.UUID { return m.householdID }
func (m Model) Label() *string         { return m.label }
func (m Model) PlaceName() string      { return m.placeName }
func (m Model) Latitude() float64      { return m.latitude }
func (m Model) Longitude() float64     { return m.longitude }
func (m Model) CreatedAt() time.Time   { return m.createdAt }
func (m Model) UpdatedAt() time.Time   { return m.updatedAt }
