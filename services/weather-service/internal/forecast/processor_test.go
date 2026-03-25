package forecast

import (
	"testing"

	"github.com/jtumidanski/home-hub/services/weather-service/internal/openmeteo"
)

func TestTransformResponse(t *testing.T) {
	resp := &openmeteo.ForecastResponse{
		Current: openmeteo.CurrentData{
			Temperature: 72.5,
			WeatherCode: 2,
		},
		Daily: openmeteo.DailyData{
			Time:           []string{"2026-03-25", "2026-03-26", "2026-03-27"},
			TemperatureMax: []float64{78.0, 65.0, 70.0},
			TemperatureMin: []float64{55.0, 48.0, 52.0},
			WeatherCode:    []int{2, 61, 0},
		},
	}

	current, daily := transformResponse(resp)

	if current.Temperature != 72.5 {
		t.Errorf("expected current temperature 72.5, got %f", current.Temperature)
	}
	if current.WeatherCode != 2 {
		t.Errorf("expected current weather code 2, got %d", current.WeatherCode)
	}
	if current.Summary != "Partly Cloudy" {
		t.Errorf("expected Partly Cloudy, got %s", current.Summary)
	}
	if current.Icon != "cloud-sun" {
		t.Errorf("expected cloud-sun, got %s", current.Icon)
	}

	if len(daily) != 3 {
		t.Fatalf("expected 3 daily forecasts, got %d", len(daily))
	}

	if daily[0].Date != "2026-03-25" {
		t.Errorf("expected date 2026-03-25, got %s", daily[0].Date)
	}
	if daily[0].HighTemperature != 78.0 {
		t.Errorf("expected high 78.0, got %f", daily[0].HighTemperature)
	}
	if daily[0].LowTemperature != 55.0 {
		t.Errorf("expected low 55.0, got %f", daily[0].LowTemperature)
	}
	if daily[0].Summary != "Partly Cloudy" {
		t.Errorf("expected Partly Cloudy, got %s", daily[0].Summary)
	}

	// Second day: rain
	if daily[1].Summary != "Rain" {
		t.Errorf("expected Rain, got %s", daily[1].Summary)
	}
	if daily[1].Icon != "cloud-rain" {
		t.Errorf("expected cloud-rain, got %s", daily[1].Icon)
	}

	// Third day: clear
	if daily[2].Summary != "Clear" {
		t.Errorf("expected Clear, got %s", daily[2].Summary)
	}
	if daily[2].Icon != "sun" {
		t.Errorf("expected sun, got %s", daily[2].Icon)
	}
}

func TestTransformResponseEmpty(t *testing.T) {
	resp := &openmeteo.ForecastResponse{
		Current: openmeteo.CurrentData{
			Temperature: 0,
			WeatherCode: 0,
		},
		Daily: openmeteo.DailyData{
			Time:           []string{},
			TemperatureMax: []float64{},
			TemperatureMin: []float64{},
			WeatherCode:    []int{},
		},
	}

	current, daily := transformResponse(resp)

	if current.Summary != "Clear" {
		t.Errorf("expected Clear for code 0, got %s", current.Summary)
	}
	if len(daily) != 0 {
		t.Errorf("expected 0 daily forecasts, got %d", len(daily))
	}
}

func TestTransformResponseUnknownWeatherCode(t *testing.T) {
	resp := &openmeteo.ForecastResponse{
		Current: openmeteo.CurrentData{
			Temperature: 20.0,
			WeatherCode: 999,
		},
		Daily: openmeteo.DailyData{
			Time:           []string{"2026-03-25"},
			TemperatureMax: []float64{25.0},
			TemperatureMin: []float64{15.0},
			WeatherCode:    []int{999},
		},
	}

	current, daily := transformResponse(resp)

	if current.Summary != "Unknown" {
		t.Errorf("expected Unknown for code 999, got %s", current.Summary)
	}
	if current.Icon != "cloud" {
		t.Errorf("expected cloud fallback icon, got %s", current.Icon)
	}
	if daily[0].Summary != "Unknown" {
		t.Errorf("expected Unknown for daily code 999, got %s", daily[0].Summary)
	}
}
