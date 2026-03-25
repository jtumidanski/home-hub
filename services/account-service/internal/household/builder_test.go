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
		{
			name: "valid with location",
			build: func() (Model, error) {
				lat := 40.7128
				lon := -74.006
				locName := "New York"
				return NewBuilder().
					SetId(uuid.New()).
					SetTenantID(uuid.New()).
					SetName("My Home").
					SetTimezone("America/New_York").
					SetUnits("imperial").
					SetLatitude(&lat).
					SetLongitude(&lon).
					SetLocationName(&locName).
					SetCreatedAt(time.Now()).
					SetUpdatedAt(time.Now()).
					Build()
			},
		},
		{
			name: "valid without location",
			build: func() (Model, error) {
				return NewBuilder().
					SetName("My Home").
					SetTimezone("UTC").
					SetUnits("metric").
					SetLatitude(nil).
					SetLongitude(nil).
					Build()
			},
		},
		{
			name: "partial coordinates - latitude only",
			build: func() (Model, error) {
				lat := 40.7128
				return NewBuilder().
					SetName("Home").
					SetTimezone("UTC").
					SetUnits("metric").
					SetLatitude(&lat).
					Build()
			},
			wantErr: ErrPartialCoordinates,
		},
		{
			name: "partial coordinates - longitude only",
			build: func() (Model, error) {
				lon := -74.006
				return NewBuilder().
					SetName("Home").
					SetTimezone("UTC").
					SetUnits("metric").
					SetLongitude(&lon).
					Build()
			},
			wantErr: ErrPartialCoordinates,
		},
		{
			name: "latitude out of range",
			build: func() (Model, error) {
				lat := 91.0
				lon := 0.0
				return NewBuilder().
					SetName("Home").
					SetTimezone("UTC").
					SetUnits("metric").
					SetLatitude(&lat).
					SetLongitude(&lon).
					Build()
			},
			wantErr: ErrLatitudeOutOfRange,
		},
		{
			name: "longitude out of range",
			build: func() (Model, error) {
				lat := 0.0
				lon := 181.0
				return NewBuilder().
					SetName("Home").
					SetTimezone("UTC").
					SetUnits("metric").
					SetLatitude(&lat).
					SetLongitude(&lon).
					Build()
			},
			wantErr: ErrLongitudeOutOfRange,
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
