package household

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
			name: "valid household",
			build: func() (Model, error) {
				return NewBuilder().
					SetId(uuid.New()).
					SetTenantID(uuid.New()).
					SetName("My Home").
					SetTimezone("America/Detroit").
					SetUnits("imperial").
					SetCreatedAt(time.Now()).
					SetUpdatedAt(time.Now()).
					Build()
			},
		},
		{
			name: "missing name",
			build: func() (Model, error) {
				return NewBuilder().
					SetTimezone("UTC").
					SetUnits("metric").
					Build()
			},
			wantErr: ErrNameRequired,
		},
		{
			name: "missing timezone",
			build: func() (Model, error) {
				return NewBuilder().
					SetName("Home").
					SetUnits("metric").
					Build()
			},
			wantErr: ErrTimezoneRequired,
		},
		{
			name: "missing units",
			build: func() (Model, error) {
				return NewBuilder().
					SetName("Home").
					SetTimezone("UTC").
					Build()
			},
			wantErr: ErrUnitsRequired,
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
