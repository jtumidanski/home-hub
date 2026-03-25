package household

import (
	"time"

	"github.com/google/uuid"
)

type Model struct {
	id           uuid.UUID
	tenantID     uuid.UUID
	name         string
	timezone     string
	units        string
	latitude     *float64
	longitude    *float64
	locationName *string
	createdAt    time.Time
	updatedAt    time.Time
}

func (m Model) Id() uuid.UUID          { return m.id }
func (m Model) TenantID() uuid.UUID    { return m.tenantID }
func (m Model) Name() string           { return m.name }
func (m Model) Timezone() string       { return m.timezone }
func (m Model) Units() string          { return m.units }
func (m Model) Latitude() *float64     { return m.latitude }
func (m Model) Longitude() *float64    { return m.longitude }
func (m Model) LocationName() *string  { return m.locationName }
func (m Model) CreatedAt() time.Time   { return m.createdAt }
func (m Model) UpdatedAt() time.Time   { return m.updatedAt }

func (m Model) HasLocation() bool {
	return m.latitude != nil && m.longitude != nil
}
