package preference

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	// Common preference keys
	THEME_KEY = "theme"
)

// Valid theme values
const (
	THEME_SYSTEM = "system"
	THEME_LIGHT  = "light"
	THEME_DARK   = "dark"
)

// Model represents an immutable user preference in the domain.
// All fields are private to enforce immutability.
type Model struct {
	id        uuid.UUID
	userId    uuid.UUID
	key       string
	value     string
	createdAt time.Time
	updatedAt time.Time
}

// Id returns the preference's unique identifier
func (m Model) Id() uuid.UUID {
	return m.id
}

// UserId returns the user ID this preference belongs to
func (m Model) UserId() uuid.UUID {
	return m.userId
}

// Key returns the preference key
func (m Model) Key() string {
	return m.key
}

// Value returns the preference value
func (m Model) Value() string {
	return m.value
}

// CreatedAt returns when the preference was created
func (m Model) CreatedAt() time.Time {
	return m.createdAt
}

// UpdatedAt returns when the preference was last updated
func (m Model) UpdatedAt() time.Time {
	return m.updatedAt
}

// ValidKey returns true if the key is valid (non-empty, reasonable length)
func (m Model) ValidKey() bool {
	return len(m.key) > 0 && len(m.key) <= 100
}

// ValidValue returns true if the value is valid (non-empty, reasonable length)
func (m Model) ValidValue() bool {
	// Basic validation - value should not be empty
	// More specific validation can be done based on key type
	return len(m.value) > 0 && len(m.value) <= 10000 // 10KB max
}

// ValidateForKey performs key-specific validation
func (m Model) ValidateForKey() error {
	switch m.key {
	case THEME_KEY:
		return m.validateThemeValue()
	default:
		// For unknown keys, just check basic value validity
		if !m.ValidValue() {
			return fmt.Errorf("invalid value for key %s", m.key)
		}
		return nil
	}
}

// validateThemeValue validates that the theme value is one of the allowed values
func (m Model) validateThemeValue() error {
	value := strings.ToLower(strings.TrimSpace(m.value))
	switch value {
	case THEME_SYSTEM, THEME_LIGHT, THEME_DARK:
		return nil
	default:
		return fmt.Errorf("invalid theme value: %s (must be 'system', 'light', or 'dark')", m.value)
	}
}

// String returns a string representation of the preference for debugging
func (m Model) String() string {
	return fmt.Sprintf("Preference[id=%s, userId=%s, key=%s, value=%s]",
		m.id.String(), m.userId.String(), m.key, m.value)
}

// MarshalJSON implements json.Marshaler for the Model
func (m Model) MarshalJSON() ([]byte, error) {
	type alias struct {
		Id        uuid.UUID `json:"id"`
		UserId    uuid.UUID `json:"userId"`
		Key       string    `json:"key"`
		Value     string    `json:"value"`
		CreatedAt time.Time `json:"createdAt"`
		UpdatedAt time.Time `json:"updatedAt"`
	}

	return json.Marshal(&alias{
		Id:        m.id,
		UserId:    m.userId,
		Key:       m.key,
		Value:     m.value,
		CreatedAt: m.createdAt,
		UpdatedAt: m.updatedAt,
	})
}

// UnmarshalJSON implements json.Unmarshaler for the Model
func (m *Model) UnmarshalJSON(data []byte) error {
	type alias struct {
		Id        uuid.UUID `json:"id"`
		UserId    uuid.UUID `json:"userId"`
		Key       string    `json:"key"`
		Value     string    `json:"value"`
		CreatedAt time.Time `json:"createdAt"`
		UpdatedAt time.Time `json:"updatedAt"`
	}

	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}

	m.id = a.Id
	m.userId = a.UserId
	m.key = a.Key
	m.value = a.Value
	m.createdAt = a.CreatedAt
	m.updatedAt = a.UpdatedAt

	return nil
}

// Is returns true if the given model represents the same preference
func (m Model) Is(other Model) bool {
	return m.id == other.id
}

// IsForUser returns true if this preference belongs to the given user
func (m Model) IsForUser(userId uuid.UUID) bool {
	return m.userId == userId
}

// HasKey returns true if this preference has the given key
func (m Model) HasKey(key string) bool {
	return m.key == key
}
