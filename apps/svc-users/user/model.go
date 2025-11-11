package user

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
)

// emailRegex is a simple email validation pattern
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// Model represents an immutable user in the domain.
// All fields are private to enforce immutability.
type Model struct {
	id          uuid.UUID
	email       string
	displayName string
	provider    string // OAuth provider: 'google' or 'github'
	householdId *uuid.UUID
	createdAt   time.Time
	updatedAt   time.Time
}

// Id returns the user's unique identifier
func (m Model) Id() uuid.UUID {
	return m.id
}

// Email returns the user's email address
func (m Model) Email() string {
	return m.email
}

// DisplayName returns the user's display name
func (m Model) DisplayName() string {
	return m.displayName
}

// Provider returns the OAuth provider used for authentication
func (m Model) Provider() string {
	return m.provider
}

// HouseholdId returns the user's household ID if associated, nil otherwise
func (m Model) HouseholdId() *uuid.UUID {
	return m.householdId
}

// CreatedAt returns when the user was created
func (m Model) CreatedAt() time.Time {
	return m.createdAt
}

// UpdatedAt returns when the user was last updated
func (m Model) UpdatedAt() time.Time {
	return m.updatedAt
}

// HasHousehold returns true if the user is associated with a household
func (m Model) HasHousehold() bool {
	return m.householdId != nil
}

// ValidEmail returns true if the email has a valid format
func (m Model) ValidEmail() bool {
	return emailRegex.MatchString(m.email)
}

// String returns a string representation of the user for debugging
func (m Model) String() string {
	householdStr := "none"
	if m.householdId != nil {
		householdStr = m.householdId.String()
	}
	return fmt.Sprintf("User[id=%s, email=%s, displayName=%s, provider=%s, householdId=%s]",
		m.id.String(), m.email, m.displayName, m.provider, householdStr)
}

// MarshalJSON implements json.Marshaler for the Model
func (m Model) MarshalJSON() ([]byte, error) {
	type alias struct {
		Id          uuid.UUID  `json:"id"`
		Email       string     `json:"email"`
		DisplayName string     `json:"displayName"`
		Provider    string     `json:"provider"`
		HouseholdId *uuid.UUID `json:"householdId,omitempty"`
		CreatedAt   time.Time  `json:"createdAt"`
		UpdatedAt   time.Time  `json:"updatedAt"`
	}

	return json.Marshal(&alias{
		Id:          m.id,
		Email:       m.email,
		DisplayName: m.displayName,
		Provider:    m.provider,
		HouseholdId: m.householdId,
		CreatedAt:   m.createdAt,
		UpdatedAt:   m.updatedAt,
	})
}

// UnmarshalJSON implements json.Unmarshaler for the Model
func (m *Model) UnmarshalJSON(data []byte) error {
	type alias struct {
		Id          uuid.UUID  `json:"id"`
		Email       string     `json:"email"`
		DisplayName string     `json:"displayName"`
		Provider    string     `json:"provider"`
		HouseholdId *uuid.UUID `json:"householdId,omitempty"`
		CreatedAt   time.Time  `json:"createdAt"`
		UpdatedAt   time.Time  `json:"updatedAt"`
	}

	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}

	m.id = a.Id
	m.email = a.Email
	m.displayName = a.DisplayName
	m.provider = a.Provider
	m.householdId = a.HouseholdId
	m.createdAt = a.CreatedAt
	m.updatedAt = a.UpdatedAt

	return nil
}

// Is returns true if the given model represents the same user
func (m Model) Is(other Model) bool {
	return m.id == other.id
}
