package forecast

import (
	"time"

	"github.com/google/uuid"
)

type CurrentRestModel struct {
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

func (r CurrentRestModel) GetName() string { return "weather-current" }
func (r CurrentRestModel) GetID() string   { return r.Id.String() }
func (r *CurrentRestModel) SetID(id string) error {
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

func TransformCurrent(m Model) CurrentRestModel {
	tempUnit := m.TemperatureUnit()
	current := m.CurrentData()

	var highTemp, lowTemp float64
	if len(m.ForecastData()) > 0 {
		today := m.ForecastData()[0]
		highTemp = today.HighTemperature
		lowTemp = today.LowTemperature
	}

	return CurrentRestModel{
		Id:              m.HouseholdID(),
		Temperature:     current.Temperature,
		TemperatureUnit: tempUnit,
		Summary:         current.Summary,
		Icon:            current.Icon,
		WeatherCode:     current.WeatherCode,
		HighTemperature: highTemp,
		LowTemperature:  lowTemp,
		FetchedAt:       m.FetchedAt(),
	}
}

type DailyRestModel struct {
	Id              string  `json:"-"`
	Date            string  `json:"date"`
	HighTemperature float64 `json:"highTemperature"`
	LowTemperature  float64 `json:"lowTemperature"`
	TemperatureUnit string  `json:"temperatureUnit"`
	Summary         string  `json:"summary"`
	Icon            string  `json:"icon"`
	WeatherCode     int     `json:"weatherCode"`
}

func (r DailyRestModel) GetName() string { return "weather-daily" }
func (r DailyRestModel) GetID() string   { return r.Id }
func (r *DailyRestModel) SetID(id string) error {
	r.Id = id
	return nil
}

func TransformForecast(m Model) []DailyRestModel {
	tempUnit := m.TemperatureUnit()
	result := make([]DailyRestModel, len(m.ForecastData()))
	for i, d := range m.ForecastData() {
		result[i] = DailyRestModel{
			Id:              d.Date,
			Date:            d.Date,
			HighTemperature: d.HighTemperature,
			LowTemperature:  d.LowTemperature,
			TemperatureUnit: tempUnit,
			Summary:         d.Summary,
			Icon:            d.Icon,
			WeatherCode:     d.WeatherCode,
		}
	}
	return result
}
