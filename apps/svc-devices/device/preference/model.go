package preference

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Valid theme values
const (
	ThemeLight = "light"
	ThemeDark  = "dark"
)

// Valid temperature unit values
const (
	TempUnitHousehold  = "household" // Defer to household preference
	TempUnitFahrenheit = "F"
	TempUnitCelsius    = "C"
)

// Model represents immutable device preferences in the domain.
// All fields are private to enforce immutability.
type Model struct {
	id              uuid.UUID
	deviceId        uuid.UUID
	theme           string
	temperatureUnit string
	createdAt       time.Time
	updatedAt       time.Time
}

// Id returns the preference's unique identifier
func (m Model) Id() uuid.UUID {
	return m.id
}

// DeviceId returns the ID of the device these preferences belong to
func (m Model) DeviceId() uuid.UUID {
	return m.deviceId
}

// Theme returns the device's theme preference
func (m Model) Theme() string {
	return m.theme
}

// TemperatureUnit returns the device's temperature unit preference
func (m Model) TemperatureUnit() string {
	return m.temperatureUnit
}

// CreatedAt returns when the preferences were created
func (m Model) CreatedAt() time.Time {
	return m.createdAt
}

// UpdatedAt returns when the preferences were last updated
func (m Model) UpdatedAt() time.Time {
	return m.updatedAt
}

// String returns a string representation of the preferences for debugging
func (m Model) String() string {
	return fmt.Sprintf("DevicePreferences[deviceId=%s, theme=%s, tempUnit=%s]",
		m.deviceId.String(), m.theme, m.temperatureUnit)
}

// MarshalJSON implements json.Marshaler for the Model
func (m Model) MarshalJSON() ([]byte, error) {
	type alias struct {
		Id              uuid.UUID `json:"id"`
		DeviceId        uuid.UUID `json:"deviceId"`
		Theme           string    `json:"theme"`
		TemperatureUnit string    `json:"temperatureUnit"`
		CreatedAt       time.Time `json:"createdAt"`
		UpdatedAt       time.Time `json:"updatedAt"`
	}

	return json.Marshal(&alias{
		Id:              m.id,
		DeviceId:        m.deviceId,
		Theme:           m.theme,
		TemperatureUnit: m.temperatureUnit,
		CreatedAt:       m.createdAt,
		UpdatedAt:       m.updatedAt,
	})
}

// UnmarshalJSON implements json.Unmarshaler for the Model
func (m *Model) UnmarshalJSON(data []byte) error {
	type alias struct {
		Id              uuid.UUID `json:"id"`
		DeviceId        uuid.UUID `json:"deviceId"`
		Theme           string    `json:"theme"`
		TemperatureUnit string    `json:"temperatureUnit"`
		CreatedAt       time.Time `json:"createdAt"`
		UpdatedAt       time.Time `json:"updatedAt"`
	}

	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}

	m.id = a.Id
	m.deviceId = a.DeviceId
	m.theme = a.Theme
	m.temperatureUnit = a.TemperatureUnit
	m.createdAt = a.CreatedAt
	m.updatedAt = a.UpdatedAt

	return nil
}

// Is returns true if the given model represents the same preferences
func (m Model) Is(other Model) bool {
	return m.id == other.id
}

// IsForDevice returns true if these preferences belong to the given device
func (m Model) IsForDevice(deviceId uuid.UUID) bool {
	return m.deviceId == deviceId
}
