package weather

import (
	"time"
)

// CurrentWeather represents current temperature data for a location
type CurrentWeather struct {
	temperatureC float64
	observedAt   time.Time
	timezone     string
}

// NewCurrentWeather creates a new CurrentWeather instance
func NewCurrentWeather(temperatureC float64, observedAt time.Time, timezone string) CurrentWeather {
	return CurrentWeather{
		temperatureC: temperatureC,
		observedAt:   observedAt,
		timezone:     timezone,
	}
}

// TemperatureC returns the temperature in Celsius
func (c CurrentWeather) TemperatureC() float64 {
	return c.temperatureC
}

// ObservedAt returns when this temperature was observed
func (c CurrentWeather) ObservedAt() time.Time {
	return c.observedAt
}

// Timezone returns the timezone for this observation
func (c CurrentWeather) Timezone() string {
	return c.timezone
}

// Age returns how old this observation is
func (c CurrentWeather) Age() time.Duration {
	return time.Since(c.observedAt)
}

// IsStale returns true if the observation is older than the given TTL
func (c CurrentWeather) IsStale(ttl time.Duration) bool {
	return c.Age() > ttl
}

// DailyForecast represents min/max temperatures for a single day
type DailyForecast struct {
	date  time.Time
	tmaxC float64
	tminC float64
}

// NewDailyForecast creates a new DailyForecast instance
func NewDailyForecast(date time.Time, tmaxC, tminC float64) DailyForecast {
	return DailyForecast{
		date:  date,
		tmaxC: tmaxC,
		tminC: tminC,
	}
}

// Date returns the forecast date
func (d DailyForecast) Date() time.Time {
	return d.date
}

// TMaxC returns the maximum temperature in Celsius
func (d DailyForecast) TMaxC() float64 {
	return d.tmaxC
}

// TMinC returns the minimum temperature in Celsius
func (d DailyForecast) TMinC() float64 {
	return d.tminC
}

// ForecastWeather represents a multi-day forecast
type ForecastWeather struct {
	days        []DailyForecast
	generatedAt time.Time
	timezone    string
}

// NewForecastWeather creates a new ForecastWeather instance
func NewForecastWeather(days []DailyForecast, generatedAt time.Time, timezone string) ForecastWeather {
	return ForecastWeather{
		days:        days,
		generatedAt: generatedAt,
		timezone:    timezone,
	}
}

// Days returns the daily forecasts
func (f ForecastWeather) Days() []DailyForecast {
	// Return a copy to maintain immutability
	result := make([]DailyForecast, len(f.days))
	copy(result, f.days)
	return result
}

// GeneratedAt returns when this forecast was generated
func (f ForecastWeather) GeneratedAt() time.Time {
	return f.generatedAt
}

// Timezone returns the timezone for this forecast
func (f ForecastWeather) Timezone() string {
	return f.timezone
}

// Age returns how old this forecast is
func (f ForecastWeather) Age() time.Duration {
	return time.Since(f.generatedAt)
}

// IsStale returns true if the forecast is older than the given TTL
func (f ForecastWeather) IsStale(ttl time.Duration) bool {
	return f.Age() > ttl
}

// WeatherMeta contains metadata about weather data
type WeatherMeta struct {
	source      string
	geokey      string
	refreshedAt time.Time
	timezone    string
}

// NewWeatherMeta creates a new WeatherMeta instance
func NewWeatherMeta(source, geokey string, refreshedAt time.Time, timezone string) WeatherMeta {
	return WeatherMeta{
		source:      source,
		geokey:      geokey,
		refreshedAt: refreshedAt,
		timezone:    timezone,
	}
}

// Source returns the weather data source
func (m WeatherMeta) Source() string {
	return m.source
}

// Geokey returns the geohash key for spatial de-duplication
func (m WeatherMeta) Geokey() string {
	return m.geokey
}

// RefreshedAt returns when this data was last refreshed
func (m WeatherMeta) RefreshedAt() time.Time {
	return m.refreshedAt
}

// Timezone returns the timezone
func (m WeatherMeta) Timezone() string {
	return m.timezone
}

// CombinedWeather represents current weather + forecast together
type CombinedWeather struct {
	current  *CurrentWeather
	forecast *ForecastWeather
	meta     WeatherMeta
}

// NewCombinedWeather creates a new CombinedWeather instance
func NewCombinedWeather(current *CurrentWeather, forecast *ForecastWeather, meta WeatherMeta) CombinedWeather {
	return CombinedWeather{
		current:  current,
		forecast: forecast,
		meta:     meta,
	}
}

// Current returns the current weather (may be nil if unavailable)
func (c CombinedWeather) Current() *CurrentWeather {
	return c.current
}

// Forecast returns the forecast weather (may be nil if unavailable)
func (c CombinedWeather) Forecast() *ForecastWeather {
	return c.forecast
}

// Meta returns the metadata
func (c CombinedWeather) Meta() WeatherMeta {
	return c.meta
}

// HasCurrent returns true if current weather is available
func (c CombinedWeather) HasCurrent() bool {
	return c.current != nil
}

// HasForecast returns true if forecast weather is available
func (c CombinedWeather) HasForecast() bool {
	return c.forecast != nil
}
