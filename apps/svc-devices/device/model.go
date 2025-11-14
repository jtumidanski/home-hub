package device

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Model represents an immutable device in the domain.
// All fields are private to enforce immutability.
type Model struct {
	id          uuid.UUID
	name        string
	deviceType  string
	householdId uuid.UUID
	createdAt   time.Time
	updatedAt   time.Time
}

// Id returns the device's unique identifier
func (m Model) Id() uuid.UUID {
	return m.id
}

// Name returns the device's name
func (m Model) Name() string {
	return m.name
}

// Type returns the device's type (e.g., "kiosk", "mobile", "tablet")
func (m Model) Type() string {
	return m.deviceType
}

// HouseholdId returns the ID of the household this device belongs to
func (m Model) HouseholdId() uuid.UUID {
	return m.householdId
}

// CreatedAt returns when the device was created
func (m Model) CreatedAt() time.Time {
	return m.createdAt
}

// UpdatedAt returns when the device was last updated
func (m Model) UpdatedAt() time.Time {
	return m.updatedAt
}

// String returns a string representation of the device for debugging
func (m Model) String() string {
	return fmt.Sprintf("Device[id=%s, name=%s, type=%s]", m.id.String(), m.name, m.deviceType)
}

// MarshalJSON implements json.Marshaler for the Model
func (m Model) MarshalJSON() ([]byte, error) {
	type alias struct {
		Id          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		Type        string    `json:"type"`
		HouseholdId uuid.UUID `json:"householdId"`
		CreatedAt   time.Time `json:"createdAt"`
		UpdatedAt   time.Time `json:"updatedAt"`
	}

	return json.Marshal(&alias{
		Id:          m.id,
		Name:        m.name,
		Type:        m.deviceType,
		HouseholdId: m.householdId,
		CreatedAt:   m.createdAt,
		UpdatedAt:   m.updatedAt,
	})
}

// UnmarshalJSON implements json.Unmarshaler for the Model
func (m *Model) UnmarshalJSON(data []byte) error {
	type alias struct {
		Id          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		Type        string    `json:"type"`
		HouseholdId uuid.UUID `json:"householdId"`
		CreatedAt   time.Time `json:"createdAt"`
		UpdatedAt   time.Time `json:"updatedAt"`
	}

	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}

	m.id = a.Id
	m.name = a.Name
	m.deviceType = a.Type
	m.householdId = a.HouseholdId
	m.createdAt = a.CreatedAt
	m.updatedAt = a.UpdatedAt

	return nil
}

// Is returns true if the given model represents the same device
func (m Model) Is(other Model) bool {
	return m.id == other.id
}
