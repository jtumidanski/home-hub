package user

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEmailRequired       = errors.New("user email is required")
	ErrDisplayNameRequired = errors.New("user display name is required")
)

type Builder struct {
	id                uuid.UUID
	email             string
	displayName       string
	givenName         string
	familyName        string
	avatarURL         string
	providerAvatarURL string
	createdAt         time.Time
	updatedAt         time.Time
}

func NewBuilder() *Builder {
	return &Builder{}
}

func (b *Builder) SetId(id uuid.UUID) *Builder              { b.id = id; return b }
func (b *Builder) SetEmail(email string) *Builder            { b.email = email; return b }
func (b *Builder) SetDisplayName(name string) *Builder       { b.displayName = name; return b }
func (b *Builder) SetGivenName(name string) *Builder         { b.givenName = name; return b }
func (b *Builder) SetFamilyName(name string) *Builder        { b.familyName = name; return b }
func (b *Builder) SetAvatarURL(url string) *Builder          { b.avatarURL = url; return b }
func (b *Builder) SetProviderAvatarURL(url string) *Builder  { b.providerAvatarURL = url; return b }
func (b *Builder) SetCreatedAt(t time.Time) *Builder         { b.createdAt = t; return b }
func (b *Builder) SetUpdatedAt(t time.Time) *Builder         { b.updatedAt = t; return b }

func (b *Builder) Build() (Model, error) {
	if b.email == "" {
		return Model{}, ErrEmailRequired
	}
	if b.displayName == "" {
		return Model{}, ErrDisplayNameRequired
	}
	return Model{
		id:                b.id,
		email:             b.email,
		displayName:       b.displayName,
		givenName:         b.givenName,
		familyName:        b.familyName,
		avatarURL:         b.avatarURL,
		providerAvatarURL: b.providerAvatarURL,
		createdAt:         b.createdAt,
		updatedAt:         b.updatedAt,
	}, nil
}
