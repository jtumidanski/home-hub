package oidcprovider

import (
	"testing"

	"github.com/jtumidanski/home-hub/services/auth-service/internal/config"
	"github.com/sirupsen/logrus/hooks/test"
)

func TestListEnabled(t *testing.T) {
	tests := []struct {
		name      string
		cfg       config.OIDCConfig
		wantCount int
	}{
		{
			name: "with client ID returns provider",
			cfg: config.OIDCConfig{
				IssuerURL: "https://accounts.google.com",
				ClientID:  "test-client-id",
			},
			wantCount: 1,
		},
		{
			name:      "no client ID returns empty",
			cfg:       config.OIDCConfig{},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, _ := test.NewNullLogger()
			p := NewProcessor(l, tt.cfg)
			models, err := p.ListEnabled()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(models) != tt.wantCount {
				t.Fatalf("expected %d providers, got %d", tt.wantCount, len(models))
			}
			if tt.wantCount > 0 {
				if models[0].Name() != "Google" {
					t.Errorf("expected name Google, got %s", models[0].Name())
				}
				if models[0].ClientID() != tt.cfg.ClientID {
					t.Errorf("expected client ID %s, got %s", tt.cfg.ClientID, models[0].ClientID())
				}
				if !models[0].Enabled() {
					t.Error("expected provider to be enabled")
				}
			}
		})
	}
}
