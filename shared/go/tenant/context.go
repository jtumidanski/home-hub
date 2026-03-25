// Package tenant provides context-based tenant extraction and propagation.
package tenant

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type contextKey string

const tenantKey contextKey = "tenant"

// Tenant holds the resolved tenant and household context for a request.
type Tenant struct {
	id          uuid.UUID
	householdID uuid.UUID
	userID      uuid.UUID
}

// New creates a new Tenant.
func New(id, householdID, userID uuid.UUID) Tenant {
	return Tenant{id: id, householdID: householdID, userID: userID}
}

// Id returns the tenant ID.
func (t Tenant) Id() uuid.UUID { return t.id }

// HouseholdId returns the household ID.
func (t Tenant) HouseholdId() uuid.UUID { return t.householdID }

// UserId returns the user ID.
func (t Tenant) UserId() uuid.UUID { return t.userID }

// WithContext returns a new context with the given tenant.
func WithContext(ctx context.Context, t Tenant) context.Context {
	return context.WithValue(ctx, tenantKey, t)
}

// FromContext extracts the tenant from the context.
// Returns the tenant and true if found, or a zero Tenant and false if not.
func FromContext(ctx context.Context) (Tenant, bool) {
	t, ok := ctx.Value(tenantKey).(Tenant)
	return t, ok
}

// MustFromContext extracts the tenant from the context or panics.
func MustFromContext(ctx context.Context) Tenant {
	t, ok := FromContext(ctx)
	if !ok {
		panic(errors.New("tenant not found in context"))
	}
	return t
}
