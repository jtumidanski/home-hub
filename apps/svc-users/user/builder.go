package user

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEmailRequired       = errors.New("email is required")
	ErrEmailInvalid        = errors.New("email format is invalid")
	ErrDisplayNameRequired = errors.New("display name is required")
	ErrDisplayNameEmpty    = errors.New("display name cannot be empty")
)

// Builder provides a fluent API for constructing valid User models
type Builder struct {
	id          *uuid.UUID
	email       *string
	displayName *string
	householdId *uuid.UUID
	createdAt   *time.Time
	updatedAt   *time.Time
}

// NewBuilder creates a new user builder
func NewBuilder() *Builder {
	return &Builder{}
}

// SetId sets the user ID
func (b *Builder) SetId(id uuid.UUID) *Builder {
	b.id = &id
	return b
}

// SetEmail sets the user email
func (b *Builder) SetEmail(email string) *Builder {
	b.email = &email
	return b
}

// SetDisplayName sets the user display name
func (b *Builder) SetDisplayName(displayName string) *Builder {
	b.displayName = &displayName
	return b
}

// SetHouseholdId sets the user's household ID
func (b *Builder) SetHouseholdId(householdId uuid.UUID) *Builder {
	b.householdId = &householdId
	return b
}

// ClearHouseholdId removes the household association
func (b *Builder) ClearHouseholdId() *Builder {
	b.householdId = nil
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

// Build validates the builder state and constructs a User Model
func (b *Builder) Build() (Model, error) {
	// Validate email
	if b.email == nil {
		return Model{}, ErrEmailRequired
	}
	email := strings.TrimSpace(*b.email)
	if email == "" {
		return Model{}, ErrEmailRequired
	}
	if !emailRegex.MatchString(email) {
		return Model{}, ErrEmailInvalid
	}

	// Validate display name
	if b.displayName == nil {
		return Model{}, ErrDisplayNameRequired
	}
	displayName := strings.TrimSpace(*b.displayName)
	if displayName == "" {
		return Model{}, ErrDisplayNameEmpty
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
		email:       email,
		displayName: displayName,
		householdId: b.householdId,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}, nil
}

// Builder creates a builder initialized with the model's current values
// This enables modification flows: model.Builder().SetEmail(newEmail).Build()
func (m Model) Builder() *Builder {
	b := &Builder{
		id:          &m.id,
		email:       &m.email,
		displayName: &m.displayName,
		createdAt:   &m.createdAt,
		updatedAt:   &m.updatedAt,
	}

	if m.householdId != nil {
		householdIdCopy := *m.householdId
		b.householdId = &householdIdCopy
	}

	return b
}
