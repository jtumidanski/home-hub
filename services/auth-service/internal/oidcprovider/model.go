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

// Make converts an Entity to a Model.
func Make(e Entity) (Model, error) {
	return Model{
		id:        e.Id,
		name:      e.Name,
		issuerURL: e.IssuerURL,
		clientID:  e.ClientID,
		enabled:   e.Enabled,
	}, nil
}
