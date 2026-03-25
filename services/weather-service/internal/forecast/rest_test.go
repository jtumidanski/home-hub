package forecast

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestTransformCurrent(t *testing.T) {
	m := Model{
		id:          uuid.New(),
		householdID: uuid.New(),
		units:       "imperial",
		currentData: CurrentData{
			Temperature: 72.5,
			WeatherCode: 2,
			Summary:     "Partly Cloudy",
			Icon:        "cloud-sun",
		},
		forecastData: []DailyForecast{
			{Date: "2026-03-25", HighTemperature: 78.0, LowTemperature: 55.0, WeatherCode: 2, Summary: "Partly Cloudy", Icon: "cloud-sun"},
		},
		fetchedAt: time.Now(),
	}

	rest := TransformCurrent(m)

	if rest.Temperature != 72.5 {
		t.Errorf("expected temperature 72.5, got %f", rest.Temperature)
	}
	if rest.TemperatureUnit != "°F" {
		t.Errorf("expected °F, got %s", rest.TemperatureUnit)
	}
	if rest.Summary != "Partly Cloudy" {
		t.Errorf("expected Partly Cloudy, got %s", rest.Summary)
	}
	if rest.Icon != "cloud-sun" {
		t.Errorf("expected cloud-sun, got %s", rest.Icon)
	}
	if rest.HighTemperature != 78.0 {
		t.Errorf("expected high 78.0, got %f", rest.HighTemperature)
	}
	if rest.LowTemperature != 55.0 {
		t.Errorf("expected low 55.0, got %f", rest.LowTemperature)
	}
}

func TestTransformCurrentMetric(t *testing.T) {
	m := Model{
		units: "metric",
		currentData: CurrentData{
			Temperature: 22.0,
			WeatherCode: 0,
			Summary:     "Clear",
			Icon:        "sun",
		},
		forecastData: []DailyForecast{
			{Date: "2026-03-25", HighTemperature: 25.0, LowTemperature: 12.0},
		},
	}

	rest := TransformCurrent(m)
	if rest.TemperatureUnit != "°C" {
		t.Errorf("expected °C, got %s", rest.TemperatureUnit)
	}
}

func TestTransformForecast(t *testing.T) {
	m := Model{
		units: "imperial",
		forecastData: []DailyForecast{
			{
				Date: "2026-03-25", HighTemperature: 78.0, LowTemperature: 55.0, WeatherCode: 2, Summary: "Partly Cloudy", Icon: "cloud-sun",
				HourlyForecast: []HourlyForecast{
					{Time: "2026-03-25T00:00", Temperature: 58.0, WeatherCode: 1, Summary: "Mostly Clear", Icon: "sun", PrecipitationProbability: 0},
					{Time: "2026-03-25T01:00", Temperature: 57.5, WeatherCode: 2, Summary: "Partly Cloudy", Icon: "cloud-sun", PrecipitationProbability: 5},
				},
			},
			{Date: "2026-03-26", HighTemperature: 65.0, LowTemperature: 48.0, WeatherCode: 61, Summary: "Rain", Icon: "cloud-rain"},
		},
	}

	rest := TransformForecast(m)

	if len(rest) != 2 {
		t.Fatalf("expected 2 days, got %d", len(rest))
	}
	if rest[0].Date != "2026-03-25" {
		t.Errorf("expected date 2026-03-25, got %s", rest[0].Date)
	}
	if rest[0].TemperatureUnit != "°F" {
		t.Errorf("expected °F, got %s", rest[0].TemperatureUnit)
	}
	if len(rest[0].HourlyForecast) != 2 {
		t.Fatalf("expected 2 hourly entries, got %d", len(rest[0].HourlyForecast))
	}
	if rest[0].HourlyForecast[0].Time != "2026-03-25T00:00" {
		t.Errorf("expected time 2026-03-25T00:00, got %s", rest[0].HourlyForecast[0].Time)
	}
	if rest[0].HourlyForecast[0].Temperature != 58.0 {
		t.Errorf("expected temp 58.0, got %f", rest[0].HourlyForecast[0].Temperature)
	}
	if rest[0].HourlyForecast[0].PrecipitationProbability != 0 {
		t.Errorf("expected precip 0, got %d", rest[0].HourlyForecast[0].PrecipitationProbability)
	}
	if rest[0].HourlyForecast[1].Summary != "Partly Cloudy" {
		t.Errorf("expected Partly Cloudy, got %s", rest[0].HourlyForecast[1].Summary)
	}
	if rest[1].Summary != "Rain" {
		t.Errorf("expected Rain, got %s", rest[1].Summary)
	}
	if rest[1].Icon != "cloud-rain" {
		t.Errorf("expected cloud-rain, got %s", rest[1].Icon)
	}
	if len(rest[1].HourlyForecast) != 0 {
		t.Errorf("expected 0 hourly entries for day 2, got %d", len(rest[1].HourlyForecast))
	}
}

func TestTransformCurrentNoForecastData(t *testing.T) {
	m := Model{
		units: "metric",
		currentData: CurrentData{
			Temperature: 15.0,
			WeatherCode: 3,
			Summary:     "Overcast",
			Icon:        "cloud",
		},
		forecastData: []DailyForecast{},
	}

	rest := TransformCurrent(m)
	if rest.HighTemperature != 0 {
		t.Errorf("expected 0 high when no forecast data, got %f", rest.HighTemperature)
	}
}
