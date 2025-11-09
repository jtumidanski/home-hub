package household

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Model represents an immutable household in the domain.
// All fields are private to enforce immutability.
type Model struct {
	id        uuid.UUID
	name      string
	createdAt time.Time
	updatedAt time.Time
}

// Id returns the household's unique identifier
func (m Model) Id() uuid.UUID {
	return m.id
}

// Name returns the household's name
func (m Model) Name() string {
	return m.name
}

// CreatedAt returns when the household was created
func (m Model) CreatedAt() time.Time {
	return m.createdAt
}

// UpdatedAt returns when the household was last updated
func (m Model) UpdatedAt() time.Time {
	return m.updatedAt
}

// String returns a string representation of the household for debugging
func (m Model) String() string {
	return fmt.Sprintf("Household[id=%s, name=%s]", m.id.String(), m.name)
}

// MarshalJSON implements json.Marshaler for the Model
func (m Model) MarshalJSON() ([]byte, error) {
	type alias struct {
		Id        uuid.UUID `json:"id"`
		Name      string    `json:"name"`
		CreatedAt time.Time `json:"createdAt"`
		UpdatedAt time.Time `json:"updatedAt"`
	}

	return json.Marshal(&alias{
		Id:        m.id,
		Name:      m.name,
		CreatedAt: m.createdAt,
		UpdatedAt: m.updatedAt,
	})
}

// UnmarshalJSON implements json.Unmarshaler for the Model
func (m *Model) UnmarshalJSON(data []byte) error {
	type alias struct {
		Id        uuid.UUID `json:"id"`
		Name      string    `json:"name"`
		CreatedAt time.Time `json:"createdAt"`
		UpdatedAt time.Time `json:"updatedAt"`
	}

	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}

	m.id = a.Id
	m.name = a.Name
	m.createdAt = a.CreatedAt
	m.updatedAt = a.UpdatedAt

	return nil
}

// Is returns true if the given model represents the same household
func (m Model) Is(other Model) bool {
	return m.id == other.id
}
