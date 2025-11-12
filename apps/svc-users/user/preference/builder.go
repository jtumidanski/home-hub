package preference

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Builder provides a fluent API for constructing preferences
type Builder struct {
	id        uuid.UUID
	userId    uuid.UUID
	key       string
	value     string
	createdAt time.Time
	updatedAt time.Time
}

// NewBuilder creates a new preference builder
func NewBuilder() *Builder {
	return &Builder{
		id: uuid.New(),
	}
}

// WithId sets the preference ID (typically used when reconstructing from DB)
func (b *Builder) WithId(id uuid.UUID) *Builder {
	b.id = id
	return b
}

// ForUser sets the user ID this preference belongs to
func (b *Builder) ForUser(userId uuid.UUID) *Builder {
	b.userId = userId
	return b
}

// WithKey sets the preference key
func (b *Builder) WithKey(key string) *Builder {
	b.key = key
	return b
}

// WithValue sets the preference value
func (b *Builder) WithValue(value string) *Builder {
	b.value = value
	return b
}

// WithCreatedAt sets the creation timestamp (typically used when reconstructing from DB)
func (b *Builder) WithCreatedAt(createdAt time.Time) *Builder {
	b.createdAt = createdAt
	return b
}

// WithUpdatedAt sets the update timestamp (typically used when reconstructing from DB)
func (b *Builder) WithUpdatedAt(updatedAt time.Time) *Builder {
	b.updatedAt = updatedAt
	return b
}

// Build constructs the preference model after validation
func (b *Builder) Build() (Model, error) {
	// Set timestamps if not provided
	now := time.Now()
	if b.createdAt.IsZero() {
		b.createdAt = now
	}
	if b.updatedAt.IsZero() {
		b.updatedAt = now
	}

	// Create the model
	model := Model{
		id:        b.id,
		userId:    b.userId,
		key:       b.key,
		value:     b.value,
		createdAt: b.createdAt,
		updatedAt: b.updatedAt,
	}

	// Validate
	if err := b.validate(model); err != nil {
		return Model{}, err
	}

	return model, nil
}

// validate checks that the preference is valid before construction
func (b *Builder) validate(model Model) error {
	// Check user ID
	if model.userId == uuid.Nil {
		return fmt.Errorf("userId is required")
	}

	// Check key
	if !model.ValidKey() {
		return fmt.Errorf("invalid key: must be non-empty and <= 100 characters")
	}

	// Check value
	if !model.ValidValue() {
		return fmt.Errorf("invalid value: must be non-empty and <= 10000 characters")
	}

	// Perform key-specific validation
	if err := model.ValidateForKey(); err != nil {
		return err
	}

	return nil
}

// Convenience constructors for common preference types

// NewThemePreference creates a new theme preference
func NewThemePreference(userId uuid.UUID, theme string) (Model, error) {
	return NewBuilder().
		ForUser(userId).
		WithKey(THEME_KEY).
		WithValue(theme).
		Build()
}

// UpdateValue creates a new preference with an updated value
// This is useful for updating an existing preference while maintaining immutability
func UpdateValue(existing Model, newValue string) (Model, error) {
	return NewBuilder().
		WithId(existing.Id()).
		ForUser(existing.UserId()).
		WithKey(existing.Key()).
		WithValue(newValue).
		WithCreatedAt(existing.CreatedAt()).
		WithUpdatedAt(time.Now()).
		Build()
}
