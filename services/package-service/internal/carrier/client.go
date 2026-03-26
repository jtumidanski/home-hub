package carrier

import (
	"context"
	"net/http"
	"time"
)

// TrackingEvent represents a single normalized tracking event from a carrier.
type TrackingEvent struct {
	Timestamp   time.Time
	Status      string
	Description string
	Location    string
	RawStatus   string
}

// TrackingResult is the normalized result from a carrier tracking API call.
type TrackingResult struct {
	Found             bool
	Status            string
	EstimatedDelivery *time.Time
	ActualDelivery    *time.Time
	Events            []TrackingEvent
}

// CarrierClient defines the interface for carrier tracking API implementations.
type CarrierClient interface {
	// Track queries the carrier for tracking information.
	Track(ctx context.Context, trackingNumber string) (TrackingResult, error)

	// Name returns the carrier identifier (e.g., "usps", "ups", "fedex").
	Name() string
}

// Registry maps carrier names to their client implementations.
type Registry struct {
	clients map[string]CarrierClient
}

// NewRegistry creates a new carrier client registry.
func NewRegistry() *Registry {
	return &Registry{clients: make(map[string]CarrierClient)}
}

// Register adds a carrier client to the registry.
func (r *Registry) Register(client CarrierClient) {
	r.clients[client.Name()] = client
}

// Get returns the carrier client for the given carrier name.
func (r *Registry) Get(name string) (CarrierClient, bool) {
	c, ok := r.clients[name]
	return c, ok
}

// NewHTTPClient creates a shared HTTP client with a sensible timeout for carrier API calls.
func NewHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 15 * time.Second,
	}
}
