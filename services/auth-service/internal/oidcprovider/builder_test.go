package oidcprovider

import (
	"testing"

	"github.com/google/uuid"
)

func TestBuilder(t *testing.T) {
	tests := []struct {
		name    string
		build   func() (Model, error)
		wantErr error
	}{
		{
			name: "valid provider",
			build: func() (Model, error) {
				return NewBuilder().
					SetId(uuid.New()).
					SetName("Google").
					SetIssuerURL("https://accounts.google.com").
					SetClientID("client-123").
					SetEnabled(true).
					Build()
			},
			wantErr: nil,
		},
		{
			name: "missing name",
			build: func() (Model, error) {
				return NewBuilder().
					SetId(uuid.New()).
					SetIssuerURL("https://accounts.google.com").
					Build()
			},
			wantErr: ErrNameRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := tt.build()
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("Build() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if m.Name() == "" {
				t.Error("expected non-empty name")
			}
		})
	}
}
