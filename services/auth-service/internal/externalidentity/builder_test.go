package externalidentity

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
			name: "valid identity",
			build: func() (Model, error) {
				return NewBuilder().
					SetId(uuid.New()).
					SetUserId(uuid.New()).
					SetProvider("google").
					SetProviderSubject("sub-123").
					Build()
			},
			wantErr: nil,
		},
		{
			name: "missing user ID",
			build: func() (Model, error) {
				return NewBuilder().
					SetProvider("google").
					SetProviderSubject("sub-123").
					Build()
			},
			wantErr: ErrUserIDRequired,
		},
		{
			name: "missing provider",
			build: func() (Model, error) {
				return NewBuilder().
					SetUserId(uuid.New()).
					SetProviderSubject("sub-123").
					Build()
			},
			wantErr: ErrProviderRequired,
		},
		{
			name: "missing subject",
			build: func() (Model, error) {
				return NewBuilder().
					SetUserId(uuid.New()).
					SetProvider("google").
					Build()
			},
			wantErr: ErrSubjectRequired,
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
			if m.Provider() != "google" {
				t.Errorf("expected provider google, got %s", m.Provider())
			}
		})
	}
}
