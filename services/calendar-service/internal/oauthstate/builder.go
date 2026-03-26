package oauthstate

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrRedirectURIRequired = errors.New("redirect URI is required")
)

type Builder struct {
	id          uuid.UUID
	tenantID    uuid.UUID
	householdID uuid.UUID
	userID      uuid.UUID
	redirectURI string
	expiresAt   time.Time
	createdAt   time.Time
}

func NewBuilder() *Builder { return &Builder{} }

func (b *Builder) SetId(id uuid.UUID) *Builder          { b.id = id; return b }
func (b *Builder) SetTenantID(id uuid.UUID) *Builder     { b.tenantID = id; return b }
func (b *Builder) SetHouseholdID(id uuid.UUID) *Builder  { b.householdID = id; return b }
func (b *Builder) SetUserID(id uuid.UUID) *Builder       { b.userID = id; return b }
func (b *Builder) SetRedirectURI(uri string) *Builder    { b.redirectURI = uri; return b }
func (b *Builder) SetExpiresAt(t time.Time) *Builder     { b.expiresAt = t; return b }
func (b *Builder) SetCreatedAt(t time.Time) *Builder     { b.createdAt = t; return b }

func (b *Builder) Build() (Model, error) {
	if b.redirectURI == "" {
		return Model{}, ErrRedirectURIRequired
	}
	return Model{
		id:          b.id,
		tenantID:    b.tenantID,
		householdID: b.householdID,
		userID:      b.userID,
		redirectURI: b.redirectURI,
		expiresAt:   b.expiresAt,
		createdAt:   b.createdAt,
	}, nil
}
