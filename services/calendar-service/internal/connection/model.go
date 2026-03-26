package connection

import (
	"time"

	"github.com/google/uuid"
)

var UserColors = []string{
	"#4285F4", "#EA4335", "#34A853", "#FBBC04",
	"#8E24AA", "#00ACC1", "#FF7043", "#78909C",
}

type Model struct {
	id                 uuid.UUID
	tenantID           uuid.UUID
	householdID        uuid.UUID
	userID             uuid.UUID
	provider           string
	status             string
	email              string
	accessToken        string
	refreshToken       string
	tokenExpiry        time.Time
	userDisplayName    string
	userColor          string
	lastSyncAt         *time.Time
	lastSyncEventCount int
	createdAt          time.Time
	updatedAt          time.Time
}

func (m Model) Id() uuid.UUID             { return m.id }
func (m Model) TenantID() uuid.UUID       { return m.tenantID }
func (m Model) HouseholdID() uuid.UUID    { return m.householdID }
func (m Model) UserID() uuid.UUID         { return m.userID }
func (m Model) Provider() string          { return m.provider }
func (m Model) Status() string            { return m.status }
func (m Model) Email() string             { return m.email }
func (m Model) AccessToken() string       { return m.accessToken }
func (m Model) RefreshToken() string      { return m.refreshToken }
func (m Model) TokenExpiry() time.Time    { return m.tokenExpiry }
func (m Model) UserDisplayName() string   { return m.userDisplayName }
func (m Model) UserColor() string         { return m.userColor }
func (m Model) LastSyncAt() *time.Time    { return m.lastSyncAt }
func (m Model) LastSyncEventCount() int   { return m.lastSyncEventCount }
func (m Model) CreatedAt() time.Time      { return m.createdAt }
func (m Model) UpdatedAt() time.Time      { return m.updatedAt }

func (m Model) IsTokenExpired() bool {
	return time.Now().UTC().After(m.tokenExpiry)
}
