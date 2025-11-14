package openmeteo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/jtumidanski/home-hub/packages/shared-go/retry"
	"github.com/sirupsen/logrus"
)

// Client defines the interface for fetching weather data from Open-Meteo
type Client interface {
	FetchCurrent(ctx context.Context, lat, lon float64) (CurrentResponse, error)
	FetchForecast(ctx context.Context, lat, lon float64, days int) (ForecastResponse, error)
}

// HTTPClient implements Client using HTTP requests to Open-Meteo API
type HTTPClient struct {
	baseURL string
	timeout time.Duration
	client  *http.Client
	logger  logrus.FieldLogger
}

// NewHTTPClient creates a new Open-Meteo HTTP client
func NewHTTPClient(baseURL string, timeout time.Duration, logger logrus.FieldLogger) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		timeout: timeout,
		client: &http.Client{
			Timeout: timeout,
		},
		logger: logger,
	}
}

// FetchCurrent fetches current weather from Open-Meteo
func (c *HTTPClient) FetchCurrent(ctx context.Context, lat, lon float64) (CurrentResponse, error) {
	queryURL := c.buildCurrentURL(lat, lon)

	c.logger.WithFields(logrus.Fields{
		"lat": lat,
		"lon": lon,
		"url": queryURL,
	}).Debug("Fetching current weather from Open-Meteo")

	var response CurrentResponse
	err := retry.Try(func(attempt int) (bool, error) {
		err := c.doRequest(ctx, queryURL, &response)
		if err != nil {
			c.logger.WithError(err).WithField("attempt", attempt).Warn("Retry fetch current weather")
			return attempt < 3, err
		}
		return false, nil
	}, 3)

	if err != nil {
		return CurrentResponse{}, fmt.Errorf("failed to fetch current weather: %w", err)
	}

	return response, nil
}

// FetchForecast fetches forecast weather from Open-Meteo
func (c *HTTPClient) FetchForecast(ctx context.Context, lat, lon float64, days int) (ForecastResponse, error) {
	queryURL := c.buildForecastURL(lat, lon, days)

	c.logger.WithFields(logrus.Fields{
		"lat":  lat,
		"lon":  lon,
		"days": days,
		"url":  queryURL,
	}).Debug("Fetching forecast weather from Open-Meteo")

	var response ForecastResponse
	err := retry.Try(func(attempt int) (bool, error) {
		err := c.doRequest(ctx, queryURL, &response)
		if err != nil {
			c.logger.WithError(err).WithField("attempt", attempt).Warn("Retry fetch forecast weather")
			return attempt < 3, err
		}
		return false, nil
	}, 3)

	if err != nil {
		return ForecastResponse{}, fmt.Errorf("failed to fetch forecast weather: %w", err)
	}

	return response, nil
}

// buildCurrentURL constructs the URL for current weather API call
func (c *HTTPClient) buildCurrentURL(lat, lon float64) string {
	params := url.Values{}
	params.Add("latitude", fmt.Sprintf("%.6f", lat))
	params.Add("longitude", fmt.Sprintf("%.6f", lon))
	params.Add("current", "temperature_2m")
	params.Add("timezone", "auto")

	return fmt.Sprintf("%s?%s", c.baseURL, params.Encode())
}

// buildForecastURL constructs the URL for forecast weather API call
func (c *HTTPClient) buildForecastURL(lat, lon float64, days int) string {
	params := url.Values{}
	params.Add("latitude", fmt.Sprintf("%.6f", lat))
	params.Add("longitude", fmt.Sprintf("%.6f", lon))
	params.Add("daily", "temperature_2m_max,temperature_2m_min")
	params.Add("timezone", "auto")
	params.Add("forecast_days", fmt.Sprintf("%d", days))

	return fmt.Sprintf("%s?%s", c.baseURL, params.Encode())
}

// doRequest performs an HTTP GET request and decodes the JSON response
func (c *HTTPClient) doRequest(ctx context.Context, url string, result interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}
