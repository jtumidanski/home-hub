package role

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// Standard role constants
const (
	Admin          = "admin"           // Full system access
	User           = "user"            // Standard authenticated user
	HouseholdAdmin = "household_admin" // Manage household settings
	DeviceManager  = "device_manager"  // Manage devices
)

// ValidRoles is the set of valid role names
var ValidRoles = map[string]bool{
	Admin:          true,
	User:           true,
	HouseholdAdmin: true,
	DeviceManager:  true,
}

// Model represents an immutable user role assignment in the domain.
// All fields are private to enforce immutability.
type Model struct {
	userId uuid.UUID
	role   string
}

// UserId returns the user's unique identifier
func (m Model) UserId() uuid.UUID {
	return m.userId
}

// Role returns the role name
func (m Model) Role() string {
	return m.role
}

// IsAdmin returns true if this is an admin role
func (m Model) IsAdmin() bool {
	return m.role == Admin
}

// IsValid returns true if the role is a recognized valid role
func (m Model) IsValid() bool {
	return ValidRoles[m.role]
}

// String returns a string representation of the user role for debugging
func (m Model) String() string {
	return fmt.Sprintf("UserRole[userId=%s, role=%s]",
		m.userId.String(), m.role)
}

// MarshalJSON implements json.Marshaler for the Model
func (m Model) MarshalJSON() ([]byte, error) {
	type alias struct {
		UserId uuid.UUID `json:"userId"`
		Role   string    `json:"role"`
	}

	return json.Marshal(&alias{
		UserId: m.userId,
		Role:   m.role,
	})
}

// UnmarshalJSON implements json.Unmarshaler for the Model
func (m *Model) UnmarshalJSON(data []byte) error {
	type alias struct {
		UserId uuid.UUID `json:"userId"`
		Role   string    `json:"role"`
	}

	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}

	m.userId = a.UserId
	m.role = a.Role

	return nil
}

// Is returns true if the given model represents the same user role
func (m Model) Is(other Model) bool {
	return m.userId == other.userId && m.role == other.role
}
