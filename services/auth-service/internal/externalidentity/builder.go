package externalidentity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrUserIDRequired   = errors.New("external identity user ID is required")
	ErrProviderRequired = errors.New("external identity provider is required")
	ErrSubjectRequired  = errors.New("external identity subject is required")
)

type Builder struct {
	id              uuid.UUID
	userId          uuid.UUID
	provider        string
	providerSubject string
	createdAt       time.Time
	updatedAt       time.Time
}

func NewBuilder() *Builder {
	return &Builder{}
}

func (b *Builder) SetId(id uuid.UUID) *Builder              { b.id = id; return b }
func (b *Builder) SetUserId(id uuid.UUID) *Builder          { b.userId = id; return b }
func (b *Builder) SetProvider(provider string) *Builder     { b.provider = provider; return b }
func (b *Builder) SetProviderSubject(sub string) *Builder   { b.providerSubject = sub; return b }
func (b *Builder) SetCreatedAt(t time.Time) *Builder        { b.createdAt = t; return b }
func (b *Builder) SetUpdatedAt(t time.Time) *Builder        { b.updatedAt = t; return b }

func (b *Builder) Build() (Model, error) {
	if b.userId == uuid.Nil {
		return Model{}, ErrUserIDRequired
	}
	if b.provider == "" {
		return Model{}, ErrProviderRequired
	}
	if b.providerSubject == "" {
		return Model{}, ErrSubjectRequired
	}
	return Model{
		id:              b.id,
		userId:          b.userId,
		provider:        b.provider,
		providerSubject: b.providerSubject,
		createdAt:       b.createdAt,
		updatedAt:       b.updatedAt,
	}, nil
}
