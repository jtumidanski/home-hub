package device

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNameRequired     = errors.New("device name is required")
	ErrNameEmpty        = errors.New("device name cannot be empty")
	ErrTypeRequired     = errors.New("device type is required")
	ErrTypeInvalid      = errors.New("device type must be 'kiosk'")
	ErrHouseholdRequired = errors.New("household ID is required")
)

// Builder provides a fluent API for constructing valid Device models
type Builder struct {
	id          *uuid.UUID
	name        *string
	deviceType  *string
	householdId *uuid.UUID
	createdAt   *time.Time
	updatedAt   *time.Time
}

// NewBuilder creates a new device builder
func NewBuilder() *Builder {
	return &Builder{}
}

// SetId sets the device ID
func (b *Builder) SetId(id uuid.UUID) *Builder {
	b.id = &id
	return b
}

// SetName sets the device name
func (b *Builder) SetName(name string) *Builder {
	b.name = &name
	return b
}

// SetType sets the device type
func (b *Builder) SetType(deviceType string) *Builder {
	b.deviceType = &deviceType
	return b
}

// SetHouseholdId sets the household ID
func (b *Builder) SetHouseholdId(householdId uuid.UUID) *Builder {
	b.householdId = &householdId
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

// Build validates the builder state and constructs a Device Model
func (b *Builder) Build() (Model, error) {
	// Validate name
	if b.name == nil {
		return Model{}, ErrNameRequired
	}
	name := strings.TrimSpace(*b.name)
	if name == "" {
		return Model{}, ErrNameEmpty
	}

	// Validate type
	if b.deviceType == nil {
		return Model{}, ErrTypeRequired
	}
	deviceType := *b.deviceType
	if deviceType != "kiosk" {
		return Model{}, ErrTypeInvalid
	}

	// Validate household ID
	if b.householdId == nil {
		return Model{}, ErrHouseholdRequired
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

	return Model{
		id:          id,
		name:        name,
		deviceType:  deviceType,
		householdId: *b.householdId,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}, nil
}

// Builder creates a builder initialized with the model's current values
// This enables modification flows: model.Builder().SetName(newName).Build()
func (m Model) Builder() *Builder {
	return &Builder{
		id:          &m.id,
		name:        &m.name,
		deviceType:  &m.deviceType,
		householdId: &m.householdId,
		createdAt:   &m.createdAt,
		updatedAt:   &m.updatedAt,
	}
}
