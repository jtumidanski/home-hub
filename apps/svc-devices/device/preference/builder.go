package preference

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrDeviceIdRequired      = errors.New("device ID is required")
	ErrInvalidTheme          = errors.New("theme must be 'light' or 'dark'")
	ErrInvalidTemperatureUnit = errors.New("temperature unit must be 'household', 'F', or 'C'")
)

// Builder provides a fluent API for constructing valid DevicePreferences models
type Builder struct {
	id              *uuid.UUID
	deviceId        *uuid.UUID
	theme           *string
	temperatureUnit *string
	createdAt       *time.Time
	updatedAt       *time.Time
}

// NewBuilder creates a new device preferences builder with defaults
func NewBuilder() *Builder {
	defaultTheme := ThemeDark
	defaultTempUnit := TempUnitHousehold
	return &Builder{
		theme:           &defaultTheme,
		temperatureUnit: &defaultTempUnit,
	}
}

// SetId sets the preference ID
func (b *Builder) SetId(id uuid.UUID) *Builder {
	b.id = &id
	return b
}

// SetDeviceId sets the device ID
func (b *Builder) SetDeviceId(deviceId uuid.UUID) *Builder {
	b.deviceId = &deviceId
	return b
}

// SetTheme sets the theme preference
func (b *Builder) SetTheme(theme string) *Builder {
	b.theme = &theme
	return b
}

// SetTemperatureUnit sets the temperature unit preference
func (b *Builder) SetTemperatureUnit(unit string) *Builder {
	b.temperatureUnit = &unit
	return b
}

// SetCreatedAt sets the creation timestamp
func (b *Builder) SetCreatedAt(createdAt time.Time) *Builder {
	b.createdAt = &createdAt
	return b
}

// SetUpdatedAt sets the update timestamp
func (b *Builder) SetUpdatedAt(updatedAt time.Time) *Builder {
	b.updatedAt = &updatedAt
	return b
}

// Build validates the builder state and constructs a DevicePreferences Model
func (b *Builder) Build() (Model, error) {
	// Validate device ID
	if b.deviceId == nil {
		return Model{}, ErrDeviceIdRequired
	}

	// Validate theme
	if b.theme != nil {
		theme := *b.theme
		if theme != ThemeLight && theme != ThemeDark {
			return Model{}, ErrInvalidTheme
		}
	}

	// Validate temperature unit
	if b.temperatureUnit != nil {
		unit := *b.temperatureUnit
		if unit != TempUnitHousehold && unit != TempUnitFahrenheit && unit != TempUnitCelsius {
			return Model{}, ErrInvalidTemperatureUnit
		}
	}

	// Generate ID if not provided
	id := uuid.New()
	if b.id != nil {
		id = *b.id
	}

	// Generate timestamps if not provided
	now := time.Now()
	createdAt := now
	if b.createdAt != nil {
		createdAt = *b.createdAt
	}
	updatedAt := now
	if b.updatedAt != nil {
		updatedAt = *b.updatedAt
	}

	// Use defaults if not set
	theme := ThemeDark
	if b.theme != nil {
		theme = *b.theme
	}
	temperatureUnit := TempUnitHousehold
	if b.temperatureUnit != nil {
		temperatureUnit = *b.temperatureUnit
	}

	return Model{
		id:              id,
		deviceId:        *b.deviceId,
		theme:           theme,
		temperatureUnit: temperatureUnit,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
	}, nil
}

// Builder creates a builder initialized with the model's current values
// This enables modification flows: model.Builder().SetTheme(newTheme).Build()
func (m Model) Builder() *Builder {
	return &Builder{
		id:              &m.id,
		deviceId:        &m.deviceId,
		theme:           &m.theme,
		temperatureUnit: &m.temperatureUnit,
		createdAt:       &m.createdAt,
		updatedAt:       &m.updatedAt,
	}
}
