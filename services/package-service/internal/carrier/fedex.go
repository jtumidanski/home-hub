package carrier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	fedexProdBaseURL    = "https://apis.fedex.com"
	fedexSandboxBaseURL = "https://apis-sandbox.fedex.com"
)

// FedExClient implements CarrierClient for the FedEx Track API v1.
type FedExClient struct {
	tokenMgr    *OAuthTokenManager
	budget      *RateBudget
	oauth       OAuthConfig
	trackingURL string
	client      *http.Client
	l           logrus.FieldLogger
}

// NewFedExClient creates a new FedEx carrier client.
// If sandbox is true, uses the FedEx sandbox environment.
func NewFedExClient(apiKey, secretKey string, sandbox bool, tokenMgr *OAuthTokenManager, budget *RateBudget, client *http.Client, l logrus.FieldLogger) *FedExClient {
	baseURL := fedexProdBaseURL
	if sandbox {
		baseURL = fedexSandboxBaseURL
	}
	return &FedExClient{
		tokenMgr:    tokenMgr,
		budget:      budget,
		trackingURL: baseURL + "/track/v1/trackingnumbers",
		client:      client,
		oauth: OAuthConfig{
			TokenURL:     baseURL + "/oauth/token",
			ClientID:     apiKey,
			ClientSecret: secretKey,
		},
		l: l,
	}
}

func (c *FedExClient) Name() string { return "fedex" }

