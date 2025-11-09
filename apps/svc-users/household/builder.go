package household

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNameRequired = errors.New("household name is required")
	ErrNameEmpty    = errors.New("household name cannot be empty")
)

// Builder provides a fluent API for constructing valid Household models
type Builder struct {
	id        *uuid.UUID
	name      *string
	createdAt *time.Time
	updatedAt *time.Time
}

// NewBuilder creates a new household builder
func NewBuilder() *Builder {
	return &Builder{}
}

// SetId sets the household ID
func (b *Builder) SetId(id uuid.UUID) *Builder {
	b.id = &id
	return b
}

// SetName sets the household name
func (b *Builder) SetName(name string) *Builder {
	b.name = &name
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

// Build validates the builder state and constructs a Household Model
func (b *Builder) Build() (Model, error) {
	// Validate name
	if b.name == nil {
		return Model{}, ErrNameRequired
	}
	name := strings.TrimSpace(*b.name)
	if name == "" {
		return Model{}, ErrNameEmpty
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
		id:        id,
		name:      name,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}, nil
}

// Builder creates a builder initialized with the model's current values
// This enables modification flows: model.Builder().SetName(newName).Build()
func (m Model) Builder() *Builder {
	return &Builder{
		id:        &m.id,
		name:      &m.name,
		createdAt: &m.createdAt,
		updatedAt: &m.updatedAt,
	}
}
