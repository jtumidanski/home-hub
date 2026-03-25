package tenant

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestBuilder(t *testing.T) {
	tests := []struct {
		name    string
		build   func() (Model, error)
		wantErr error
	}{
		{
			name: "valid tenant",
			build: func() (Model, error) {
				return NewBuilder().
					SetId(uuid.New()).
					SetName("Test Tenant").
					SetCreatedAt(time.Now()).
					SetUpdatedAt(time.Now()).
					Build()
			},
		},
		{
			name: "missing name",
			build: func() (Model, error) {
				return NewBuilder().
					SetId(uuid.New()).
					Build()
			},
			wantErr: ErrNameRequired,
		},
		{
			name: "empty name",
			build: func() (Model, error) {
				return NewBuilder().
					SetId(uuid.New()).
					SetName("").
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
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
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
