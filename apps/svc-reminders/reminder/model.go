package reminder

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Model represents an immutable reminder in the domain.
// All fields are private to enforce immutability.
type Model struct {
	id          uuid.UUID
	name        string
	description string
	userId      uuid.UUID
	householdId uuid.UUID
	createdAt   time.Time
	remindAt    time.Time
	snoozeCount int
	status      Status
	dismissedAt *time.Time
	updatedAt   time.Time
}

// Id returns the reminder's unique identifier
func (m Model) Id() uuid.UUID {
	return m.id
}

// Name returns the reminder's name/title
func (m Model) Name() string {
	return m.name
}

// Description returns the reminder's description
func (m Model) Description() string {
	return m.description
}

// UserId returns the ID of the user who owns this reminder
func (m Model) UserId() uuid.UUID {
	return m.userId
}

// HouseholdId returns the ID of the household this reminder belongs to
func (m Model) HouseholdId() uuid.UUID {
	return m.householdId
}

// CreatedAt returns when the reminder was created
func (m Model) CreatedAt() time.Time {
	return m.createdAt
}

// RemindAt returns when the reminder should trigger
func (m Model) RemindAt() time.Time {
	return m.remindAt
}

// SnoozeCount returns the number of times this reminder has been snoozed
func (m Model) SnoozeCount() int {
	return m.snoozeCount
}

// Status returns the reminder's current status
func (m Model) Status() Status {
	return m.status
}

// DismissedAt returns when the reminder was dismissed, nil if not dismissed
func (m Model) DismissedAt() *time.Time {
	return m.dismissedAt
}

// UpdatedAt returns when the reminder was last updated
func (m Model) UpdatedAt() time.Time {
	return m.updatedAt
}

// IsActive returns true if the reminder is active
func (m Model) IsActive() bool {
	return m.status == StatusActive
}

// IsSnoozed returns true if the reminder is snoozed
func (m Model) IsSnoozed() bool {
	return m.status == StatusSnoozed
}

// IsDismissed returns true if the reminder is dismissed
func (m Model) IsDismissed() bool {
	return m.status == StatusDismissed
}

// HasDescription returns true if the reminder has a description
func (m Model) HasDescription() bool {
	return m.description != ""
}

// String returns a string representation of the reminder for debugging
func (m Model) String() string {
	dismissedStr := "not dismissed"
	if m.dismissedAt != nil {
		dismissedStr = m.dismissedAt.Format(time.RFC3339)
	}
	return fmt.Sprintf("Reminder[id=%s, userId=%s, name=%s, status=%s, remindAt=%s, snoozeCount=%d, dismissedAt=%s]",
		m.id.String(), m.userId.String(), m.name, m.status, m.remindAt.Format(time.RFC3339), m.snoozeCount, dismissedStr)
}

// MarshalJSON implements json.Marshaler for the Model
func (m Model) MarshalJSON() ([]byte, error) {
	type alias struct {
		Id          uuid.UUID  `json:"id"`
		Name        string     `json:"name"`
		Description string     `json:"description,omitempty"`
		UserId      uuid.UUID  `json:"userId"`
		HouseholdId uuid.UUID  `json:"householdId"`
		CreatedAt   time.Time  `json:"createdAt"`
		RemindAt    time.Time  `json:"remindAt"`
		SnoozeCount int        `json:"snoozeCount"`
		Status      Status     `json:"status"`
		DismissedAt *time.Time `json:"dismissedAt,omitempty"`
		UpdatedAt   time.Time  `json:"updatedAt"`
	}

	return json.Marshal(&alias{
		Id:          m.id,
		Name:        m.name,
		Description: m.description,
		UserId:      m.userId,
		HouseholdId: m.householdId,
		CreatedAt:   m.createdAt,
		RemindAt:    m.remindAt,
		SnoozeCount: m.snoozeCount,
		Status:      m.status,
		DismissedAt: m.dismissedAt,
		UpdatedAt:   m.updatedAt,
	})
}

// UnmarshalJSON implements json.Unmarshaler for the Model
func (m *Model) UnmarshalJSON(data []byte) error {
	type alias struct {
		Id          uuid.UUID  `json:"id"`
		Name        string     `json:"name"`
		Description string     `json:"description,omitempty"`
		UserId      uuid.UUID  `json:"userId"`
		HouseholdId uuid.UUID  `json:"householdId"`
		CreatedAt   time.Time  `json:"createdAt"`
		RemindAt    time.Time  `json:"remindAt"`
		SnoozeCount int        `json:"snoozeCount"`
		Status      Status     `json:"status"`
		DismissedAt *time.Time `json:"dismissedAt,omitempty"`
		UpdatedAt   time.Time  `json:"updatedAt"`
	}

	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}

	m.id = a.Id
	m.name = a.Name
	m.description = a.Description
	m.userId = a.UserId
	m.householdId = a.HouseholdId
	m.createdAt = a.CreatedAt
	m.remindAt = a.RemindAt
	m.snoozeCount = a.SnoozeCount
	m.status = a.Status
	m.dismissedAt = a.DismissedAt
	m.updatedAt = a.UpdatedAt

	return nil
}

// Is returns true if the given model represents the same reminder
func (m Model) Is(other Model) bool {
	return m.id == other.id
}
