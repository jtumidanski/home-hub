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
	upsTokenURL    = "https://onlinetools.ups.com/security/v1/oauth/token"
	upsTrackingURL = "https://onlinetools.ups.com/api/track/v1/details"
)

// UPSClient implements CarrierClient for the UPS Tracking API v1.
type UPSClient struct {
	tokenMgr *OAuthTokenManager
	budget   *RateBudget
	oauth    OAuthConfig
	client   *http.Client
	l        logrus.FieldLogger
}

// NewUPSClient creates a new UPS carrier client.
func NewUPSClient(clientID, clientSecret string, tokenMgr *OAuthTokenManager, budget *RateBudget, client *http.Client, l logrus.FieldLogger) *UPSClient {
	return &UPSClient{
		tokenMgr: tokenMgr,
		budget:   budget,
		client:   client,
		oauth: OAuthConfig{
			TokenURL:     upsTokenURL,
			ClientID:     clientID,
			ClientSecret: clientSecret,
		},
		l: l,
	}
}

func (c *UPSClient) Name() string { return "ups" }

func (c *UPSClient) Track(ctx context.Context, trackingNumber string) (TrackingResult, error) {
	if !c.budget.CanRequest(c.Name()) {
		return TrackingResult{}, fmt.Errorf("UPS daily rate budget exceeded")
	}

	token, err := c.tokenMgr.GetToken(ctx, c.Name(), c.oauth)
	if err != nil {
		return TrackingResult{}, fmt.Errorf("UPS auth failed: %w", err)
	}

	reqURL := fmt.Sprintf("%s/%s", upsTrackingURL, trackingNumber)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return TrackingResult{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("transId", fmt.Sprintf("home-hub-%d", time.Now().UnixNano()))
	req.Header.Set("transactionSrc", "home-hub")

	resp, err := c.client.Do(req)
	if err != nil {
		return TrackingResult{}, fmt.Errorf("UPS tracking request failed: %w", err)
	}
	defer resp.Body.Close()
	c.budget.Record(c.Name())

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return TrackingResult{}, fmt.Errorf("failed to read UPS response: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusBadRequest {
		return TrackingResult{Found: false}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return TrackingResult{}, fmt.Errorf("UPS returned status %d: %s", resp.StatusCode, string(body))
	}

	return c.parseResponse(body)
}

func (c *UPSClient) parseResponse(body []byte) (TrackingResult, error) {
	var resp upsTrackingResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return TrackingResult{}, fmt.Errorf("failed to parse UPS response: %w", err)
	}

	if len(resp.TrackResponse.Shipment) == 0 || len(resp.TrackResponse.Shipment[0].Package) == 0 {
		return TrackingResult{Found: false}, nil
	}

	pkg := resp.TrackResponse.Shipment[0].Package[0]
	result := TrackingResult{
		Found:  true,
		Status: normalizeUPSStatus(pkg.CurrentStatus.Type),
	}

	if pkg.DeliveryDate != nil && len(pkg.DeliveryDate) > 0 {
		dateStr := pkg.DeliveryDate[0].Date
		if t, err := time.Parse("20060102", dateStr); err == nil {
			result.EstimatedDelivery = &t
		}
	}

	for _, act := range pkg.Activity {
		ts := parseUPSDateTime(act.Date, act.Time)
		loc := ""
		if act.Location.Address.City != "" {
			loc = act.Location.Address.City
			if act.Location.Address.StateProvince != "" {
				loc += ", " + act.Location.Address.StateProvince
			}
		}
		result.Events = append(result.Events, TrackingEvent{
			Timestamp:   ts,
			Status:      normalizeUPSStatus(act.Status.Type),
			Description: act.Status.Description,
			Location:    loc,
			RawStatus:   act.Status.Code,
		})
	}

	return result, nil
}

type upsTrackingResponse struct {
	TrackResponse struct {
		Shipment []struct {
			Package []struct {
				CurrentStatus struct {
					Type        string `json:"type"`
					Description string `json:"description"`
					Code        string `json:"code"`
				} `json:"currentStatus"`
				DeliveryDate []struct {
					Date string `json:"date"`
				} `json:"deliveryDate"`
				Activity []struct {
					Date     string `json:"date"`
					Time     string `json:"time"`
					Location struct {
						Address struct {
							City          string `json:"city"`
							StateProvince string `json:"stateProvince"`
							Country       string `json:"country"`
						} `json:"address"`
					} `json:"location"`
					Status struct {
						Type        string `json:"type"`
						Description string `json:"description"`
						Code        string `json:"code"`
					} `json:"status"`
				} `json:"activity"`
			} `json:"package"`
		} `json:"shipment"`
	} `json:"trackResponse"`
}

func normalizeUPSStatus(statusType string) string {
	switch statusType {
	case "M", "MV", "P":
		return "pre_transit"
	case "I", "X":
		return "in_transit"
	case "O":
		return "out_for_delivery"
	case "D":
		return "delivered"
	case "RS", "NA", "MN":
		return "exception"
	default:
		return "in_transit"
	}
}

func parseUPSDateTime(date, timeStr string) time.Time {
	combined := date + timeStr
	t, err := time.Parse("20060102150405", combined)
	if err != nil {
		t, _ = time.Parse("20060102", date)
	}
	return t
}
