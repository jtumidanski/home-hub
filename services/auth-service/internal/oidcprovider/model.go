package oidcprovider

import "github.com/google/uuid"

// Model is the immutable domain model for an OIDC provider.
type Model struct {
	id        uuid.UUID
	name      string
	issuerURL string
	clientID  string
	enabled   bool
}

func (m Model) Id() uuid.UUID   { return m.id }
func (m Model) Name() string    { return m.name }
func (m Model) IssuerURL() string { return m.issuerURL }
func (m Model) ClientID() string { return m.clientID }
func (m Model) Enabled() bool   { return m.enabled }

// ToEntity converts the domain model back to a database entity.
func (m Model) ToEntity() Entity {
	return Entity{
		Id:        m.id,
		Name:      m.name,
		IssuerURL: m.issuerURL,
		ClientID:  m.clientID,
		Enabled:   m.enabled,
	}
}

