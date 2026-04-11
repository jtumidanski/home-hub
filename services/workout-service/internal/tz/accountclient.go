package tz

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// NewAccountHouseholdLookup returns a HouseholdLookup that fetches the
// household's configured timezone from account-service. The Authorization
// header from the incoming request is forwarded so account-service can
// validate the caller — no shared service token exists today.
//
// Returns empty string (with no error) if the household has no timezone set;
// Resolve treats that as "fall through to UTC".
func NewAccountHouseholdLookup(baseURL, authHeader string) HouseholdLookup {
	client := &http.Client{Timeout: 3 * time.Second}
	return func(ctx context.Context, householdID uuid.UUID) (string, error) {
		if baseURL == "" || householdID == uuid.Nil {
			return "", errors.New("missing base URL or household id")
		}
		url := fmt.Sprintf("%s/api/v1/households/%s", baseURL, householdID.String())
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return "", err
		}
		if authHeader != "" {
			req.Header.Set("Authorization", authHeader)
		}
		req.Header.Set("Accept", "application/vnd.api+json")

		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
			return "", fmt.Errorf("account-service returned %d: %s", resp.StatusCode, string(body))
		}

		var doc struct {
			Data struct {
				Attributes struct {
					Timezone string `json:"timezone"`
				} `json:"attributes"`
			} `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
			return "", err
		}
		return doc.Data.Attributes.Timezone, nil
	}
}
