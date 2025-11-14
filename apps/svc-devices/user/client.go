package user

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Client handles communication with svc-users
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new users service client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// UserResponse represents the response from /api/me endpoint
type UserResponse struct {
	Data struct {
		Type       string `json:"type"`
		ID         string `json:"id"`
		Attributes struct {
			Email       string   `json:"email"`
			DisplayName string   `json:"displayName"`
			Provider    string   `json:"provider"`
			HouseholdID *string  `json:"householdId,omitempty"`
			Roles       []string `json:"roles,omitempty"`
		} `json:"attributes"`
	} `json:"data"`
}

// UserInfo contains user identity information from the users service
type UserInfo struct {
	UserID      uuid.UUID
	HouseholdID uuid.UUID
}

// GetCurrentUser calls /api/me on svc-users with the provided auth headers
// Returns the user information including user ID and household ID
func (c *Client) GetCurrentUser(ctx context.Context, authHeaders map[string]string) (*UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/me", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Forward all auth headers to users service
	for key, value := range authHeaders {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call users service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("users service returned %d: %s", resp.StatusCode, string(body))
	}

	var userResp UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&userResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	userID, err := uuid.Parse(userResp.Data.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	var householdID uuid.UUID
	if userResp.Data.Attributes.HouseholdID != nil {
		householdID, err = uuid.Parse(*userResp.Data.Attributes.HouseholdID)
		if err != nil {
			return nil, fmt.Errorf("invalid household ID format: %w", err)
		}
	}

	return &UserInfo{
		UserID:      userID,
		HouseholdID: householdID,
	}, nil
}
