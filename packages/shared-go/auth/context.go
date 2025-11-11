package auth

import (
	"context"

	"github.com/google/uuid"
)

// contextKey is a private type for context keys to avoid collisions
type contextKey struct{}

// authContextKey is the key used to store AuthContext in context.Context
var authContextKey = contextKey{}

// Context represents the authenticated user's identity and roles
type Context struct {
	UserId   uuid.UUID
	Email    string
	Name     string
	Provider string
	Roles    []string
}

// WithContext attaches an AuthContext to a context.Context
func WithContext(ctx context.Context, auth Context) context.Context {
	return context.WithValue(ctx, authContextKey, auth)
}

// FromContext extracts the AuthContext from context.Context
// Returns the context and a boolean indicating if it was found
func FromContext(ctx context.Context) (Context, bool) {
	auth, ok := ctx.Value(authContextKey).(Context)
	return auth, ok
}

// MustFromContext extracts the AuthContext from context.Context
// Panics if the context is not found (should only be used in authenticated routes)
func MustFromContext(ctx context.Context) Context {
	auth, ok := FromContext(ctx)
	if !ok {
		panic("auth context not found - middleware not applied?")
	}
	return auth
}

// HasRole checks if the authenticated user has a specific role
func (c Context) HasRole(role string) bool {
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole checks if the authenticated user has any of the specified roles
func (c Context) HasAnyRole(roles []string) bool {
	for _, role := range roles {
		if c.HasRole(role) {
			return true
		}
	}
	return false
}

// HasAllRoles checks if the authenticated user has all of the specified roles
func (c Context) HasAllRoles(roles []string) bool {
	for _, role := range roles {
		if !c.HasRole(role) {
			return false
		}
	}
	return true
}

// IsAdmin checks if the authenticated user has the admin role
func (c Context) IsAdmin() bool {
	return c.HasRole("admin")
}
