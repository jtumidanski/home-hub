package openmeteo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	forecastBaseURL  = "https://api.open-meteo.com/v1/forecast"
	geocodingBaseURL = "https://geocoding-api.open-meteo.com/v1/search"
)

type Client struct {
	httpClient   *http.Client
	forecastURL  string
	geocodingURL string
	mu           sync.Mutex
	lastCall     time.Time
}

func NewClient() *Client {
	return &Client{
		httpClient:   &http.Client{Timeout: 10 * time.Second},
		forecastURL:  forecastBaseURL,
		geocodingURL: geocodingBaseURL,
	}
}

// NewClientWithEndpoints constructs a Client with caller-supplied upstream
// URLs. Intended for cross-package tests that point the client at an
// httptest server.
func NewClientWithEndpoints(forecastURL, geocodingURL string) *Client {
	return &Client{
		httpClient:   &http.Client{Timeout: 10 * time.Second},
		forecastURL:  forecastURL,
		geocodingURL: geocodingURL,
	}
}

func (c *Client) throttle() {
	c.mu.Lock()
	defer c.mu.Unlock()
	since := time.Since(c.lastCall)
	if since < time.Second {
		time.Sleep(time.Second - since)
	}
	c.lastCall = time.Now()
}

func temperatureUnit(units string) string {
	if units == "imperial" {
		return "fahrenheit"
	}
	return "celsius"
}

func (c *Client) FetchForecast(lat, lon float64, units, timezone string) (*ForecastResponse, error) {
	c.throttle()

	params := url.Values{}
	params.Set("latitude", fmt.Sprintf("%f", lat))
	params.Set("longitude", fmt.Sprintf("%f", lon))
	params.Set("current", "temperature_2m,weather_code")
	params.Set("daily", "temperature_2m_max,temperature_2m_min,weather_code")
	params.Set("hourly", "temperature_2m,weather_code,precipitation_probability")
	params.Set("temperature_unit", temperatureUnit(units))
	params.Set("timezone", timezone)
	params.Set("forecast_days", "7")

	resp, err := c.httpClient.Get(c.forecastURL + "?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("open-meteo forecast request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("open-meteo forecast returned status %d: %s", resp.StatusCode, string(body))
	}

	var result ForecastResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("open-meteo forecast decode failed: %w", err)
	}
	return &result, nil
}

func (c *Client) SearchPlaces(query string) ([]GeocodingResult, error) {
	c.throttle()

	params := url.Values{}
	params.Set("name", query)
	params.Set("count", "10")
	params.Set("language", "en")

	resp, err := c.httpClient.Get(c.geocodingURL + "?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("open-meteo geocoding request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("open-meteo geocoding returned status %d: %s", resp.StatusCode, string(body))
	}

	var result GeocodingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("open-meteo geocoding decode failed: %w", err)
	}
	return result.Results, nil
}
