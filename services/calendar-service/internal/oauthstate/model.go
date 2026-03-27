package oauthstate

import (
	"time"

	"github.com/google/uuid"
)

type Model struct {
	id          uuid.UUID
	tenantID    uuid.UUID
	householdID uuid.UUID
	userID      uuid.UUID
	redirectURI string
	reauthorize bool
	expiresAt   time.Time
	createdAt   time.Time
}

func (m Model) Id() uuid.UUID          { return m.id }
func (m Model) TenantID() uuid.UUID    { return m.tenantID }
func (m Model) HouseholdID() uuid.UUID { return m.householdID }
func (m Model) UserID() uuid.UUID      { return m.userID }
func (m Model) RedirectURI() string    { return m.redirectURI }
func (m Model) Reauthorize() bool     { return m.reauthorize }
func (m Model) ExpiresAt() time.Time   { return m.expiresAt }
func (m Model) CreatedAt() time.Time   { return m.createdAt }

func (m Model) IsExpired() bool {
	return time.Now().UTC().After(m.expiresAt)
}
