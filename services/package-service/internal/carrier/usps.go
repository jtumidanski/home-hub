package carrier

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	uspsTokenURL    = "https://api.usps.com/oauth2/v3/token"
	uspsTrackingURL = "https://api.usps.com/tracking/v3/tracking"
)

// USPSClient implements CarrierClient for the USPS Tracking API v3.
type USPSClient struct {
	tokenMgr *OAuthTokenManager
	budget   *RateBudget
	oauth    OAuthConfig
	client   *http.Client
	l        logrus.FieldLogger
}

// NewUSPSClient creates a new USPS carrier client.
func NewUSPSClient(clientID, clientSecret string, tokenMgr *OAuthTokenManager, budget *RateBudget, client *http.Client, l logrus.FieldLogger) *USPSClient {
	return &USPSClient{
		tokenMgr: tokenMgr,
		budget:   budget,
		client:   client,
		oauth: OAuthConfig{
			TokenURL:     uspsTokenURL,
			ClientID:     clientID,
			ClientSecret: clientSecret,
		},
		l: l,
	}
}

func (c *USPSClient) Name() string { return "usps" }

func (c *USPSClient) Track(ctx context.Context, trackingNumber string) (TrackingResult, error) {
	if !c.budget.CanRequest(c.Name()) {
		return TrackingResult{}, fmt.Errorf("USPS daily rate budget exceeded")
	}

	token, err := c.tokenMgr.GetToken(ctx, c.Name(), c.oauth)
	if err != nil {
		return TrackingResult{}, fmt.Errorf("USPS auth failed: %w", err)
	}

	reqURL := fmt.Sprintf("%s/%s", uspsTrackingURL, trackingNumber)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return TrackingResult{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return TrackingResult{}, fmt.Errorf("USPS tracking request failed: %w", err)
	}
	defer resp.Body.Close()
	c.budget.Record(c.Name())

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return TrackingResult{}, fmt.Errorf("failed to read USPS response: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return TrackingResult{Found: false}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return TrackingResult{}, fmt.Errorf("USPS returned status %d: %s", resp.StatusCode, string(body))
	}

	return c.parseResponse(body)
}

func (c *USPSClient) parseResponse(body []byte) (TrackingResult, error) {
	var resp uspsTrackingResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return TrackingResult{}, fmt.Errorf("failed to parse USPS response: %w", err)
	}

	if resp.TrackingNumber == "" {
		return TrackingResult{Found: false}, nil
	}

	result := TrackingResult{
		Found:  true,
		Status: normalizeUSPSStatus(resp.StatusCategory),
	}

	if resp.ExpectedDeliveryDate != "" {
		if t, err := time.Parse("2006-01-02", resp.ExpectedDeliveryDate); err == nil {
			result.EstimatedDelivery = &t
		}
	}

	if resp.ActualDeliveryDate != "" {
		if t, err := time.Parse(time.RFC3339, resp.ActualDeliveryDate); err == nil {
			result.ActualDelivery = &t
		}
	}

	for _, e := range resp.TrackingEvents {
		ts, _ := time.Parse(time.RFC3339, e.EventTimestamp)
		loc := ""
		if e.EventCity != "" {
			loc = e.EventCity
			if e.EventState != "" {
				loc += ", " + e.EventState
			}
		}
		result.Events = append(result.Events, TrackingEvent{
			Timestamp:   ts,
			Status:      normalizeUSPSStatus(e.EventType),
			Description: e.EventDescription,
			Location:    loc,
			RawStatus:   e.EventType,
		})
	}

	return result, nil
}

type uspsTrackingResponse struct {
	TrackingNumber       string              `json:"trackingNumber"`
	StatusCategory       string              `json:"statusCategory"`
	ExpectedDeliveryDate string              `json:"expectedDeliveryDate"`
	ActualDeliveryDate   string              `json:"actualDeliveryDate"`
	TrackingEvents       []uspsTrackingEvent `json:"trackingEvents"`
}

type uspsTrackingEvent struct {
	EventTimestamp   string `json:"eventTimestamp"`
	EventType        string `json:"eventType"`
	EventDescription string `json:"eventDescription"`
	EventCity        string `json:"eventCity"`
	EventState       string `json:"eventState"`
}

func normalizeUSPSStatus(status string) string {
	switch status {
	case "Pre-Shipment", "Accepted", "Label Created":
		return "pre_transit"
	case "In Transit", "In-Transit", "Arrived", "Departed", "Processing":
		return "in_transit"
	case "Out for Delivery":
		return "out_for_delivery"
	case "Delivered":
		return "delivered"
	case "Alert", "Return to Sender", "Undeliverable":
		return "exception"
	default:
		return "in_transit"
	}
}
