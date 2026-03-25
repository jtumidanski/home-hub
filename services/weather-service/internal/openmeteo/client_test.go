package openmeteo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchForecast(t *testing.T) {
	expected := ForecastResponse{
		Current: CurrentData{
			Temperature: 72.5,
			WeatherCode: 2,
		},
		Daily: DailyData{
			Time:           []string{"2026-03-25", "2026-03-26"},
			TemperatureMax: []float64{78.0, 65.0},
			TemperatureMin: []float64{55.0, 48.0},
			WeatherCode:    []int{2, 61},
		},
		Hourly: HourlyData{
			Time:                     []string{"2026-03-25T00:00", "2026-03-25T01:00", "2026-03-26T00:00"},
			Temperature:              []float64{58.0, 57.5, 50.0},
			WeatherCode:              []int{1, 1, 61},
			PrecipitationProbability: []int{0, 5, 70},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("latitude") == "" {
			t.Error("expected latitude parameter")
		}
		if r.URL.Query().Get("longitude") == "" {
			t.Error("expected longitude parameter")
		}
		if r.URL.Query().Get("temperature_unit") != "fahrenheit" {
			t.Errorf("expected fahrenheit, got %s", r.URL.Query().Get("temperature_unit"))
		}
		if r.URL.Query().Get("timezone") != "America/New_York" {
			t.Errorf("expected America/New_York, got %s", r.URL.Query().Get("timezone"))
		}
		if r.URL.Query().Get("forecast_days") != "7" {
			t.Errorf("expected forecast_days=7, got %s", r.URL.Query().Get("forecast_days"))
		}
		if r.URL.Query().Get("hourly") != "temperature_2m,weather_code,precipitation_probability" {
			t.Errorf("expected hourly params, got %s", r.URL.Query().Get("hourly"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, srv.URL)

	resp, err := c.FetchForecast(40.71, -74.01, "imperial", "America/New_York")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.Current.Temperature != 72.5 {
		t.Errorf("expected temperature 72.5, got %f", resp.Current.Temperature)
	}
	if resp.Current.WeatherCode != 2 {
		t.Errorf("expected weather code 2, got %d", resp.Current.WeatherCode)
	}
	if len(resp.Daily.Time) != 2 {
		t.Errorf("expected 2 days, got %d", len(resp.Daily.Time))
	}
	if len(resp.Hourly.Time) != 3 {
		t.Errorf("expected 3 hourly entries, got %d", len(resp.Hourly.Time))
	}
	if resp.Hourly.Temperature[0] != 58.0 {
		t.Errorf("expected hourly temp 58.0, got %f", resp.Hourly.Temperature[0])
	}
	if resp.Hourly.PrecipitationProbability[2] != 70 {
		t.Errorf("expected precip 70, got %d", resp.Hourly.PrecipitationProbability[2])
	}
}

func TestFetchForecastServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, srv.URL)

	_, err := c.FetchForecast(40.71, -74.01, "metric", "UTC")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestSearchPlaces(t *testing.T) {
	expected := GeocodingResponse{
		Results: []GeocodingResult{
			{ID: 5128581, Name: "New York", Country: "United States", Admin1: "New York", Latitude: 40.71, Longitude: -74.01},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("name") != "New York" {
			t.Errorf("expected name=New York, got %s", r.URL.Query().Get("name"))
		}
		if r.URL.Query().Get("count") != "10" {
			t.Errorf("expected count=10, got %s", r.URL.Query().Get("count"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, srv.URL)

	results, err := c.SearchPlaces("New York")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Name != "New York" {
		t.Errorf("expected New York, got %s", results[0].Name)
	}
	if results[0].Country != "United States" {
		t.Errorf("expected United States, got %s", results[0].Country)
	}
}

func TestSearchPlacesServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("unavailable"))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, srv.URL)

	_, err := c.SearchPlaces("London")
	if err == nil {
		t.Fatal("expected error for 503 response")
	}
}

func TestSearchPlacesEmptyResults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"results":[]}`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, srv.URL)

	results, err := c.SearchPlaces("xyznonexistent")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestTemperatureUnit(t *testing.T) {
	tests := []struct {
		units    string
		expected string
	}{
		{"imperial", "fahrenheit"},
		{"metric", "celsius"},
		{"", "celsius"},
	}

	for _, tt := range tests {
		t.Run(tt.units, func(t *testing.T) {
			result := temperatureUnit(tt.units)
			if result != tt.expected {
				t.Errorf("temperatureUnit(%q) = %q, want %q", tt.units, result, tt.expected)
			}
		})
	}
}
