package refreshtoken

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrUserIDRequired    = errors.New("refresh token user ID is required")
	ErrTokenHashRequired = errors.New("refresh token hash is required")
	ErrExpiresAtRequired = errors.New("refresh token expiration is required")
)

type Builder struct {
	id        uuid.UUID
	userId    uuid.UUID
	tokenHash string
	expiresAt time.Time
	revoked   bool
	createdAt time.Time
	updatedAt time.Time
}

func NewBuilder() *Builder {
	return &Builder{}
}

func (b *Builder) SetId(id uuid.UUID) *Builder           { b.id = id; return b }
func (b *Builder) SetUserId(id uuid.UUID) *Builder       { b.userId = id; return b }
func (b *Builder) SetTokenHash(hash string) *Builder     { b.tokenHash = hash; return b }
func (b *Builder) SetExpiresAt(t time.Time) *Builder     { b.expiresAt = t; return b }
func (b *Builder) SetRevoked(revoked bool) *Builder      { b.revoked = revoked; return b }
func (b *Builder) SetCreatedAt(t time.Time) *Builder     { b.createdAt = t; return b }
func (b *Builder) SetUpdatedAt(t time.Time) *Builder     { b.updatedAt = t; return b }

func (b *Builder) Build() (Model, error) {
	if b.userId == uuid.Nil {
		return Model{}, ErrUserIDRequired
	}
	if b.tokenHash == "" {
		return Model{}, ErrTokenHashRequired
	}
	if b.expiresAt.IsZero() {
		return Model{}, ErrExpiresAtRequired
	}
	return Model{
		id:        b.id,
		userId:    b.userId,
		tokenHash: b.tokenHash,
		expiresAt: b.expiresAt,
		revoked:   b.revoked,
		createdAt: b.createdAt,
		updatedAt: b.updatedAt,
	}, nil
}
