package weather

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/apps/svc-weather/cache"
	"github.com/jtumidanski/home-hub/apps/svc-weather/geokey"
	"github.com/jtumidanski/home-hub/apps/svc-weather/household"
	"github.com/jtumidanski/home-hub/apps/svc-weather/openmeteo"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/singleflight"
)

// Provider defines the interface for weather operations
type Provider interface {
	GetCombined(ctx context.Context, householdID uuid.UUID) (CombinedWeather, bool, error)
	GetCurrent(ctx context.Context, householdID uuid.UUID) (CurrentWeather, bool, error)
	GetForecast(ctx context.Context, householdID uuid.UUID, days int) (ForecastWeather, bool, error)
	Refresh(ctx context.Context, householdID uuid.UUID) error
	RefreshCurrent(ctx context.Context, householdID uuid.UUID) error
	RefreshForecast(ctx context.Context, householdID uuid.UUID) error
	Purge(ctx context.Context, householdID uuid.UUID) error
	PurgeAll(ctx context.Context) error
}

// CacheProvider implements Provider with cache-only reads
type CacheProvider struct {
	cache            cache.Client
	householdResolver household.Resolver
	meteoClient      openmeteo.Client
	geoGen           geokey.Generator
	keyBuilder       geokey.KeyBuilder
	currentTTL       time.Duration
	forecastTTL      time.Duration
	staleMax         time.Duration
	logger           logrus.FieldLogger
	sf               singleflight.Group
}

// NewCacheProvider creates a new weather provider
func NewCacheProvider(
	cacheClient cache.Client,
	householdResolver household.Resolver,
	meteoClient openmeteo.Client,
	geoGen geokey.Generator,
	keyBuilder geokey.KeyBuilder,
	currentTTL time.Duration,
	forecastTTL time.Duration,
	staleMax time.Duration,
	logger logrus.FieldLogger,
) *CacheProvider {
	return &CacheProvider{
		cache:            cacheClient,
		householdResolver: householdResolver,
		meteoClient:      meteoClient,
		geoGen:           geoGen,
		keyBuilder:       keyBuilder,
		currentTTL:       currentTTL,
		forecastTTL:      forecastTTL,
		staleMax:         staleMax,
		logger:           logger,
	}
}

// GetCombined retrieves both current and forecast weather
func (p *CacheProvider) GetCombined(ctx context.Context, householdID uuid.UUID) (CombinedWeather, bool, error) {
	location, gk, err := p.resolveLocation(ctx, householdID)
	if err != nil {
		return CombinedWeather{}, false, err
	}

	current, currentStale, currentErr := p.getCurrent(ctx, gk)
	forecast, forecastStale, forecastErr := p.getForecast(ctx, gk, 7)

	// Log when one or both fail to help debug inconsistent data
	if currentErr != nil || forecastErr != nil {
		p.logger.WithFields(logrus.Fields{
			"household_id":   householdID,
			"geokey":         gk,
			"current_error":  currentErr != nil,
			"forecast_error": forecastErr != nil,
		}).Warn("GetCombined: partial data available")
	}

	// If both failed, return error
	if currentErr != nil && forecastErr != nil {
		return CombinedWeather{}, false, fmt.Errorf("no weather data available")
	}

	// Determine if combined data is stale
	stale := currentStale || forecastStale

	meta := NewWeatherMeta("open-meteo", gk, time.Now(), location.Timezone)

	var currentPtr *CurrentWeather
	if currentErr == nil {
		currentPtr = &current
	} else {
		p.logger.WithFields(logrus.Fields{
			"household_id": householdID,
			"geokey":       gk,
			"error":        currentErr.Error(),
		}).Debug("Current weather unavailable in GetCombined")
	}

	var forecastPtr *ForecastWeather
	if forecastErr == nil {
		forecastPtr = &forecast
	} else {
		p.logger.WithFields(logrus.Fields{
			"household_id": householdID,
			"geokey":       gk,
			"error":        forecastErr.Error(),
		}).Debug("Forecast weather unavailable in GetCombined")
	}

	return NewCombinedWeather(currentPtr, forecastPtr, meta), stale, nil
}

// GetCurrent retrieves current weather from cache
func (p *CacheProvider) GetCurrent(ctx context.Context, householdID uuid.UUID) (CurrentWeather, bool, error) {
	_, gk, err := p.resolveLocation(ctx, householdID)
	if err != nil {
		return CurrentWeather{}, false, err
	}

	return p.getCurrent(ctx, gk)
}

// GetForecast retrieves forecast weather from cache
func (p *CacheProvider) GetForecast(ctx context.Context, householdID uuid.UUID, days int) (ForecastWeather, bool, error) {
	_, gk, err := p.resolveLocation(ctx, householdID)
	if err != nil {
		return ForecastWeather{}, false, err
	}

	return p.getForecast(ctx, gk, days)
}

