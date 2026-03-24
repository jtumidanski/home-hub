package user

import (
	"time"

	"github.com/google/uuid"
)

// Model is the immutable domain model for a user.
type Model struct {
	id          uuid.UUID
	email       string
	displayName string
	givenName   string
	familyName  string
	avatarURL   string
	createdAt   time.Time
	updatedAt   time.Time
}

func (m Model) Id() uuid.UUID      { return m.id }
func (m Model) Email() string      { return m.email }
func (m Model) DisplayName() string { return m.displayName }
func (m Model) GivenName() string   { return m.givenName }
func (m Model) FamilyName() string  { return m.familyName }
func (m Model) AvatarURL() string   { return m.avatarURL }
func (m Model) CreatedAt() time.Time { return m.createdAt }
func (m Model) UpdatedAt() time.Time { return m.updatedAt }

// ToEntity converts the domain model back to a database entity.
func (m Model) ToEntity() Entity {
	return Entity{
		Id:          m.id,
		Email:       m.email,
		DisplayName: m.displayName,
		GivenName:   m.givenName,
		FamilyName:  m.familyName,
		AvatarURL:   m.avatarURL,
		CreatedAt:   m.createdAt,
		UpdatedAt:   m.updatedAt,
	}
}
