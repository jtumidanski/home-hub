package connection

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrProviderRequired        = errors.New("provider is required")
	ErrEmailRequired           = errors.New("email is required")
	ErrAccessTokenRequired     = errors.New("access token is required")
	ErrRefreshTokenRequired    = errors.New("refresh token is required")
	ErrUserDisplayNameRequired = errors.New("user display name is required")
)

type Builder struct {
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
	writeAccess        bool
	lastSyncAt          *time.Time
	lastSyncAttemptAt   *time.Time
	lastSyncEventCount  int
	errorCode           *string
	errorMessage        *string
	lastErrorAt         *time.Time
	consecutiveFailures int
	createdAt           time.Time
	updatedAt           time.Time
}

func NewBuilder() *Builder { return &Builder{} }

func (b *Builder) SetId(id uuid.UUID) *Builder              { b.id = id; return b }
func (b *Builder) SetTenantID(id uuid.UUID) *Builder         { b.tenantID = id; return b }
func (b *Builder) SetHouseholdID(id uuid.UUID) *Builder      { b.householdID = id; return b }
func (b *Builder) SetUserID(id uuid.UUID) *Builder           { b.userID = id; return b }
func (b *Builder) SetProvider(p string) *Builder             { b.provider = p; return b }
func (b *Builder) SetStatus(s string) *Builder               { b.status = s; return b }
func (b *Builder) SetEmail(e string) *Builder                { b.email = e; return b }
func (b *Builder) SetAccessToken(t string) *Builder          { b.accessToken = t; return b }
func (b *Builder) SetRefreshToken(t string) *Builder         { b.refreshToken = t; return b }
func (b *Builder) SetTokenExpiry(t time.Time) *Builder       { b.tokenExpiry = t; return b }
func (b *Builder) SetUserDisplayName(n string) *Builder      { b.userDisplayName = n; return b }
func (b *Builder) SetUserColor(c string) *Builder            { b.userColor = c; return b }
func (b *Builder) SetWriteAccess(w bool) *Builder            { b.writeAccess = w; return b }
func (b *Builder) SetLastSyncAt(t *time.Time) *Builder       { b.lastSyncAt = t; return b }
func (b *Builder) SetLastSyncAttemptAt(t *time.Time) *Builder { b.lastSyncAttemptAt = t; return b }
func (b *Builder) SetLastSyncEventCount(c int) *Builder      { b.lastSyncEventCount = c; return b }
func (b *Builder) SetErrorCode(c *string) *Builder           { b.errorCode = c; return b }
func (b *Builder) SetErrorMessage(m *string) *Builder        { b.errorMessage = m; return b }
func (b *Builder) SetLastErrorAt(t *time.Time) *Builder      { b.lastErrorAt = t; return b }
func (b *Builder) SetConsecutiveFailures(n int) *Builder     { b.consecutiveFailures = n; return b }
func (b *Builder) SetCreatedAt(t time.Time) *Builder         { b.createdAt = t; return b }
func (b *Builder) SetUpdatedAt(t time.Time) *Builder         { b.updatedAt = t; return b }

func (b *Builder) Build() (Model, error) {
	if b.provider == "" {
		return Model{}, ErrProviderRequired
	}
	if b.email == "" {
		return Model{}, ErrEmailRequired
	}
	if b.accessToken == "" {
		return Model{}, ErrAccessTokenRequired
	}
	if b.refreshToken == "" {
		return Model{}, ErrRefreshTokenRequired
	}
	if b.userDisplayName == "" {
		return Model{}, ErrUserDisplayNameRequired
	}
	return Model{
		id:                 b.id,
		tenantID:           b.tenantID,
		householdID:        b.householdID,
		userID:             b.userID,
		provider:           b.provider,
		status:             b.status,
		email:              b.email,
		accessToken:        b.accessToken,
		refreshToken:       b.refreshToken,
		tokenExpiry:        b.tokenExpiry,
		userDisplayName:    b.userDisplayName,
		userColor:          b.userColor,
		writeAccess:         b.writeAccess,
		lastSyncAt:          b.lastSyncAt,
		lastSyncAttemptAt:   b.lastSyncAttemptAt,
		lastSyncEventCount:  b.lastSyncEventCount,
		errorCode:           b.errorCode,
		errorMessage:        b.errorMessage,
		lastErrorAt:         b.lastErrorAt,
		consecutiveFailures: b.consecutiveFailures,
		createdAt:           b.createdAt,
		updatedAt:           b.updatedAt,
	}, nil
}
