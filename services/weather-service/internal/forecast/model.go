package forecast

import (
	"time"

	"github.com/google/uuid"
)

type CurrentData struct {
	Temperature float64 `json:"temperature"`
	WeatherCode int     `json:"weatherCode"`
	Summary     string  `json:"summary"`
	Icon        string  `json:"icon"`
}

type DailyForecast struct {
	Date            string  `json:"date"`
	HighTemperature float64 `json:"highTemperature"`
	LowTemperature  float64 `json:"lowTemperature"`
	WeatherCode     int     `json:"weatherCode"`
	Summary         string  `json:"summary"`
	Icon            string  `json:"icon"`
}

type Model struct {
	id           uuid.UUID
	tenantID     uuid.UUID
	householdID  uuid.UUID
	latitude     float64
	longitude    float64
	units        string
	currentData  CurrentData
	forecastData []DailyForecast
	fetchedAt    time.Time
	createdAt    time.Time
	updatedAt    time.Time
}

func (m Model) Id() uuid.UUID              { return m.id }
func (m Model) TenantID() uuid.UUID        { return m.tenantID }
func (m Model) HouseholdID() uuid.UUID     { return m.householdID }
func (m Model) Latitude() float64          { return m.latitude }
func (m Model) Longitude() float64         { return m.longitude }
func (m Model) Units() string              { return m.units }
func (m Model) CurrentData() CurrentData   { return m.currentData }
func (m Model) ForecastData() []DailyForecast { return m.forecastData }
func (m Model) FetchedAt() time.Time       { return m.fetchedAt }
func (m Model) CreatedAt() time.Time       { return m.createdAt }
func (m Model) UpdatedAt() time.Time       { return m.updatedAt }

func (m Model) TemperatureUnit() string {
	if m.units == "imperial" {
		return "°F"
	}
	return "°C"
}
