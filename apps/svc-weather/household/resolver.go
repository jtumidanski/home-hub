package household

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	gocache "github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
)

// Location represents the geographic location and timezone of a household
type Location struct {
	Latitude  float64
	Longitude float64
	Timezone  string
}

// Resolver defines the interface for resolving household locations
type Resolver interface {
	Resolve(ctx context.Context, householdID uuid.UUID) (Location, error)
}

// HTTPResolver resolves household locations from svc-users via HTTP
type HTTPResolver struct {
	baseURL string
	timeout time.Duration
	cache   *gocache.Cache
	client  *http.Client
	logger  logrus.FieldLogger
}

// NewHTTPResolver creates a new household resolver with in-memory cache
func NewHTTPResolver(baseURL string, timeout, cacheTTL time.Duration, logger logrus.FieldLogger) *HTTPResolver {
	return &HTTPResolver{
		baseURL: baseURL,
		timeout: timeout,
		cache:   gocache.New(cacheTTL, cacheTTL*2),
		client: &http.Client{
			Timeout: timeout,
		},
		logger: logger,
	}
}

// Resolve fetches household location from svc-users (with 24h cache)
func (r *HTTPResolver) Resolve(ctx context.Context, householdID uuid.UUID) (Location, error) {
	cacheKey := fmt.Sprintf("household:%s", householdID.String())

	// Check cache first
	if cached, found := r.cache.Get(cacheKey); found {
		r.logger.WithField("household_id", householdID).Debug("Household location cache hit")
		return cached.(Location), nil
	}

	r.logger.WithField("household_id", householdID).Debug("Household location cache miss, fetching from svc-users")

	// Fetch from svc-users
	location, err := r.fetchFromSvcUsers(ctx, householdID)
	if err != nil {
		return Location{}, err
	}

	// Cache the result
	r.cache.SetDefault(cacheKey, location)

	return location, nil
}

// fetchFromSvcUsers performs the HTTP request to svc-users
func (r *HTTPResolver) fetchFromSvcUsers(ctx context.Context, householdID uuid.UUID) (Location, error) {
	url := fmt.Sprintf("%s/api/v1/households/%s", r.baseURL, householdID.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Location{}, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return Location{}, fmt.Errorf("HTTP request to svc-users failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return Location{}, fmt.Errorf("household %s not found", householdID)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return Location{}, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var response HouseholdResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return Location{}, fmt.Errorf("failed to decode response: %w", err)
	}

	// Validate that location data exists
	if response.Data.Attributes.Latitude == nil || response.Data.Attributes.Longitude == nil {
		return Location{}, fmt.Errorf("household %s has no location data", householdID)
	}

	if response.Data.Attributes.Timezone == nil {
		return Location{}, fmt.Errorf("household %s has no timezone data", householdID)
	}

	return Location{
		Latitude:  *response.Data.Attributes.Latitude,
		Longitude: *response.Data.Attributes.Longitude,
		Timezone:  *response.Data.Attributes.Timezone,
	}, nil
}

// HouseholdResponse represents the JSON:API response from svc-users
type HouseholdResponse struct {
	Data HouseholdData `json:"data"`
}

// HouseholdData represents the data section of the JSON:API response
type HouseholdData struct {
	Type       string               `json:"type"`
	ID         string               `json:"id"`
	Attributes HouseholdAttributes `json:"attributes"`
}

// HouseholdAttributes contains the household attributes
type HouseholdAttributes struct {
	Name      string   `json:"name"`
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
	Timezone  *string  `json:"timezone,omitempty"`
}
