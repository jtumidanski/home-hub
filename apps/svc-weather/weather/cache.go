package weather

import (
	"encoding/json"
	"time"
)

// CurrentCacheEntry represents a serializable cache entry for current weather
type CurrentCacheEntry struct {
	TemperatureC float64   `json:"temperature_c"`
	ObservedAt   time.Time `json:"observed_at"`
	Timezone     string    `json:"timezone"`
	CachedAt     time.Time `json:"cached_at"`
	TTL          int64     `json:"ttl_seconds"`
}

// NewCurrentCacheEntry creates a cache entry from a domain model
func NewCurrentCacheEntry(current CurrentWeather, ttl time.Duration) CurrentCacheEntry {
	return CurrentCacheEntry{
		TemperatureC: current.TemperatureC(),
		ObservedAt:   current.ObservedAt(),
		Timezone:     current.Timezone(),
		CachedAt:     time.Now(),
		TTL:          int64(ttl.Seconds()),
	}
}

// ToModel converts a cache entry to a domain model
func (c CurrentCacheEntry) ToModel() CurrentWeather {
	return NewCurrentWeather(c.TemperatureC, c.ObservedAt, c.Timezone)
}

// IsStale returns true if the cache entry has exceeded its TTL
func (c CurrentCacheEntry) IsStale() bool {
	expiresAt := c.CachedAt.Add(time.Duration(c.TTL) * time.Second)
	return time.Now().After(expiresAt)
}

// Age returns the age of the cached data
func (c CurrentCacheEntry) Age() time.Duration {
	return time.Since(c.CachedAt)
}

// Marshal serializes the cache entry to JSON
func (c CurrentCacheEntry) Marshal() ([]byte, error) {
	return json.Marshal(c)
}

// UnmarshalCurrentCacheEntry deserializes a cache entry from JSON
func UnmarshalCurrentCacheEntry(data []byte) (CurrentCacheEntry, error) {
	var entry CurrentCacheEntry
	err := json.Unmarshal(data, &entry)
	return entry, err
}

// DailyForecastCache represents a serializable daily forecast
type DailyForecastCache struct {
	Date  string  `json:"date"`
	TMaxC float64 `json:"tmax_c"`
	TMinC float64 `json:"tmin_c"`
}

// ForecastCacheEntry represents a serializable cache entry for forecast weather
type ForecastCacheEntry struct {
	Days        []DailyForecastCache `json:"days"`
	GeneratedAt time.Time            `json:"generated_at"`
	Timezone    string               `json:"timezone"`
	CachedAt    time.Time            `json:"cached_at"`
	TTL         int64                `json:"ttl_seconds"`
}

// NewForecastCacheEntry creates a cache entry from a domain model
func NewForecastCacheEntry(forecast ForecastWeather, ttl time.Duration) ForecastCacheEntry {
	days := make([]DailyForecastCache, len(forecast.Days()))
	for i, day := range forecast.Days() {
		days[i] = DailyForecastCache{
			Date:  day.Date().Format("2006-01-02"),
			TMaxC: day.TMaxC(),
			TMinC: day.TMinC(),
		}
	}

	return ForecastCacheEntry{
		Days:        days,
		GeneratedAt: forecast.GeneratedAt(),
		Timezone:    forecast.Timezone(),
		CachedAt:    time.Now(),
		TTL:         int64(ttl.Seconds()),
	}
}

// ToModel converts a cache entry to a domain model
func (f ForecastCacheEntry) ToModel() (ForecastWeather, error) {
	days := make([]DailyForecast, len(f.Days))
	for i, day := range f.Days {
		date, err := time.Parse("2006-01-02", day.Date)
		if err != nil {
			return ForecastWeather{}, err
		}
		days[i] = NewDailyForecast(date, day.TMaxC, day.TMinC)
	}

	return NewForecastWeather(days, f.GeneratedAt, f.Timezone), nil
}

// IsStale returns true if the cache entry has exceeded its TTL
func (f ForecastCacheEntry) IsStale() bool {
	expiresAt := f.CachedAt.Add(time.Duration(f.TTL) * time.Second)
	return time.Now().After(expiresAt)
}

// Age returns the age of the cached data
func (f ForecastCacheEntry) Age() time.Duration {
	return time.Since(f.CachedAt)
}

// Marshal serializes the cache entry to JSON
func (f ForecastCacheEntry) Marshal() ([]byte, error) {
	return json.Marshal(f)
}

// UnmarshalForecastCacheEntry deserializes a cache entry from JSON
func UnmarshalForecastCacheEntry(data []byte) (ForecastCacheEntry, error) {
	var entry ForecastCacheEntry
	err := json.Unmarshal(data, &entry)
	return entry, err
}
