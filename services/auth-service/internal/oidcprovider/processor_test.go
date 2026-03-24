package oidcprovider

import (
	"testing"

	"github.com/jtumidanski/home-hub/services/auth-service/internal/config"
	"github.com/sirupsen/logrus/hooks/test"
)

func TestListEnabled_WithClientID(t *testing.T) {
	l, _ := test.NewNullLogger()
	cfg := config.OIDCConfig{
		IssuerURL: "https://accounts.google.com",
		ClientID:  "test-client-id",
	}

	p := NewProcessor(l, cfg)
	models, err := p.ListEnabled()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(models) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(models))
	}
	if models[0].Name() != "Google" {
		t.Errorf("expected name Google, got %s", models[0].Name())
	}
	if models[0].ClientID() != "test-client-id" {
		t.Errorf("expected client ID test-client-id, got %s", models[0].ClientID())
	}
	if !models[0].Enabled() {
		t.Error("expected provider to be enabled")
	}
}

func TestListEnabled_NoClientID(t *testing.T) {
	l, _ := test.NewNullLogger()
	cfg := config.OIDCConfig{}

	p := NewProcessor(l, cfg)
	models, err := p.ListEnabled()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(models) != 0 {
		t.Errorf("expected 0 providers, got %d", len(models))
	}
}