// Refresh triggers a manual refresh of weather data (both current and forecast)
func (p *CacheProvider) Refresh(ctx context.Context, householdID uuid.UUID) error {
	location, gk, err := p.resolveLocation(ctx, householdID)
	if err != nil {
		return err
	}

	p.logger.WithFields(logrus.Fields{
		"household_id": householdID,
		"geokey":       gk,
	}).Info("Refreshing weather data")

	// Refresh current and forecast in parallel
	errCh := make(chan error, 2)

	go func() {
		errCh <- p.refreshCurrent(ctx, location.Latitude, location.Longitude, gk)
	}()

	go func() {
		errCh <- p.refreshForecast(ctx, location.Latitude, location.Longitude, gk, 7)
	}()

	// Wait for both to complete
	var errs []error
	for i := 0; i < 2; i++ {
		if err := <-errCh; err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("refresh failed: %v", errs)
	}

	return nil
}

// RefreshCurrent triggers a manual refresh of current weather only
func (p *CacheProvider) RefreshCurrent(ctx context.Context, householdID uuid.UUID) error {
	location, gk, err := p.resolveLocation(ctx, householdID)
	if err != nil {
		return err
	}

	p.logger.WithFields(logrus.Fields{
		"household_id": householdID,
		"geokey":       gk,
	}).Debug("Refreshing current weather")

	return p.refreshCurrent(ctx, location.Latitude, location.Longitude, gk)
}

// RefreshForecast triggers a manual refresh of forecast weather only
func (p *CacheProvider) RefreshForecast(ctx context.Context, householdID uuid.UUID) error {
	location, gk, err := p.resolveLocation(ctx, householdID)
	if err != nil {
		return err
	}

	p.logger.WithFields(logrus.Fields{
		"household_id": householdID,
		"geokey":       gk,
	}).Debug("Refreshing forecast weather")

	return p.refreshForecast(ctx, location.Latitude, location.Longitude, gk, 7)
}

// Purge removes cached weather data for a household
func (p *CacheProvider) Purge(ctx context.Context, householdID uuid.UUID) error {
	_, gk, err := p.resolveLocation(ctx, householdID)
	if err != nil {
		return err
	}

	p.logger.WithFields(logrus.Fields{
		"household_id": householdID,
		"geokey":       gk,
	}).Info("Purging weather cache")

	currentKey := p.keyBuilder.CurrentKey(gk)
	forecastKey := p.keyBuilder.ForecastKey(gk, 7)

	if err := p.cache.Delete(ctx, currentKey); err != nil {
		return err
	}

	if err := p.cache.Delete(ctx, forecastKey); err != nil {
		return err
	}

	return nil
}

// PurgeAll removes all cached weather data
func (p *CacheProvider) PurgeAll(ctx context.Context) error {
	p.logger.Warn("Purging all weather cache")
	return p.cache.DeletePattern(ctx, p.keyBuilder.Pattern())
}

// resolveLocation resolves household location and generates geokey
func (p *CacheProvider) resolveLocation(ctx context.Context, householdID uuid.UUID) (household.Location, string, error) {
	location, err := p.householdResolver.Resolve(ctx, householdID)
	if err != nil {
		return household.Location{}, "", err
	}

	gk := p.geoGen.Generate(location.Latitude, location.Longitude)
	return location, gk, nil
}

// getCurrent retrieves current weather from cache (cache-only)
func (p *CacheProvider) getCurrent(ctx context.Context, gk string) (CurrentWeather, bool, error) {
	key := p.keyBuilder.CurrentKey(gk)

	data, err := p.cache.Get(ctx, key)
	if err != nil {
		if cache.IsCacheMiss(err) {
			return CurrentWeather{}, false, fmt.Errorf("current weather not available (cache miss)")
		}
		return CurrentWeather{}, false, err
	}

	entry, err := UnmarshalCurrentCacheEntry(data)
	if err != nil {
		return CurrentWeather{}, false, err
	}

	// Check if stale
	stale := entry.IsStale()

	// If too old (> staleMax), return error
	if entry.Age() > p.staleMax {
		return CurrentWeather{}, false, fmt.Errorf("current weather too stale (age: %s)", entry.Age())
	}

	return entry.ToModel(), stale, nil
}

