package source

import (
	"time"

	"github.com/google/uuid"
)

type Model struct {
	id           uuid.UUID
	tenantID     uuid.UUID
	householdID  uuid.UUID
	connectionID uuid.UUID
	externalID   string
	name         string
	primary      bool
	visible      bool
	color        string
	syncToken    string
	createdAt    time.Time
	updatedAt    time.Time
}

func (m Model) Id() uuid.UUID           { return m.id }
func (m Model) TenantID() uuid.UUID     { return m.tenantID }
func (m Model) HouseholdID() uuid.UUID  { return m.householdID }
func (m Model) ConnectionID() uuid.UUID { return m.connectionID }
func (m Model) ExternalID() string      { return m.externalID }
func (m Model) Name() string            { return m.name }
func (m Model) Primary() bool           { return m.primary }
func (m Model) Visible() bool           { return m.visible }
func (m Model) Color() string           { return m.color }
func (m Model) SyncToken() string       { return m.syncToken }
func (m Model) CreatedAt() time.Time    { return m.createdAt }
func (m Model) UpdatedAt() time.Time    { return m.updatedAt }
