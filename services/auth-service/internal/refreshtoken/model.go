package refreshtoken

import (
	"time"

	"github.com/google/uuid"
)

// Model is the immutable domain model for a refresh token.
type Model struct {
	id        uuid.UUID
	userId    uuid.UUID
	tokenHash string
	expiresAt time.Time
	revoked   bool
	createdAt time.Time
	updatedAt time.Time
}

func (m Model) Id() uuid.UUID        { return m.id }
func (m Model) UserId() uuid.UUID    { return m.userId }
func (m Model) TokenHash() string    { return m.tokenHash }
func (m Model) ExpiresAt() time.Time { return m.expiresAt }
func (m Model) Revoked() bool        { return m.revoked }
func (m Model) CreatedAt() time.Time { return m.createdAt }
func (m Model) UpdatedAt() time.Time { return m.updatedAt }

// ToEntity converts the domain model back to a database entity.
func (m Model) ToEntity() Entity {
	return Entity{
		Id:        m.id,
		UserId:    m.userId,
		TokenHash: m.tokenHash,
		ExpiresAt: m.expiresAt,
		Revoked:   m.revoked,
		CreatedAt: m.createdAt,
		UpdatedAt: m.updatedAt,
	}
}
