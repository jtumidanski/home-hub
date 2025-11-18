package meal

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Model represents an immutable meal in the domain.
// All fields are private to enforce immutability.
type Model struct {
	id                 uuid.UUID
	householdId        uuid.UUID
	userId             uuid.UUID
	title              string
	description        string
	rawIngredientText  string
	createdAt          time.Time
	updatedAt          time.Time
}

// Id returns the meal's unique identifier
func (m Model) Id() uuid.UUID {
	return m.id
}

// HouseholdId returns the ID of the household this meal belongs to
func (m Model) HouseholdId() uuid.UUID {
	return m.householdId
}

// UserId returns the ID of the user who created this meal
func (m Model) UserId() uuid.UUID {
	return m.userId
}

// Title returns the meal's title
func (m Model) Title() string {
	return m.title
}

// Description returns the meal's description
func (m Model) Description() string {
	return m.description
}

// RawIngredientText returns the original ingredient text
func (m Model) RawIngredientText() string {
	return m.rawIngredientText
}

// CreatedAt returns when the meal was created
func (m Model) CreatedAt() time.Time {
	return m.createdAt
}

// UpdatedAt returns when the meal was last updated
func (m Model) UpdatedAt() time.Time {
	return m.updatedAt
}

// HasDescription returns true if the meal has a description
func (m Model) HasDescription() bool {
	return m.description != ""
}

// String returns a string representation of the meal for debugging
func (m Model) String() string {
	return fmt.Sprintf("Meal[id=%s, household=%s, title=%s, createdAt=%s]",
		m.id.String(), m.householdId.String(), m.title, m.createdAt.Format(time.RFC3339))
}

// MarshalJSON implements json.Marshaler for the Model
func (m Model) MarshalJSON() ([]byte, error) {
	type alias struct {
		Id                uuid.UUID `json:"id"`
		HouseholdId       uuid.UUID `json:"householdId"`
		UserId            uuid.UUID `json:"userId"`
		Title             string    `json:"title"`
		Description       string    `json:"description,omitempty"`
		RawIngredientText string    `json:"rawIngredientText"`
		CreatedAt         time.Time `json:"createdAt"`
		UpdatedAt         time.Time `json:"updatedAt"`
	}

	return json.Marshal(&alias{
		Id:                m.id,
		HouseholdId:       m.householdId,
		UserId:            m.userId,
		Title:             m.title,
		Description:       m.description,
		RawIngredientText: m.rawIngredientText,
		CreatedAt:         m.createdAt,
		UpdatedAt:         m.updatedAt,
	})
}

// UnmarshalJSON implements json.Unmarshaler for the Model
func (m *Model) UnmarshalJSON(data []byte) error {
	type alias struct {
		Id                uuid.UUID `json:"id"`
		HouseholdId       uuid.UUID `json:"householdId"`
		UserId            uuid.UUID `json:"userId"`
		Title             string    `json:"title"`
		Description       string    `json:"description,omitempty"`
		RawIngredientText string    `json:"rawIngredientText"`
		CreatedAt         time.Time `json:"createdAt"`
		UpdatedAt         time.Time `json:"updatedAt"`
	}

	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}

	m.id = a.Id
	m.householdId = a.HouseholdId
	m.userId = a.UserId
	m.title = a.Title
	m.description = a.Description
	m.rawIngredientText = a.RawIngredientText
	m.createdAt = a.CreatedAt
	m.updatedAt = a.UpdatedAt

	return nil
}

// Is returns true if the given model represents the same meal
func (m Model) Is(other Model) bool {
	return m.id == other.id
}
