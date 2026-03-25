package externalidentity

import (
	"time"

	"github.com/google/uuid"
)

// Model is the immutable domain model for an external identity.
type Model struct {
	id              uuid.UUID
	userId          uuid.UUID
	provider        string
	providerSubject string
	createdAt       time.Time
	updatedAt       time.Time
}

func (m Model) Id() uuid.UUID          { return m.id }
func (m Model) UserId() uuid.UUID      { return m.userId }
func (m Model) Provider() string       { return m.provider }
func (m Model) ProviderSubject() string { return m.providerSubject }
func (m Model) CreatedAt() time.Time   { return m.createdAt }
func (m Model) UpdatedAt() time.Time   { return m.updatedAt }

// ToEntity converts the domain model back to a database entity.
func (m Model) ToEntity() Entity {
	return Entity{
		Id:              m.id,
		UserId:          m.userId,
		Provider:        m.provider,
		ProviderSubject: m.providerSubject,
		CreatedAt:       m.createdAt,
		UpdatedAt:       m.updatedAt,
	}
}
