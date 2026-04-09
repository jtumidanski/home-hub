package forecast

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/weather-service/internal/openmeteo"
	"github.com/jtumidanski/home-hub/services/weather-service/internal/weathercode"
	"github.com/jtumidanski/home-hub/shared/go/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var ErrNoLocation = errors.New("household has no location configured")

type Processor struct {
	l        logrus.FieldLogger
	ctx      context.Context
	db       *gorm.DB
	client   *openmeteo.Client
	cacheTTL time.Duration
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB, client *openmeteo.Client, cacheTTL time.Duration) *Processor {
	return &Processor{l: l, ctx: ctx, db: db, client: client, cacheTTL: cacheTTL}
}

func (p *Processor) ByHouseholdAndLocationProvider(householdID uuid.UUID, locationID *uuid.UUID) model.Provider[Model] {
	return model.Map(Make)(getByHouseholdAndLocation(householdID, locationID)(p.db.WithContext(p.ctx)))
}

func (p *Processor) AllProvider() model.Provider[[]Model] {
	return model.SliceMap(Make)(getAll()(p.db.WithContext(p.ctx)))
}

func (p *Processor) GetCurrent(tenantID, householdID uuid.UUID, locationID *uuid.UUID, lat, lon float64, units, timezone string) (Model, error) {
	return p.getOrFetch(tenantID, householdID, locationID, lat, lon, units, timezone)
}

func (p *Processor) GetForecast(tenantID, householdID uuid.UUID, locationID *uuid.UUID, lat, lon float64, units, timezone string) (Model, error) {
	return p.getOrFetch(tenantID, householdID, locationID, lat, lon, units, timezone)
}

// WarmLocationCache implements locationofinterest.CacheWarmer. It synchronously
// fetches and caches the forecast for a newly-created saved location so the
// first user view does not show a loading state.
func (p *Processor) WarmLocationCache(tenantID, householdID, locationID uuid.UUID, lat, lon float64) error {
	_, err := p.fetchAndCache(tenantID, householdID, &locationID, lat, lon, "metric", "UTC")
	return err
}

func (p *Processor) getOrFetch(tenantID, householdID uuid.UUID, locationID *uuid.UUID, lat, lon float64, units, timezone string) (Model, error) {
	m, err := p.ByHouseholdAndLocationProvider(householdID, locationID)()
	if err == nil {
		// Check if cached coordinates/units still match
		if m.Latitude() == lat && m.Longitude() == lon && m.Units() == units {
			// Treat entries without hourly data as stale
			if len(m.ForecastData()) > 0 && len(m.ForecastData()[0].HourlyForecast) > 0 {
				if p.cacheTTL <= 0 || time.Since(m.FetchedAt()) < p.cacheTTL {
					return m, nil
				}
			}
		}
		// Stale cache — coordinates, units changed, missing hourly data, or TTL expired, re-fetch
	}

	return p.fetchAndCache(tenantID, householdID, locationID, lat, lon, units, timezone)
}

func (p *Processor) fetchAndCache(tenantID, householdID uuid.UUID, locationID *uuid.UUID, lat, lon float64, units, timezone string) (Model, error) {
	resp, err := p.client.FetchForecast(lat, lon, units, timezone)
	if err != nil {
		return Model{}, err
	}

	current, daily := transformResponse(resp)

	e, err := create(p.db.WithContext(p.ctx), tenantID, householdID, locationID, lat, lon, units, current, daily)
	if err != nil {
		return Model{}, err
	}

	return Make(e)
}

func (p *Processor) RefreshCache(m Model) error {
	// Use stored timezone or default to UTC
	timezone := "UTC"

	resp, err := p.client.FetchForecast(m.Latitude(), m.Longitude(), m.Units(), timezone)
	if err != nil {
		return err
	}

	current, daily := transformResponse(resp)

	_, err = create(p.db.WithContext(p.ctx), m.TenantID(), m.HouseholdID(), m.LocationID(), m.Latitude(), m.Longitude(), m.Units(), current, daily)
	return err
}

func (p *Processor) InvalidateCache(householdID uuid.UUID) error {
	return deleteByHouseholdID(p.db.WithContext(p.ctx), householdID)
}

func transformResponse(resp *openmeteo.ForecastResponse) (CurrentData, []DailyForecast) {
	currentSummary, currentIcon := weathercode.Lookup(resp.Current.WeatherCode)
	current := CurrentData{
		Temperature: resp.Current.Temperature,
		WeatherCode: resp.Current.WeatherCode,
		Summary:     currentSummary,
		Icon:        currentIcon,
	}

	// Group hourly data by date prefix
	hourlyByDate := make(map[string][]HourlyForecast)
	for i := range resp.Hourly.Time {
		t := resp.Hourly.Time[i]
		date := t[:10] // "2026-03-25T00:00" → "2026-03-25"
		summary, icon := weathercode.Lookup(resp.Hourly.WeatherCode[i])
		hourlyByDate[date] = append(hourlyByDate[date], HourlyForecast{
			Time:                     t,
			Temperature:              resp.Hourly.Temperature[i],
			WeatherCode:              resp.Hourly.WeatherCode[i],
			Summary:                  summary,
			Icon:                     icon,
			PrecipitationProbability: resp.Hourly.PrecipitationProbability[i],
		})
	}

	daily := make([]DailyForecast, len(resp.Daily.Time))
	for i := range resp.Daily.Time {
		summary, icon := weathercode.Lookup(resp.Daily.WeatherCode[i])
		daily[i] = DailyForecast{
			Date:            resp.Daily.Time[i],
			HighTemperature: resp.Daily.TemperatureMax[i],
			LowTemperature:  resp.Daily.TemperatureMin[i],
			WeatherCode:     resp.Daily.WeatherCode[i],
			Summary:         summary,
			Icon:            icon,
			HourlyForecast:  hourlyByDate[resp.Daily.Time[i]],
		}
	}

	return current, daily
}
