package forecast

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestBuilderValid(t *testing.T) {
	tenantID := uuid.New()
	householdID := uuid.New()
	now := time.Now()

	m, err := NewBuilder().
		SetId(uuid.New()).
		SetTenantID(tenantID).
		SetHouseholdID(householdID).
		SetLatitude(40.71).
		SetLongitude(-74.01).
		SetUnits("imperial").
		SetCurrentData(CurrentData{Temperature: 72.5, WeatherCode: 2, Summary: "Partly Cloudy", Icon: "cloud-sun"}).
		SetForecastData([]DailyForecast{{Date: "2026-03-25", HighTemperature: 78.0, LowTemperature: 55.0}}).
		SetFetchedAt(now).
		SetCreatedAt(now).
		SetUpdatedAt(now).
		Build()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if m.TenantID() != tenantID {
		t.Errorf("expected tenantID %s, got %s", tenantID, m.TenantID())
	}
	if m.HouseholdID() != householdID {
		t.Errorf("expected householdID %s, got %s", householdID, m.HouseholdID())
	}
	if m.Latitude() != 40.71 {
		t.Errorf("expected latitude 40.71, got %f", m.Latitude())
	}
	if m.Longitude() != -74.01 {
		t.Errorf("expected longitude -74.01, got %f", m.Longitude())
	}
	if m.Units() != "imperial" {
		t.Errorf("expected imperial, got %s", m.Units())
	}
	if m.CurrentData().Temperature != 72.5 {
		t.Errorf("expected temperature 72.5, got %f", m.CurrentData().Temperature)
	}
	if len(m.ForecastData()) != 1 {
		t.Errorf("expected 1 forecast day, got %d", len(m.ForecastData()))
	}
}

func TestBuilderMissingTenantID(t *testing.T) {
	_, err := NewBuilder().
		SetHouseholdID(uuid.New()).
		SetLatitude(40.71).
		SetLongitude(-74.01).
		SetUnits("metric").
		Build()

	if err != ErrTenantIDRequired {
		t.Errorf("expected ErrTenantIDRequired, got %v", err)
	}
}

func TestBuilderMissingHouseholdID(t *testing.T) {
	_, err := NewBuilder().
		SetTenantID(uuid.New()).
		SetLatitude(40.71).
		SetLongitude(-74.01).
		SetUnits("metric").
		Build()

	if err != ErrHouseholdIDRequired {
		t.Errorf("expected ErrHouseholdIDRequired, got %v", err)
	}
}

func TestBuilderLatitudeOutOfRange(t *testing.T) {
	tests := []struct {
		name string
		lat  float64
	}{
		{"too low", -91.0},
		{"too high", 91.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewBuilder().
				SetTenantID(uuid.New()).
				SetHouseholdID(uuid.New()).
				SetLatitude(tt.lat).
				SetLongitude(-74.01).
				SetUnits("metric").
				Build()

			if err != ErrLatitudeOutOfRange {
				t.Errorf("expected ErrLatitudeOutOfRange, got %v", err)
			}
		})
	}
}

func TestBuilderLongitudeOutOfRange(t *testing.T) {
	tests := []struct {
		name string
		lon  float64
	}{
		{"too low", -181.0},
		{"too high", 181.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewBuilder().
				SetTenantID(uuid.New()).
				SetHouseholdID(uuid.New()).
				SetLatitude(40.71).
				SetLongitude(tt.lon).
				SetUnits("metric").
				Build()

			if err != ErrLongitudeOutOfRange {
				t.Errorf("expected ErrLongitudeOutOfRange, got %v", err)
			}
		})
	}
}

func TestBuilderMissingUnits(t *testing.T) {
	_, err := NewBuilder().
		SetTenantID(uuid.New()).
		SetHouseholdID(uuid.New()).
		SetLatitude(40.71).
		SetLongitude(-74.01).
		Build()

	if err != ErrUnitsRequired {
		t.Errorf("expected ErrUnitsRequired, got %v", err)
	}
}

func TestBuilderBoundaryCoordinates(t *testing.T) {
	tests := []struct {
		name string
		lat  float64
		lon  float64
	}{
		{"min bounds", -90.0, -180.0},
		{"max bounds", 90.0, 180.0},
		{"zero", 0.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewBuilder().
				SetTenantID(uuid.New()).
				SetHouseholdID(uuid.New()).
				SetLatitude(tt.lat).
				SetLongitude(tt.lon).
				SetUnits("metric").
				Build()

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if m.Latitude() != tt.lat {
				t.Errorf("expected latitude %f, got %f", tt.lat, m.Latitude())
			}
			if m.Longitude() != tt.lon {
				t.Errorf("expected longitude %f, got %f", tt.lon, m.Longitude())
			}
		})
	}
}
