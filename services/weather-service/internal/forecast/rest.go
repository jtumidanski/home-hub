package forecast

import (
	"time"

	"github.com/google/uuid"
)

type RestModel struct {
	Id              uuid.UUID `json:"-"`
	Temperature     float64   `json:"temperature"`
	TemperatureUnit string    `json:"temperatureUnit"`
	Summary         string    `json:"summary"`
	Icon            string    `json:"icon"`
	WeatherCode     int       `json:"weatherCode"`
	HighTemperature float64   `json:"highTemperature"`
	LowTemperature  float64   `json:"lowTemperature"`
	FetchedAt       time.Time `json:"fetchedAt"`
}

func (r RestModel) GetName() string { return "weather-current" }
func (r RestModel) GetID() string   { return r.Id.String() }
func (r *RestModel) SetID(id string) error {
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

func Transform(m Model) (RestModel, error) {
	tempUnit := m.TemperatureUnit()
	current := m.CurrentData()

	var highTemp, lowTemp float64
	if len(m.ForecastData()) > 0 {
		today := m.ForecastData()[0]
		highTemp = today.HighTemperature
		lowTemp = today.LowTemperature
	}

	return RestModel{
		Id:              m.HouseholdID(),
		Temperature:     current.Temperature,
		TemperatureUnit: tempUnit,
		Summary:         current.Summary,
		Icon:            current.Icon,
		WeatherCode:     current.WeatherCode,
		HighTemperature: highTemp,
		LowTemperature:  lowTemp,
		FetchedAt:       m.FetchedAt(),
	}, nil
}

func TransformSlice(models []Model) ([]RestModel, error) {
	result := make([]RestModel, len(models))
	for i, m := range models {
		rm, err := Transform(m)
		if err != nil {
			return nil, err
		}
		result[i] = rm
	}
	return result, nil
}

type HourlyRestModel struct {
	Time                     string  `json:"time"`
	Temperature              float64 `json:"temperature"`
	WeatherCode              int     `json:"weatherCode"`
	Summary                  string  `json:"summary"`
	Icon                     string  `json:"icon"`
	PrecipitationProbability int     `json:"precipitationProbability"`
}

type DailyRestModel struct {
	Id              string            `json:"-"`
	Date            string            `json:"date"`
	HighTemperature float64           `json:"highTemperature"`
	LowTemperature  float64           `json:"lowTemperature"`
	TemperatureUnit string            `json:"temperatureUnit"`
	Summary         string            `json:"summary"`
	Icon            string            `json:"icon"`
	WeatherCode     int               `json:"weatherCode"`
	HourlyForecast  []HourlyRestModel `json:"hourlyForecast"`
}

func (r DailyRestModel) GetName() string { return "weather-daily" }
func (r DailyRestModel) GetID() string   { return r.Id }
func (r *DailyRestModel) SetID(id string) error {
	r.Id = id
	return nil
}

// TransformDaily converts a single forecast Model into the per-day REST slice.
// This is forecast-specific (one Model fans out into N daily entries) and is
// distinct from the canonical Transform/TransformSlice pair above.
func TransformDaily(m Model) ([]DailyRestModel, error) {
	tempUnit := m.TemperatureUnit()
	result := make([]DailyRestModel, len(m.ForecastData()))
	for i, d := range m.ForecastData() {
		hourly := make([]HourlyRestModel, len(d.HourlyForecast))
		for j, h := range d.HourlyForecast {
			hourly[j] = HourlyRestModel{
				Time:                     h.Time,
				Temperature:              h.Temperature,
				WeatherCode:              h.WeatherCode,
				Summary:                  h.Summary,
				Icon:                     h.Icon,
				PrecipitationProbability: h.PrecipitationProbability,
			}
		}
		result[i] = DailyRestModel{
			Id:              d.Date,
			Date:            d.Date,
			HighTemperature: d.HighTemperature,
			LowTemperature:  d.LowTemperature,
			TemperatureUnit: tempUnit,
			Summary:         d.Summary,
			Icon:            d.Icon,
			WeatherCode:     d.WeatherCode,
			HourlyForecast:  hourly,
		}
	}
	return result, nil
}
