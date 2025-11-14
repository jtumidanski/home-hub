package weather

import (
	"testing"
	"time"
)

func TestCurrentCacheEntry_Serialization(t *testing.T) {
	current := NewCurrentWeather(15.5, time.Now(), "America/Detroit")
	ttl := 5 * time.Minute

	entry := NewCurrentCacheEntry(current, ttl)

	// Marshal to JSON
	data, err := entry.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshal cache entry: %v", err)
	}

	// Unmarshal from JSON
	unmarshaled, err := UnmarshalCurrentCacheEntry(data)
	if err != nil {
		t.Fatalf("Failed to unmarshal cache entry: %v", err)
	}

	// Verify data integrity
	if unmarshaled.TemperatureC != entry.TemperatureC {
		t.Errorf("TemperatureC = %v, want %v", unmarshaled.TemperatureC, entry.TemperatureC)
	}

	if unmarshaled.Timezone != entry.Timezone {
		t.Errorf("Timezone = %v, want %v", unmarshaled.Timezone, entry.Timezone)
	}

	if unmarshaled.TTL != entry.TTL {
		t.Errorf("TTL = %v, want %v", unmarshaled.TTL, entry.TTL)
	}
}

func TestCurrentCacheEntry_IsStale(t *testing.T) {
	tests := []struct {
		name          string
		cachedAt      time.Time
		ttl           time.Duration
		expectedStale bool
	}{
		{
			name:          "Fresh cache",
			cachedAt:      time.Now().Add(-2 * time.Minute),
			ttl:           5 * time.Minute,
			expectedStale: false,
		},
		{
			name:          "Stale cache",
			cachedAt:      time.Now().Add(-10 * time.Minute),
			ttl:           5 * time.Minute,
			expectedStale: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := CurrentCacheEntry{
				TemperatureC: 15.5,
				ObservedAt:   time.Now(),
				Timezone:     "America/Detroit",
				CachedAt:     tt.cachedAt,
				TTL:          int64(tt.ttl.Seconds()),
			}

			if got := entry.IsStale(); got != tt.expectedStale {
				t.Errorf("CurrentCacheEntry.IsStale() = %v, want %v", got, tt.expectedStale)
			}
		})
	}
}

func TestCurrentCacheEntry_ToModel(t *testing.T) {
	observedAt := time.Now()
	entry := CurrentCacheEntry{
		TemperatureC: 15.5,
		ObservedAt:   observedAt,
		Timezone:     "America/Detroit",
		CachedAt:     time.Now(),
		TTL:          300,
	}

	model := entry.ToModel()

	if model.TemperatureC() != entry.TemperatureC {
		t.Errorf("TemperatureC = %v, want %v", model.TemperatureC(), entry.TemperatureC)
	}

	if !model.ObservedAt().Equal(observedAt) {
		t.Errorf("ObservedAt = %v, want %v", model.ObservedAt(), observedAt)
	}

	if model.Timezone() != entry.Timezone {
		t.Errorf("Timezone = %v, want %v", model.Timezone(), entry.Timezone)
	}
}

func TestForecastCacheEntry_Serialization(t *testing.T) {
	days := []DailyForecast{
		NewDailyForecast(time.Date(2025, 11, 15, 0, 0, 0, 0, time.UTC), 20.0, 10.0),
		NewDailyForecast(time.Date(2025, 11, 16, 0, 0, 0, 0, time.UTC), 22.0, 12.0),
	}
	forecast := NewForecastWeather(days, time.Now(), "America/Detroit")
	ttl := 1 * time.Hour

	entry := NewForecastCacheEntry(forecast, ttl)

	// Marshal to JSON
	data, err := entry.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshal cache entry: %v", err)
	}

	// Unmarshal from JSON
	unmarshaled, err := UnmarshalForecastCacheEntry(data)
	if err != nil {
		t.Fatalf("Failed to unmarshal cache entry: %v", err)
	}

	// Verify data integrity
	if len(unmarshaled.Days) != len(entry.Days) {
		t.Errorf("Days length = %v, want %v", len(unmarshaled.Days), len(entry.Days))
	}

	if unmarshaled.Timezone != entry.Timezone {
		t.Errorf("Timezone = %v, want %v", unmarshaled.Timezone, entry.Timezone)
	}

	if unmarshaled.TTL != entry.TTL {
		t.Errorf("TTL = %v, want %v", unmarshaled.TTL, entry.TTL)
	}
}

