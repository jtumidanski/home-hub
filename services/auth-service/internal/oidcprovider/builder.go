package oidcprovider

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrNameRequired = errors.New("OIDC provider name is required")
)

type Builder struct {
	id        uuid.UUID
	name      string
	issuerURL string
	clientID  string
	enabled   bool
}

func NewBuilder() *Builder {
	return &Builder{}
}

func (b *Builder) SetId(id uuid.UUID) *Builder         { b.id = id; return b }
func (b *Builder) SetName(name string) *Builder         { b.name = name; return b }
func (b *Builder) SetIssuerURL(url string) *Builder     { b.issuerURL = url; return b }
func (b *Builder) SetClientID(cid string) *Builder      { b.clientID = cid; return b }
func (b *Builder) SetEnabled(enabled bool) *Builder     { b.enabled = enabled; return b }

func (b *Builder) Build() (Model, error) {
	if b.name == "" {
		return Model{}, ErrNameRequired
	}
	return Model{
		id:        b.id,
		name:      b.name,
		issuerURL: b.issuerURL,
		clientID:  b.clientID,
		enabled:   b.enabled,
	}, nil
}
