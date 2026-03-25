package reminder

import (
	"time"

	"github.com/google/uuid"
)

type Model struct {
	id               uuid.UUID
	tenantID         uuid.UUID
	householdID      uuid.UUID
	title            string
	notes            string
	scheduledFor     time.Time
	lastDismissedAt  *time.Time
	lastSnoozedUntil *time.Time
	createdAt        time.Time
	updatedAt        time.Time
}

func (m Model) Id() uuid.UUID               { return m.id }
func (m Model) TenantID() uuid.UUID          { return m.tenantID }
func (m Model) HouseholdID() uuid.UUID       { return m.householdID }
func (m Model) Title() string                { return m.title }
func (m Model) Notes() string                { return m.notes }
func (m Model) ScheduledFor() time.Time      { return m.scheduledFor }
func (m Model) LastDismissedAt() *time.Time   { return m.lastDismissedAt }
func (m Model) LastSnoozedUntil() *time.Time  { return m.lastSnoozedUntil }
func (m Model) CreatedAt() time.Time         { return m.createdAt }
func (m Model) UpdatedAt() time.Time         { return m.updatedAt }

func (m Model) IsActive() bool {
	now := time.Now().UTC()
	if m.lastDismissedAt != nil {
		return false
	}
	if m.lastSnoozedUntil != nil && m.lastSnoozedUntil.After(now) {
		return false
	}
	return m.scheduledFor.Before(now) || m.scheduledFor.Equal(now)
}
