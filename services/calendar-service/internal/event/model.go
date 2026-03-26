package event

import (
	"time"

	"github.com/google/uuid"
)

type Model struct {
	id              uuid.UUID
	tenantID        uuid.UUID
	householdID     uuid.UUID
	connectionID    uuid.UUID
	sourceID        uuid.UUID
	userID          uuid.UUID
	externalID      string
	title           string
	description     string
	startTime       time.Time
	endTime         time.Time
	allDay          bool
	location        string
	visibility      string
	userDisplayName string
	userColor       string
	createdAt       time.Time
	updatedAt       time.Time
}

func (m Model) Id() uuid.UUID             { return m.id }
func (m Model) TenantID() uuid.UUID       { return m.tenantID }
func (m Model) HouseholdID() uuid.UUID    { return m.householdID }
func (m Model) ConnectionID() uuid.UUID   { return m.connectionID }
func (m Model) SourceID() uuid.UUID       { return m.sourceID }
func (m Model) UserID() uuid.UUID         { return m.userID }
func (m Model) ExternalID() string        { return m.externalID }
func (m Model) Title() string             { return m.title }
func (m Model) Description() string       { return m.description }
func (m Model) StartTime() time.Time      { return m.startTime }
func (m Model) EndTime() time.Time        { return m.endTime }
func (m Model) AllDay() bool              { return m.allDay }
func (m Model) Location() string          { return m.location }
func (m Model) Visibility() string        { return m.visibility }
func (m Model) UserDisplayName() string   { return m.userDisplayName }
func (m Model) UserColor() string         { return m.userColor }
func (m Model) CreatedAt() time.Time      { return m.createdAt }
func (m Model) UpdatedAt() time.Time      { return m.updatedAt }

func (m Model) IsPrivate() bool {
	return m.visibility == "private" || m.visibility == "confidential"
}
