package forecast

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMake(t *testing.T) {
	id := uuid.New()
	tenantID := uuid.New()
	householdID := uuid.New()
	now := time.Now().UTC()

	e := Entity{
		Id:          id,
		TenantId:    tenantID,
		HouseholdId: householdID,
		Latitude:    40.71,
		Longitude:   -74.01,
		Units:       "imperial",
		CurrentData: JSONCurrentData{
			Temperature: 72.5,
			WeatherCode: 2,
			Summary:     "Partly Cloudy",
			Icon:        "cloud-sun",
		},
		ForecastData: JSONForecastData{
			{
				Date: "2026-03-25", HighTemperature: 78.0, LowTemperature: 55.0, WeatherCode: 2, Summary: "Partly Cloudy", Icon: "cloud-sun",
				HourlyForecast: []HourlyForecast{
					{Time: "2026-03-25T00:00", Temperature: 58.0, WeatherCode: 1, Summary: "Mostly Clear", Icon: "sun", PrecipitationProbability: 0},
					{Time: "2026-03-25T01:00", Temperature: 57.5, WeatherCode: 1, Summary: "Mostly Clear", Icon: "sun", PrecipitationProbability: 5},
				},
			},
		},
		FetchedAt: now,
		CreatedAt: now,
		UpdatedAt: now,
	}

	m, err := Make(e)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if m.Id() != id {
		t.Errorf("expected id %s, got %s", id, m.Id())
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
	if m.Units() != "imperial" {
		t.Errorf("expected imperial, got %s", m.Units())
	}
	if m.CurrentData().Temperature != 72.5 {
		t.Errorf("expected temperature 72.5, got %f", m.CurrentData().Temperature)
	}
	if len(m.ForecastData()) != 1 {
		t.Errorf("expected 1 forecast day, got %d", len(m.ForecastData()))
	}
	if m.TemperatureUnit() != "°F" {
		t.Errorf("expected °F, got %s", m.TemperatureUnit())
	}
	if len(m.ForecastData()[0].HourlyForecast) != 2 {
		t.Errorf("expected 2 hourly entries, got %d", len(m.ForecastData()[0].HourlyForecast))
	}
	if m.ForecastData()[0].HourlyForecast[0].Temperature != 58.0 {
		t.Errorf("expected hourly temp 58.0, got %f", m.ForecastData()[0].HourlyForecast[0].Temperature)
	}
}

func TestToEntity(t *testing.T) {
	tenantID := uuid.New()
	householdID := uuid.New()
	now := time.Now().UTC()

	m, err := NewBuilder().
		SetId(uuid.New()).
		SetTenantID(tenantID).
		SetHouseholdID(householdID).
		SetLatitude(40.71).
		SetLongitude(-74.01).
		SetUnits("metric").
		SetCurrentData(CurrentData{Temperature: 22.0, WeatherCode: 0, Summary: "Clear", Icon: "sun"}).
		SetForecastData([]DailyForecast{{Date: "2026-03-25", HighTemperature: 25.0, LowTemperature: 12.0}}).
		SetFetchedAt(now).
		SetCreatedAt(now).
		SetUpdatedAt(now).
		Build()
	if err != nil {
		t.Fatalf("expected no error building model, got %v", err)
	}

	e := m.ToEntity()

	if e.Id != m.Id() {
		t.Errorf("expected entity id %s, got %s", m.Id(), e.Id)
	}
	if e.TenantId != tenantID {
		t.Errorf("expected entity tenantId %s, got %s", tenantID, e.TenantId)
	}
	if e.HouseholdId != householdID {
		t.Errorf("expected entity householdId %s, got %s", householdID, e.HouseholdId)
	}
	if e.Latitude != 40.71 {
		t.Errorf("expected entity latitude 40.71, got %f", e.Latitude)
	}
	if e.Longitude != -74.01 {
		t.Errorf("expected entity longitude -74.01, got %f", e.Longitude)
	}
	if e.Units != "metric" {
		t.Errorf("expected entity units metric, got %s", e.Units)
	}
	if CurrentData(e.CurrentData).Temperature != 22.0 {
		t.Errorf("expected entity current temperature 22.0, got %f", CurrentData(e.CurrentData).Temperature)
	}
	if len(e.ForecastData) != 1 {
		t.Errorf("expected 1 forecast day in entity, got %d", len(e.ForecastData))
	}
}

func TestMakeAndToEntityRoundTrip(t *testing.T) {
	id := uuid.New()
	tenantID := uuid.New()
	householdID := uuid.New()
	now := time.Now().UTC()

	original := Entity{
		Id:          id,
		TenantId:    tenantID,
		HouseholdId: householdID,
		Latitude:    51.5,
		Longitude:   -0.12,
		Units:       "metric",
		CurrentData: JSONCurrentData{Temperature: 15.0, WeatherCode: 3, Summary: "Overcast", Icon: "cloud"},
		ForecastData: JSONForecastData{
			{
				Date: "2026-03-25", HighTemperature: 18.0, LowTemperature: 8.0, WeatherCode: 3, Summary: "Overcast", Icon: "cloud",
				HourlyForecast: []HourlyForecast{
					{Time: "2026-03-25T12:00", Temperature: 16.0, WeatherCode: 3, Summary: "Overcast", Icon: "cloud", PrecipitationProbability: 20},
				},
			},
		},
		FetchedAt: now,
		CreatedAt: now,
		UpdatedAt: now,
	}

	m, err := Make(original)
	if err != nil {
		t.Fatalf("Make failed: %v", err)
	}

	roundTripped := m.ToEntity()

	if roundTripped.Id != original.Id {
		t.Errorf("roundtrip id mismatch: %s vs %s", roundTripped.Id, original.Id)
	}
	if roundTripped.TenantId != original.TenantId {
		t.Errorf("roundtrip tenantId mismatch")
	}
	if roundTripped.HouseholdId != original.HouseholdId {
		t.Errorf("roundtrip householdId mismatch")
	}
	if roundTripped.Latitude != original.Latitude {
		t.Errorf("roundtrip latitude mismatch: %f vs %f", roundTripped.Latitude, original.Latitude)
	}
	if roundTripped.Longitude != original.Longitude {
		t.Errorf("roundtrip longitude mismatch")
	}
	if roundTripped.Units != original.Units {
		t.Errorf("roundtrip units mismatch")
	}
	if len(roundTripped.ForecastData[0].HourlyForecast) != 1 {
		t.Errorf("roundtrip hourly data mismatch: expected 1, got %d", len(roundTripped.ForecastData[0].HourlyForecast))
	}
	if roundTripped.ForecastData[0].HourlyForecast[0].Temperature != 16.0 {
		t.Errorf("roundtrip hourly temp mismatch: expected 16.0, got %f", roundTripped.ForecastData[0].HourlyForecast[0].Temperature)
	}
}
