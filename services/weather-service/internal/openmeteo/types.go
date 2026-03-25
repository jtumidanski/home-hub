package openmeteo

type ForecastResponse struct {
	Current      CurrentData  `json:"current"`
	CurrentUnits CurrentUnits `json:"current_units"`
	Daily        DailyData    `json:"daily"`
	DailyUnits   DailyUnits   `json:"daily_units"`
	Hourly       HourlyData   `json:"hourly"`
	HourlyUnits  HourlyUnits  `json:"hourly_units"`
}

type CurrentData struct {
	Temperature float64 `json:"temperature_2m"`
	WeatherCode int     `json:"weather_code"`
}

type CurrentUnits struct {
	Temperature string `json:"temperature_2m"`
}

type DailyData struct {
	Time           []string  `json:"time"`
	TemperatureMax []float64 `json:"temperature_2m_max"`
	TemperatureMin []float64 `json:"temperature_2m_min"`
	WeatherCode    []int     `json:"weather_code"`
}

type DailyUnits struct {
	TemperatureMax string `json:"temperature_2m_max"`
	TemperatureMin string `json:"temperature_2m_min"`
}

type HourlyData struct {
	Time                     []string  `json:"time"`
	Temperature              []float64 `json:"temperature_2m"`
	WeatherCode              []int     `json:"weather_code"`
	PrecipitationProbability []int     `json:"precipitation_probability"`
}

type HourlyUnits struct {
	Temperature              string `json:"temperature_2m"`
	WeatherCode              string `json:"weather_code"`
	PrecipitationProbability string `json:"precipitation_probability"`
}

type GeocodingResponse struct {
	Results []GeocodingResult `json:"results"`
}

type GeocodingResult struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	Country   string  `json:"country"`
	Admin1    string  `json:"admin1"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