// getForecast retrieves forecast weather from cache (cache-only)
func (p *CacheProvider) getForecast(ctx context.Context, gk string, days int) (ForecastWeather, bool, error) {
	key := p.keyBuilder.ForecastKey(gk, days)

	data, err := p.cache.Get(ctx, key)
	if err != nil {
		if cache.IsCacheMiss(err) {
			return ForecastWeather{}, false, fmt.Errorf("forecast weather not available (cache miss)")
		}
		return ForecastWeather{}, false, err
	}

	entry, err := UnmarshalForecastCacheEntry(data)
	if err != nil {
		return ForecastWeather{}, false, err
	}

	// Check if stale
	stale := entry.IsStale()

	// If too old (> staleMax), return error
	if entry.Age() > p.staleMax {
		return ForecastWeather{}, false, fmt.Errorf("forecast weather too stale (age: %s)", entry.Age())
	}

	forecast, err := entry.ToModel()
	if err != nil {
		return ForecastWeather{}, false, err
	}

	return forecast, stale, nil
}

// refreshCurrent fetches and caches current weather (used by scheduler)
func (p *CacheProvider) refreshCurrent(ctx context.Context, lat, lon float64, gk string) error {
	// Use singleflight to prevent duplicate refreshes
	_, err, _ := p.sf.Do(fmt.Sprintf("current:%s", gk), func() (interface{}, error) {
		response, err := p.meteoClient.FetchCurrent(ctx, lat, lon)
		if err != nil {
			return nil, err
		}

		// Convert Open-Meteo response to domain model
		current, err := convertCurrentResponse(response)
		if err != nil {
			return nil, err
		}

		entry := NewCurrentCacheEntry(current, p.currentTTL)
		data, err := entry.Marshal()
		if err != nil {
			return nil, err
		}

		key := p.keyBuilder.CurrentKey(gk)
		return nil, p.cache.Set(ctx, key, data, p.currentTTL)
	})

	return err
}

// refreshForecast fetches and caches forecast weather (used by scheduler)
func (p *CacheProvider) refreshForecast(ctx context.Context, lat, lon float64, gk string, days int) error {
	// Use singleflight to prevent duplicate refreshes
	_, err, _ := p.sf.Do(fmt.Sprintf("forecast:%s:%d", gk, days), func() (interface{}, error) {
		response, err := p.meteoClient.FetchForecast(ctx, lat, lon, days)
		if err != nil {
			return nil, err
		}

		// Convert Open-Meteo response to domain model
		forecast, err := convertForecastResponse(response)
		if err != nil {
			return nil, err
		}

		entry := NewForecastCacheEntry(forecast, p.forecastTTL)
		data, err := entry.Marshal()
		if err != nil {
			return nil, err
		}

		key := p.keyBuilder.ForecastKey(gk, days)
		return nil, p.cache.Set(ctx, key, data, p.forecastTTL)
	})

	return err
}

// convertCurrentResponse converts Open-Meteo current response to domain model
func convertCurrentResponse(response openmeteo.CurrentResponse) (CurrentWeather, error) {
	// Load the timezone from the response
	loc, err := time.LoadLocation(response.Timezone)
	if err != nil {
		return CurrentWeather{}, fmt.Errorf("failed to load timezone %s: %w", response.Timezone, err)
	}

	// Parse the time string in the location's timezone
	observedAt, err := time.ParseInLocation("2006-01-02T15:04", response.Current.Time, loc)
	if err != nil {
		return CurrentWeather{}, fmt.Errorf("failed to parse current time: %w", err)
	}

	current := NewCurrentWeather(
		response.Current.Temperature2m,
		observedAt,
		response.Timezone,
	)

	// Debug logging to trace the observation time
	logrus.WithFields(logrus.Fields{
		"raw_time_string":     response.Current.Time,
		"parsed_observed_at":  observedAt.Format(time.RFC3339),
		"observation_age_sec": int64(current.Age().Seconds()),
		"timezone":            response.Timezone,
		"temperature":         response.Current.Temperature2m,
	}).Info("Converted OpenMeteo current weather response")

	return current, nil
}

// convertForecastResponse converts Open-Meteo forecast response to domain model
func convertForecastResponse(response openmeteo.ForecastResponse) (ForecastWeather, error) {
	if len(response.Daily.Time) == 0 {
		return ForecastWeather{}, fmt.Errorf("no forecast data in response")
	}

	if len(response.Daily.Time) != len(response.Daily.Temperature2mMax) ||
		len(response.Daily.Time) != len(response.Daily.Temperature2mMin) {
		return ForecastWeather{}, fmt.Errorf("mismatched forecast data arrays")
	}

	days := make([]DailyForecast, len(response.Daily.Time))
	for i, dateStr := range response.Daily.Time {
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return ForecastWeather{}, fmt.Errorf("failed to parse date %s: %w", dateStr, err)
		}

		days[i] = NewDailyForecast(
			date,
			response.Daily.Temperature2mMax[i],
			response.Daily.Temperature2mMin[i],
		)
	}

	return NewForecastWeather(
		days,
		time.Now(), // Use current time as generation time
		response.Timezone,
	), nil
}