func TestForecastCacheEntry_ToModel(t *testing.T) {
	entry := ForecastCacheEntry{
		Days: []DailyForecastCache{
			{Date: "2025-11-15", TMaxC: 20.0, TMinC: 10.0},
			{Date: "2025-11-16", TMaxC: 22.0, TMinC: 12.0},
		},
		GeneratedAt: time.Now(),
		Timezone:    "America/Detroit",
		CachedAt:    time.Now(),
		TTL:         3600,
	}

	model, err := entry.ToModel()
	if err != nil {
		t.Fatalf("Failed to convert to model: %v", err)
	}

	days := model.Days()
	if len(days) != 2 {
		t.Errorf("Days length = %v, want 2", len(days))
	}

	if days[0].TMaxC() != 20.0 {
		t.Errorf("Day 0 TMaxC = %v, want 20.0", days[0].TMaxC())
	}

	if days[0].TMinC() != 10.0 {
		t.Errorf("Day 0 TMinC = %v, want 10.0", days[0].TMinC())
	}

	if model.Timezone() != "America/Detroit" {
		t.Errorf("Timezone = %v, want America/Detroit", model.Timezone())
	}
}

func TestForecastCacheEntry_IsStale(t *testing.T) {
	tests := []struct {
		name          string
		cachedAt      time.Time
		ttl           time.Duration
		expectedStale bool
	}{
		{
			name:          "Fresh cache",
			cachedAt:      time.Now().Add(-30 * time.Minute),
			ttl:           1 * time.Hour,
			expectedStale: false,
		},
		{
			name:          "Stale cache",
			cachedAt:      time.Now().Add(-2 * time.Hour),
			ttl:           1 * time.Hour,
			expectedStale: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := ForecastCacheEntry{
				Days: []DailyForecastCache{
					{Date: "2025-11-15", TMaxC: 20.0, TMinC: 10.0},
				},
				GeneratedAt: time.Now(),
				Timezone:    "America/Detroit",
				CachedAt:    tt.cachedAt,
				TTL:         int64(tt.ttl.Seconds()),
			}

			if got := entry.IsStale(); got != tt.expectedStale {
				t.Errorf("ForecastCacheEntry.IsStale() = %v, want %v", got, tt.expectedStale)
			}
		})
	}
}

func TestCurrentCacheEntry_Age(t *testing.T) {
	cachedAt := time.Now().Add(-5 * time.Minute)
	entry := CurrentCacheEntry{
		TemperatureC: 15.5,
		ObservedAt:   time.Now(),
		Timezone:     "America/Detroit",
		CachedAt:     cachedAt,
		TTL:          300,
	}

	age := entry.Age()

	if age < 4*time.Minute || age > 6*time.Minute {
		t.Errorf("CurrentCacheEntry.Age() = %v, expected between 4-6 minutes", age)
	}
}

func TestForecastCacheEntry_Age(t *testing.T) {
	cachedAt := time.Now().Add(-10 * time.Minute)
	entry := ForecastCacheEntry{
		Days: []DailyForecastCache{
			{Date: "2025-11-15", TMaxC: 20.0, TMinC: 10.0},
		},
		GeneratedAt: time.Now(),
		Timezone:    "America/Detroit",
		CachedAt:    cachedAt,
		TTL:         3600,
	}

	age := entry.Age()

	if age < 9*time.Minute || age > 11*time.Minute {
		t.Errorf("ForecastCacheEntry.Age() = %v, expected between 9-11 minutes", age)
	}
}
