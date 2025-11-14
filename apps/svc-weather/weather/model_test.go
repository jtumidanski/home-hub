package weather

import (
	"testing"
	"time"
)

func TestCurrentWeather_IsStale(t *testing.T) {
	tests := []struct {
		name           string
		observedAt     time.Time
		ttl            time.Duration
		expectedStale  bool
	}{
		{
			name:          "Fresh data",
			observedAt:    time.Now().Add(-2 * time.Minute),
			ttl:           5 * time.Minute,
			expectedStale: false,
		},
		{
			name:          "Stale data",
			observedAt:    time.Now().Add(-10 * time.Minute),
			ttl:           5 * time.Minute,
			expectedStale: true,
		},
		{
			name:          "Exactly at TTL boundary",
			observedAt:    time.Now().Add(-5 * time.Minute),
			ttl:           5 * time.Minute,
			expectedStale: true, // Should be stale at boundary
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current := NewCurrentWeather(15.5, tt.observedAt, "America/Detroit")

			if got := current.IsStale(tt.ttl); got != tt.expectedStale {
				t.Errorf("CurrentWeather.IsStale() = %v, want %v", got, tt.expectedStale)
			}
		})
	}
}

func TestCurrentWeather_Age(t *testing.T) {
	observedAt := time.Now().Add(-3 * time.Minute)
	current := NewCurrentWeather(15.5, observedAt, "America/Detroit")

	age := current.Age()

	if age < 2*time.Minute || age > 4*time.Minute {
		t.Errorf("CurrentWeather.Age() = %v, expected between 2-4 minutes", age)
	}
}

func TestCurrentWeather_Accessors(t *testing.T) {
	tempC := 15.5
	observedAt := time.Now()
	timezone := "America/Detroit"

	current := NewCurrentWeather(tempC, observedAt, timezone)

	if got := current.TemperatureC(); got != tempC {
		t.Errorf("CurrentWeather.TemperatureC() = %v, want %v", got, tempC)
	}

	if got := current.ObservedAt(); !got.Equal(observedAt) {
		t.Errorf("CurrentWeather.ObservedAt() = %v, want %v", got, observedAt)
	}

	if got := current.Timezone(); got != timezone {
		t.Errorf("CurrentWeather.Timezone() = %v, want %v", got, timezone)
	}
}

func TestDailyForecast_Accessors(t *testing.T) {
	date := time.Date(2025, 11, 15, 0, 0, 0, 0, time.UTC)
	tmaxC := 18.5
	tminC := 8.2

	daily := NewDailyForecast(date, tmaxC, tminC)

	if got := daily.Date(); !got.Equal(date) {
		t.Errorf("DailyForecast.Date() = %v, want %v", got, date)
	}

	if got := daily.TMaxC(); got != tmaxC {
		t.Errorf("DailyForecast.TMaxC() = %v, want %v", got, tmaxC)
	}

	if got := daily.TMinC(); got != tminC {
		t.Errorf("DailyForecast.TMinC() = %v, want %v", got, tminC)
	}
}

func TestForecastWeather_IsStale(t *testing.T) {
	tests := []struct {
		name          string
		generatedAt   time.Time
		ttl           time.Duration
		expectedStale bool
	}{
		{
			name:          "Fresh forecast",
			generatedAt:   time.Now().Add(-30 * time.Minute),
			ttl:           1 * time.Hour,
			expectedStale: false,
		},
		{
			name:          "Stale forecast",
			generatedAt:   time.Now().Add(-2 * time.Hour),
			ttl:           1 * time.Hour,
			expectedStale: true,
		},
		{
			name:          "Exactly at TTL boundary",
			generatedAt:   time.Now().Add(-1 * time.Hour),
			ttl:           1 * time.Hour,
			expectedStale: true, // Should be stale at boundary
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			days := []DailyForecast{
				NewDailyForecast(time.Now(), 20.0, 10.0),
			}
			forecast := NewForecastWeather(days, tt.generatedAt, "America/Detroit")

			if got := forecast.IsStale(tt.ttl); got != tt.expectedStale {
				t.Errorf("ForecastWeather.IsStale() = %v, want %v", got, tt.expectedStale)
			}
		})
	}
}

func TestForecastWeather_Immutability(t *testing.T) {
	days := []DailyForecast{
		NewDailyForecast(time.Now(), 20.0, 10.0),
		NewDailyForecast(time.Now().Add(24*time.Hour), 22.0, 12.0),
	}

	forecast := NewForecastWeather(days, time.Now(), "America/Detroit")

	// Get days and modify the slice
	gotDays := forecast.Days()
	gotDays[0] = NewDailyForecast(time.Now(), 99.9, 99.9)

	// Original forecast should be unchanged
	originalDays := forecast.Days()
	if originalDays[0].TMaxC() == 99.9 {
		t.Error("ForecastWeather.Days() is not immutable - original data was modified")
	}
}

func TestCombinedWeather_Accessors(t *testing.T) {
	current := NewCurrentWeather(15.5, time.Now(), "America/Detroit")
	days := []DailyForecast{
		NewDailyForecast(time.Now(), 20.0, 10.0),
	}
	forecast := NewForecastWeather(days, time.Now(), "America/Detroit")
	meta := NewWeatherMeta("open-meteo", "dpjz9", time.Now(), "America/Detroit")

	combined := NewCombinedWeather(&current, &forecast, meta)

	if !combined.HasCurrent() {
		t.Error("CombinedWeather.HasCurrent() = false, want true")
	}

	if !combined.HasForecast() {
		t.Error("CombinedWeather.HasForecast() = false, want true")
	}

	if combined.Current() == nil {
		t.Error("CombinedWeather.Current() = nil, want non-nil")
	}

	if combined.Forecast() == nil {
		t.Error("CombinedWeather.Forecast() = nil, want non-nil")
	}

	if combined.Meta().Source() != "open-meteo" {
		t.Errorf("CombinedWeather.Meta().Source() = %v, want open-meteo", combined.Meta().Source())
	}
}

func TestCombinedWeather_NilValues(t *testing.T) {
	meta := NewWeatherMeta("open-meteo", "dpjz9", time.Now(), "America/Detroit")

	// Test with nil current
	combined := NewCombinedWeather(nil, nil, meta)

	if combined.HasCurrent() {
		t.Error("CombinedWeather.HasCurrent() = true, want false")
	}

	if combined.HasForecast() {
		t.Error("CombinedWeather.HasForecast() = true, want false")
	}

	if combined.Current() != nil {
		t.Error("CombinedWeather.Current() should be nil")
	}

	if combined.Forecast() != nil {
		t.Error("CombinedWeather.Forecast() should be nil")
	}
}

func TestWeatherMeta_Accessors(t *testing.T) {
	source := "open-meteo"
	geokey := "dpjz9"
	refreshedAt := time.Now()
	timezone := "America/Detroit"

	meta := NewWeatherMeta(source, geokey, refreshedAt, timezone)

	if got := meta.Source(); got != source {
		t.Errorf("WeatherMeta.Source() = %v, want %v", got, source)
	}

	if got := meta.Geokey(); got != geokey {
		t.Errorf("WeatherMeta.Geokey() = %v, want %v", got, geokey)
	}

	if got := meta.RefreshedAt(); !got.Equal(refreshedAt) {
		t.Errorf("WeatherMeta.RefreshedAt() = %v, want %v", got, refreshedAt)
	}

	if got := meta.Timezone(); got != timezone {
		t.Errorf("WeatherMeta.Timezone() = %v, want %v", got, timezone)
	}
}
