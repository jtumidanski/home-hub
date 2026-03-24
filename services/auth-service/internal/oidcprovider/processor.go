package oidcprovider

import (
	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/auth-service/internal/config"
	"github.com/sirupsen/logrus"
)

// googleProviderID is a fixed UUID for the Google OIDC provider.
var googleProviderID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

// Processor provides OIDC provider information from configuration.
// When DB-backed provider management is needed, this can be updated to
// read from the database instead.
type Processor struct {
	l    logrus.FieldLogger
	oidc config.OIDCConfig
}

func NewProcessor(l logrus.FieldLogger, oidc config.OIDCConfig) *Processor {
	return &Processor{l: l, oidc: oidc}
}

// ListEnabled returns all enabled OIDC providers.
func (p *Processor) ListEnabled() ([]Model, error) {
	var models []Model
	if p.oidc.ClientID != "" {
		m, err := NewBuilder().
			SetId(googleProviderID).
			SetName("Google").
			SetIssuerURL(p.oidc.IssuerURL).
			SetClientID(p.oidc.ClientID).
			SetEnabled(true).
			Build()
		if err != nil {
			return nil, err
		}
		models = append(models, m)
	}
	return models, nil
}
