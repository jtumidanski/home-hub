package openmeteo

import (
	"time"
)

// CurrentResponse represents the Open-Meteo API response for current weather
type CurrentResponse struct {
	Latitude  float64        `json:"latitude"`
	Longitude float64        `json:"longitude"`
	Timezone  string         `json:"timezone"`
	Current   CurrentUnits   `json:"current"`
	CurrentUnits CurrentUnitsInfo `json:"current_units"`
}

// CurrentUnits contains current weather data
type CurrentUnits struct {
	Time         string  `json:"time"`
	Temperature2m float64 `json:"temperature_2m"`
}

// CurrentUnitsInfo contains unit information
type CurrentUnitsInfo struct {
	Time         string `json:"time"`
	Temperature2m string `json:"temperature_2m"`
}

// ForecastResponse represents the Open-Meteo API response for forecast weather
type ForecastResponse struct {
	Latitude  float64      `json:"latitude"`
	Longitude float64      `json:"longitude"`
	Timezone  string       `json:"timezone"`
	Daily     DailyUnits   `json:"daily"`
	DailyUnits DailyUnitsInfo `json:"daily_units"`
}

// DailyUnits contains daily forecast data
type DailyUnits struct {
	Time           []string  `json:"time"`
	Temperature2mMax []float64 `json:"temperature_2m_max"`
	Temperature2mMin []float64 `json:"temperature_2m_min"`
}

// DailyUnitsInfo contains unit information for daily data
type DailyUnitsInfo struct {
	Time           string `json:"time"`
	Temperature2mMax string `json:"temperature_2m_max"`
	Temperature2mMin string `json:"temperature_2m_min"`
}

// ParseCurrentTime parses the time string from Open-Meteo current response
func ParseCurrentTime(timeStr string) (time.Time, error) {
	// Open-Meteo returns time in ISO 8601 format: "2025-11-11T18:55"
	return time.Parse("2006-01-02T15:04", timeStr)
}

// ParseDailyTime parses the date string from Open-Meteo daily response
func ParseDailyTime(dateStr string) (time.Time, error) {
	// Open-Meteo returns dates in format: "2025-11-11"
	return time.Parse("2006-01-02", dateStr)
}