func (c *FedExClient) Track(ctx context.Context, trackingNumber string) (TrackingResult, error) {
	if !c.budget.CanRequest(c.Name()) {
		return TrackingResult{}, fmt.Errorf("FedEx daily rate budget exceeded")
	}

	token, err := c.tokenMgr.GetToken(ctx, c.Name(), c.oauth)
	if err != nil {
		return TrackingResult{}, fmt.Errorf("FedEx auth failed: %w", err)
	}

	reqBody := fedexTrackRequest{
		IncludeDetailedScans: true,
		TrackingInfo: []fedexTrackingInfo{
			{TrackingNumberInfo: fedexTrackingNumberInfo{TrackingNumber: trackingNumber}},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return TrackingResult{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.trackingURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return TrackingResult{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-locale", "en_US")

	resp, err := c.client.Do(req)
	if err != nil {
		return TrackingResult{}, fmt.Errorf("FedEx tracking request failed: %w", err)
	}
	defer resp.Body.Close()
	c.budget.Record(c.Name())

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return TrackingResult{}, fmt.Errorf("failed to read FedEx response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return TrackingResult{}, fmt.Errorf("FedEx returned status %d: %s", resp.StatusCode, string(body))
	}

	return c.parseResponse(body)
}

func (c *FedExClient) parseResponse(body []byte) (TrackingResult, error) {
	var resp fedexTrackResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		c.l.WithField("body", string(body)).Debug("raw FedEx response")
		return TrackingResult{}, fmt.Errorf("failed to parse FedEx response: %w", err)
	}

	if len(resp.Output.CompleteTrackResults) == 0 ||
		len(resp.Output.CompleteTrackResults[0].TrackResults) == 0 {
		return TrackingResult{Found: false}, nil
	}

	track := resp.Output.CompleteTrackResults[0].TrackResults[0]

	if track.Error != nil && track.Error.Code != "" {
		return TrackingResult{Found: false}, nil
	}

	result := TrackingResult{
		Found:  true,
		Status: normalizeFedExStatus(track.LatestStatusDetail.StatusByLocale),
	}

	// Parse estimatedDeliveryTimeWindow — FedEx returns either an object or array
	if len(track.EstimatedDeliveryTimeWindow) > 0 {
		var windows []fedexDeliveryWindow
		// Try array first
		if err := json.Unmarshal(track.EstimatedDeliveryTimeWindow, &windows); err != nil {
			// Try single object
			var single fedexDeliveryWindow
			if err2 := json.Unmarshal(track.EstimatedDeliveryTimeWindow, &single); err2 == nil {
				windows = []fedexDeliveryWindow{single}
			}
		}
		for _, w := range windows {
			if w.Window.Ends != "" {
				if t, err := time.Parse(time.RFC3339, w.Window.Ends); err == nil {
					result.EstimatedDelivery = &t
				}
			}
		}
	}

	// Parse dateAndTimes for delivery date (more reliable for delivered packages)
	for _, dt := range track.DateAndTimes {
		if dt.Type == "ACTUAL_DELIVERY" && dt.DateTime != "" {
			if t, err := time.Parse(time.RFC3339, dt.DateTime); err == nil {
				result.ActualDelivery = &t
			}
		}
		if dt.Type == "ESTIMATED_DELIVERY" && dt.DateTime != "" && result.EstimatedDelivery == nil {
			if t, err := time.Parse(time.RFC3339, dt.DateTime); err == nil {
				result.EstimatedDelivery = &t
			}
		}
	}

	for _, scan := range track.ScanEvents {
		ts, _ := time.Parse(time.RFC3339, scan.Date)
		loc := ""
		if scan.ScanLocation.City != "" {
			loc = scan.ScanLocation.City
			if scan.ScanLocation.StateOrProvinceCode != "" {
				loc += ", " + scan.ScanLocation.StateOrProvinceCode
			}
		}
		result.Events = append(result.Events, TrackingEvent{
			Timestamp:   ts,
			Status:      normalizeFedExDerivedStatus(scan.DerivedStatusCode),
			Description: scan.EventDescription,
			Location:    loc,
			RawStatus:   scan.DerivedStatusCode,
		})
	}

	return result, nil
}

// Request types
type fedexTrackRequest struct {
	IncludeDetailedScans bool               `json:"includeDetailedScans"`
	TrackingInfo         []fedexTrackingInfo `json:"trackingInfo"`
}

type fedexTrackingInfo struct {
	TrackingNumberInfo fedexTrackingNumberInfo `json:"trackingNumberInfo"`
}

type fedexTrackingNumberInfo struct {
	TrackingNumber string `json:"trackingNumber"`
}

// Response types
type fedexDeliveryWindow struct {
	Window struct {
		Ends string `json:"ends"`
	} `json:"window"`
}

type fedexTrackResponse struct {
	Output struct {
		CompleteTrackResults []struct {
			TrackResults []struct {
				LatestStatusDetail struct {
					StatusByLocale string `json:"statusByLocale"`
					Code           string `json:"code"`
				} `json:"latestStatusDetail"`
				EstimatedDeliveryTimeWindow json.RawMessage `json:"estimatedDeliveryTimeWindow"`
				DateAndTimes                []struct {
					Type     string `json:"type"`
					DateTime string `json:"dateTime"`
				} `json:"dateAndTimes"`
				ScanEvents []struct {
					Date              string `json:"date"`
					EventType         string `json:"eventType"`
					EventDescription  string `json:"eventDescription"`
					DerivedStatusCode string `json:"derivedStatusCode"`
					ScanLocation      struct {
						City                string `json:"city"`
						StateOrProvinceCode string `json:"stateOrProvinceCode"`
						CountryCode         string `json:"countryCode"`
					} `json:"scanLocation"`
				} `json:"scanEvents"`
				Error *struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				} `json:"error"`
			} `json:"trackResults"`
		} `json:"completeTrackResults"`
	} `json:"output"`
}

// normalizeFedExStatus normalizes human-readable FedEx status strings (used for latestStatusDetail).
func normalizeFedExStatus(status string) string {
	switch status {
	case "Label created", "Initiated", "Picked Up":
		return "pre_transit"
	case "In transit", "In Transit", "At local FedEx facility", "At destination sort facility",
		"Departed FedEx location", "Arrived at FedEx location", "International shipment release":
		return "in_transit"
	case "On FedEx vehicle for delivery", "Out for Delivery", "Out for delivery":
		return "out_for_delivery"
	case "Delivered":
		return "delivered"
	case "Delivery exception", "Customer not available or business closed":
		return "exception"
	default:
		return "in_transit"
	}
}

// normalizeFedExDerivedStatus normalizes FedEx derivedStatusCode values (used for scan events).
func normalizeFedExDerivedStatus(code string) string {
	switch code {
	case "PU", "OC":
		return "pre_transit"
	case "IT", "AR", "DP", "AF", "CC", "CD", "ED", "HP", "IX", "OX", "SP", "TR":
		return "in_transit"
	case "OD":
		return "out_for_delivery"
	case "DL":
		return "delivered"
	case "DE", "CA", "RS", "SE":
		return "exception"
	default:
		return "in_transit"
	}
}
