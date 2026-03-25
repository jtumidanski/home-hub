package forecast

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/weather-service/internal/openmeteo"
	"github.com/jtumidanski/home-hub/services/weather-service/internal/weathercode"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var ErrNoLocation = errors.New("household has no location configured")

type Processor struct {
	l      logrus.FieldLogger
	ctx    context.Context
	db     *gorm.DB
	client *openmeteo.Client
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB, client *openmeteo.Client) *Processor {
	return &Processor{l: l, ctx: ctx, db: db, client: client}
}

func (p *Processor) GetCurrent(tenantID, householdID uuid.UUID, lat, lon float64, units, timezone string) (Model, error) {
	cache, err := p.getOrFetch(tenantID, householdID, lat, lon, units, timezone)
	if err != nil {
		return Model{}, err
	}
	return cache, nil
}

func (p *Processor) GetForecast(tenantID, householdID uuid.UUID, lat, lon float64, units, timezone string) (Model, error) {
	return p.getOrFetch(tenantID, householdID, lat, lon, units, timezone)
}

func (p *Processor) getOrFetch(tenantID, householdID uuid.UUID, lat, lon float64, units, timezone string) (Model, error) {
	e, err := getByHouseholdID(householdID)(p.db.WithContext(p.ctx))()
	if err == nil {
		// Check if cached coordinates/units still match
		if e.Latitude == lat && e.Longitude == lon && e.Units == units {
			m, err := Make(e)
			if err == nil {
				return m, nil
			}
		}
		// Stale cache — coordinates or units changed, re-fetch
	}

	return p.fetchAndCache(tenantID, householdID, lat, lon, units, timezone)
}

func (p *Processor) fetchAndCache(tenantID, householdID uuid.UUID, lat, lon float64, units, timezone string) (Model, error) {
	resp, err := p.client.FetchForecast(lat, lon, units, timezone)
	if err != nil {
		return Model{}, err
	}

	currentSummary, currentIcon := weathercode.Lookup(resp.Current.WeatherCode)
	current := CurrentData{
		Temperature: resp.Current.Temperature,
		WeatherCode: resp.Current.WeatherCode,
		Summary:     currentSummary,
		Icon:        currentIcon,
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
		}
	}

	e, err := upsert(p.db.WithContext(p.ctx), tenantID, householdID, lat, lon, units, current, daily)
	if err != nil {
		return Model{}, err
	}

	return Make(e)
}

func (p *Processor) RefreshCache(e Entity) error {
	// Use stored timezone or default to UTC
	timezone := "UTC"

	resp, err := p.client.FetchForecast(e.Latitude, e.Longitude, e.Units, timezone)
	if err != nil {
		return err
	}

	currentSummary, currentIcon := weathercode.Lookup(resp.Current.WeatherCode)
	current := CurrentData{
		Temperature: resp.Current.Temperature,
		WeatherCode: resp.Current.WeatherCode,
		Summary:     currentSummary,
		Icon:        currentIcon,
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
		}
	}

	_, err = upsert(p.db.WithContext(p.ctx), e.TenantId, e.HouseholdId, e.Latitude, e.Longitude, e.Units, current, daily)
	return err
}

func (p *Processor) InvalidateCache(householdID uuid.UUID) error {
	return deleteByHouseholdID(p.db.WithContext(p.ctx), householdID)
}

func (p *Processor) AllCacheEntries() ([]Entity, error) {
	return getAll()(p.db.WithContext(p.ctx))()
}
