package task

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Model represents an immutable task in the domain.
// All fields are private to enforce immutability.
type Model struct {
	id          uuid.UUID
	userId      uuid.UUID
	householdId uuid.UUID
	day         time.Time
	title       string
	description string
	status      Status
	createdAt   time.Time
	completedAt *time.Time
	updatedAt   time.Time
}

// Id returns the task's unique identifier
func (m Model) Id() uuid.UUID {
	return m.id
}

// UserId returns the ID of the user who owns this task
func (m Model) UserId() uuid.UUID {
	return m.userId
}

// HouseholdId returns the ID of the household this task belongs to
func (m Model) HouseholdId() uuid.UUID {
	return m.householdId
}

// Day returns the date this task is scheduled for
func (m Model) Day() time.Time {
	return m.day
}

// Title returns the task's title
func (m Model) Title() string {
	return m.title
}

// Description returns the task's description
func (m Model) Description() string {
	return m.description
}

// Status returns the task's current status
func (m Model) Status() Status {
	return m.status
}

// CreatedAt returns when the task was created
func (m Model) CreatedAt() time.Time {
	return m.createdAt
}

// CompletedAt returns when the task was completed, nil if not completed
func (m Model) CompletedAt() *time.Time {
	return m.completedAt
}

// UpdatedAt returns when the task was last updated
func (m Model) UpdatedAt() time.Time {
	return m.updatedAt
}

// IsComplete returns true if the task is completed
func (m Model) IsComplete() bool {
	return m.status == StatusComplete
}

// IsIncomplete returns true if the task is not completed
func (m Model) IsIncomplete() bool {
	return m.status == StatusIncomplete
}

// HasDescription returns true if the task has a description
func (m Model) HasDescription() bool {
	return m.description != ""
}

// String returns a string representation of the task for debugging
func (m Model) String() string {
	completedStr := "not completed"
	if m.completedAt != nil {
		completedStr = m.completedAt.Format(time.RFC3339)
	}
	return fmt.Sprintf("Task[id=%s, userId=%s, title=%s, status=%s, day=%s, completedAt=%s]",
		m.id.String(), m.userId.String(), m.title, m.status, m.day.Format("2006-01-02"), completedStr)
}

// MarshalJSON implements json.Marshaler for the Model
func (m Model) MarshalJSON() ([]byte, error) {
	type alias struct {
		Id          uuid.UUID  `json:"id"`
		UserId      uuid.UUID  `json:"userId"`
		HouseholdId uuid.UUID  `json:"householdId"`
		Day         string     `json:"day"` // Format as date string
		Title       string     `json:"title"`
		Description string     `json:"description,omitempty"`
		Status      Status     `json:"status"`
		CreatedAt   time.Time  `json:"createdAt"`
		CompletedAt *time.Time `json:"completedAt,omitempty"`
		UpdatedAt   time.Time  `json:"updatedAt"`
	}

	return json.Marshal(&alias{
		Id:          m.id,
		UserId:      m.userId,
		HouseholdId: m.householdId,
		Day:         m.day.Format("2006-01-02"),
		Title:       m.title,
		Description: m.description,
		Status:      m.status,
		CreatedAt:   m.createdAt,
		CompletedAt: m.completedAt,
		UpdatedAt:   m.updatedAt,
	})
}

// UnmarshalJSON implements json.Unmarshaler for the Model
func (m *Model) UnmarshalJSON(data []byte) error {
	type alias struct {
		Id          uuid.UUID  `json:"id"`
		UserId      uuid.UUID  `json:"userId"`
		HouseholdId uuid.UUID  `json:"householdId"`
		Day         string     `json:"day"`
		Title       string     `json:"title"`
		Description string     `json:"description,omitempty"`
		Status      Status     `json:"status"`
		CreatedAt   time.Time  `json:"createdAt"`
		CompletedAt *time.Time `json:"completedAt,omitempty"`
		UpdatedAt   time.Time  `json:"updatedAt"`
	}

	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}

	day, err := time.Parse("2006-01-02", a.Day)
	if err != nil {
		return fmt.Errorf("invalid day format: %w", err)
	}

	m.id = a.Id
	m.userId = a.UserId
	m.householdId = a.HouseholdId
	m.day = day
	m.title = a.Title
	m.description = a.Description
	m.status = a.Status
	m.createdAt = a.CreatedAt
	m.completedAt = a.CompletedAt
	m.updatedAt = a.UpdatedAt

	return nil
}

// Is returns true if the given model represents the same task
func (m Model) Is(other Model) bool {
	return m.id == other.id
}
