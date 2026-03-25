package openmeteo

import (
	"net/http"
	"time"
)

func newTestClient(forecastURL, geocodingURL string) *Client {
	return &Client{
		httpClient:   &http.Client{Timeout: 5 * time.Second},
		forecastURL:  forecastURL,
		geocodingURL: geocodingURL,
	}
}
