package role

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrUserIdRequired = errors.New("user ID is required")
	ErrRoleRequired   = errors.New("role is required")
	ErrRoleEmpty      = errors.New("role cannot be empty")
	ErrRoleInvalid    = errors.New("role is not a valid role name")
)

// Builder provides a fluent API for constructing valid UserRole models
type Builder struct {
	userId *uuid.UUID
	role   *string
}

// NewBuilder creates a new user role builder
func NewBuilder() *Builder {
	return &Builder{}
}

// SetUserId sets the user ID
func (b *Builder) SetUserId(userId uuid.UUID) *Builder {
	b.userId = &userId
	return b
}

// SetRole sets the role name
func (b *Builder) SetRole(role string) *Builder {
	b.role = &role
	return b
}

// Build validates the builder state and constructs a UserRole Model
func (b *Builder) Build() (Model, error) {
	// Validate user ID
	if b.userId == nil {
		return Model{}, ErrUserIdRequired
	}

	// Validate role
	if b.role == nil {
		return Model{}, ErrRoleRequired
	}
	role := strings.TrimSpace(*b.role)
	if role == "" {
		return Model{}, ErrRoleEmpty
	}
	if !ValidRoles[role] {
		return Model{}, ErrRoleInvalid
	}

	return Model{
		userId: *b.userId,
		role:   role,
	}, nil
}

// Builder creates a builder initialized with the model's current values
func (m Model) Builder() *Builder {
	return &Builder{
		userId: &m.userId,
		role:   &m.role,
	}
}
