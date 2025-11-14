package weather

import (
	"github.com/google/uuid"
	"time"
)

// RestModel represents the JSON:API representation of combined weather
type RestModel struct {
	Id         uuid.UUID              `json:"-"`
	Current    *CurrentAttributes     `json:"current,omitempty"`
	Daily      []DailyAttributes      `json:"daily,omitempty"`
	Units      string                 `json:"units"`
	Meta       MetaAttributes         `json:"meta"`
	Stale      bool                   `json:"stale"`
}

// CurrentAttributes represents current weather in JSON:API format
type CurrentAttributes struct {
	TemperatureC float64 `json:"temperature_c"`
	ObservedAt   string  `json:"observed_at"`
	Stale        bool    `json:"stale"`
	AgeSeconds   int64   `json:"age_seconds"`
}

// DailyAttributes represents a single day forecast in JSON:API format
type DailyAttributes struct {
	Date  string  `json:"date"`
	TMaxC float64 `json:"tmax_c"`
	TMinC float64 `json:"tmin_c"`
}

// MetaAttributes contains metadata in JSON:API format
type MetaAttributes struct {
	Source      string `json:"source"`
	Timezone    string `json:"timezone"`
	Geokey      string `json:"geokey"`
	RefreshedAt string `json:"refreshed_at"`
}

// GetName returns the JSON:API resource name
func (r RestModel) GetName() string {
	return "weather"
}

// GetID returns the JSON:API resource ID
func (r RestModel) GetID() string {
	return r.Id.String()
}

// SetID sets the JSON:API resource ID
func (r *RestModel) SetID(idStr string) error {
	if idStr == "" {
		r.Id = uuid.Nil
		return nil
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return err
	}

	r.Id = id
	return nil
}

// Transform converts a CombinedWeather domain model to JSON:API format
func Transform(householdID uuid.UUID, combined CombinedWeather, stale bool) RestModel {
	model := RestModel{
		Id:    householdID,
		Units: "celsius",
		Stale: stale,
		Meta: MetaAttributes{
			Source:      combined.Meta().Source(),
			Timezone:    combined.Meta().Timezone(),
			Geokey:      combined.Meta().Geokey(),
			RefreshedAt: combined.Meta().RefreshedAt().Format(time.RFC3339),
		},
	}

	// Add current weather if available
	if combined.HasCurrent() {
		current := combined.Current()
		model.Current = &CurrentAttributes{
			TemperatureC: current.TemperatureC(),
			ObservedAt:   current.ObservedAt().Format(time.RFC3339),
			Stale:        current.IsStale(5 * time.Minute), // TODO: Use actual TTL
			AgeSeconds:   int64(current.Age().Seconds()),
		}
	}

	// Add forecast if available
	if combined.HasForecast() {
		forecast := combined.Forecast()
		days := forecast.Days()
		model.Daily = make([]DailyAttributes, len(days))

		for i, day := range days {
			model.Daily[i] = DailyAttributes{
				Date:  day.Date().Format("2006-01-02"),
				TMaxC: day.TMaxC(),
				TMinC: day.TMinC(),
			}
		}
	}

	return model
}

// TransformCurrent converts a CurrentWeather domain model to JSON:API format
func TransformCurrent(householdID uuid.UUID, current CurrentWeather, stale bool, geokey, timezone string) RestModel {
	return RestModel{
		Id:    householdID,
		Units: "celsius",
		Stale: stale,
		Current: &CurrentAttributes{
			TemperatureC: current.TemperatureC(),
			ObservedAt:   current.ObservedAt().Format(time.RFC3339),
			Stale:        stale,
			AgeSeconds:   int64(current.Age().Seconds()),
		},
		Meta: MetaAttributes{
			Source:      "open-meteo",
			Timezone:    timezone,
			Geokey:      geokey,
			RefreshedAt: time.Now().Format(time.RFC3339),
		},
	}
}

// TransformForecast converts a ForecastWeather domain model to JSON:API format
func TransformForecast(householdID uuid.UUID, forecast ForecastWeather, stale bool, geokey, timezone string) RestModel {
	days := forecast.Days()
	dailyAttrs := make([]DailyAttributes, len(days))

	for i, day := range days {
		dailyAttrs[i] = DailyAttributes{
			Date:  day.Date().Format("2006-01-02"),
			TMaxC: day.TMaxC(),
			TMinC: day.TMinC(),
		}
	}

	return RestModel{
		Id:    householdID,
		Units: "celsius",
		Stale: stale,
		Daily: dailyAttrs,
		Meta: MetaAttributes{
			Source:      "open-meteo",
			Timezone:    timezone,
			Geokey:      geokey,
			RefreshedAt: forecast.GeneratedAt().Format(time.RFC3339),
		},
	}
}
