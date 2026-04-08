package googlecal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	authEndpoint     = "https://accounts.google.com/o/oauth2/v2/auth"
	tokenEndpoint    = "https://oauth2.googleapis.com/token"
	revokeEndpoint   = "https://oauth2.googleapis.com/revoke"
	calendarListURL  = "https://www.googleapis.com/calendar/v3/users/me/calendarList"
	eventsURLPattern = "https://www.googleapis.com/calendar/v3/calendars/%s/events"
	calendarScope    = "https://www.googleapis.com/auth/calendar email"

	maxRetries     = 3
	initialBackoff = 1 * time.Second
)

type Client struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client
	l            logrus.FieldLogger
}

func NewClient(clientID, clientSecret string, l logrus.FieldLogger) *Client {
	return &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
		l:            l,
	}
}

func AuthURL(clientID, redirectURI, state string, forceConsent bool) string {
	params := url.Values{
		"response_type": {"code"},
		"client_id":     {clientID},
		"redirect_uri":  {redirectURI},
		"scope":         {calendarScope},
		"state":         {state},
		"access_type":   {"offline"},
	}
	if forceConsent {
		params.Set("prompt", "consent")
	}
	return authEndpoint + "?" + params.Encode()
}

func (c *Client) ExchangeCode(ctx context.Context, code, redirectURI string) (*TokenResponse, error) {
	data := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"client_id":     {c.clientID},
		"client_secret": {c.clientSecret},
	}

	resp, err := c.httpClient.PostForm(tokenEndpoint, data)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token endpoint returned %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}
	return &tokenResp, nil
}

func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {c.clientID},
		"client_secret": {c.clientSecret},
	}

	resp, err := c.httpClient.PostForm(tokenEndpoint, data)
	if err != nil {
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		var parsed struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}
		_ = json.Unmarshal(body, &parsed)
		return nil, &TokenRefreshError{
			StatusCode: resp.StatusCode,
			OAuthError: parsed.Error,
			Body:       string(body),
		}
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse refresh response: %w", err)
	}
	return &tokenResp, nil
}

func (c *Client) RevokeToken(ctx context.Context, token string) error {
	data := url.Values{"token": {token}}
	resp, err := c.httpClient.PostForm(revokeEndpoint, data)
	if err != nil {
		return fmt.Errorf("token revocation failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.l.WithField("status", resp.StatusCode).Warn("token revocation returned non-200")
	}
	return nil
}

func (c *Client) ListCalendars(ctx context.Context, accessToken string) (*CalendarListResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, calendarListURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	var result CalendarListResponse
	if err := c.doWithRetry(req, &result); err != nil {
		return nil, fmt.Errorf("list calendars failed: %w", err)
	}
	return &result, nil
}

func (c *Client) ListEvents(ctx context.Context, accessToken, calendarID string, timeMin, timeMax time.Time, syncToken string) (*EventsResponse, error) {
	u := fmt.Sprintf(eventsURLPattern, url.PathEscape(calendarID))

	params := url.Values{
		"singleEvents": {"true"},
		"orderBy":      {"startTime"},
		"maxResults":   {"2500"},
	}

	if syncToken != "" {
		params.Set("syncToken", syncToken)
		params.Set("showDeleted", "true")
	} else {
		params.Set("timeMin", timeMin.Format(time.RFC3339))
		params.Set("timeMax", timeMax.Format(time.RFC3339))
	}

	var allEvents []Event
	var finalSyncToken string

	for {
		reqURL := u + "?" + params.Encode()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)

		var page EventsResponse
		if err := c.doWithRetry(req, &page); err != nil {
			return nil, fmt.Errorf("list events failed: %w", err)
		}

		allEvents = append(allEvents, page.Items...)

		if page.NextPageToken == "" {
			finalSyncToken = page.NextSyncToken
			break
		}
		params.Set("pageToken", page.NextPageToken)
	}

	return &EventsResponse{
		Items:         allEvents,
		NextSyncToken: finalSyncToken,
	}, nil
}

func (c *Client) InsertEvent(ctx context.Context, accessToken, calendarID string, event InsertEventRequest) (*Event, error) {
	u := fmt.Sprintf(eventsURLPattern, url.PathEscape(calendarID))
	body, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal insert event request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	var result Event
	if err := c.doWithRetry(req, &result); err != nil {
		return nil, fmt.Errorf("insert event failed: %w", err)
	}
	return &result, nil
}

func (c *Client) UpdateEvent(ctx context.Context, accessToken, calendarID, eventID string, event UpdateEventRequest) (*Event, error) {
	u := fmt.Sprintf(eventsURLPattern+"/%s", url.PathEscape(calendarID), url.PathEscape(eventID))
	body, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal update event request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, u, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	var result Event
	if err := c.doWithRetry(req, &result); err != nil {
		return nil, fmt.Errorf("update event failed: %w", err)
	}
	return &result, nil
}

func (c *Client) DeleteEvent(ctx context.Context, accessToken, calendarID, eventID string) error {
	u := fmt.Sprintf(eventsURLPattern+"/%s", url.PathEscape(calendarID), url.PathEscape(eventID))

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(math.Pow(2, float64(attempt-1))) * initialBackoff
			time.Sleep(backoff)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK {
			return nil
		}
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("Google API returned %d", resp.StatusCode)
			continue
		}
		return fmt.Errorf("Google API delete returned %d", resp.StatusCode)
	}
	return fmt.Errorf("Google API delete failed after %d retries: %w", maxRetries, lastErr)
}

func (c *Client) doWithRetry(req *http.Request, result interface{}) error {
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(math.Pow(2, float64(attempt-1))) * initialBackoff
			c.l.WithField("backoff", backoff).WithField("attempt", attempt).Debug("retrying Google API call")
			time.Sleep(backoff)

			// Reset request body for retry if GetBody is available
			if req.GetBody != nil {
				newBody, err := req.GetBody()
				if err == nil {
					req.Body = newBody
				}
			}
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode == http.StatusOK {
			return json.Unmarshal(body, result)
		}

		if resp.StatusCode == http.StatusGone {
			return &SyncTokenInvalidError{StatusCode: resp.StatusCode, Body: string(body)}
		}

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("Google API returned %d: %s", resp.StatusCode, string(body))
			continue
		}

		return fmt.Errorf("Google API returned %d: %s", resp.StatusCode, string(body))
	}
	return fmt.Errorf("Google API call failed after %d retries: %w", maxRetries, lastErr)
}

func (c *Client) FetchUserEmail(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch user info failed: %w", err)
	}
	defer resp.Body.Close()

	var info struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", fmt.Errorf("failed to parse user info: %w", err)
	}
	return info.Email, nil
}

// SyncTokenInvalidError indicates the sync token has been invalidated and a full sync is needed.
type SyncTokenInvalidError struct {
	StatusCode int
	Body       string
}

func (e *SyncTokenInvalidError) Error() string {
	return fmt.Sprintf("sync token invalid (HTTP %d): %s", e.StatusCode, e.Body)
}

// IsSyncTokenInvalid checks if an error is a sync token invalidation error.
func IsSyncTokenInvalid(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "sync token invalid")
}

// TokenRefreshError indicates a non-200 response from the OAuth refresh endpoint.
// OAuthError carries the parsed `error` field from Google's JSON body (e.g. "invalid_grant"),
// or "" if the body could not be parsed as JSON.
type TokenRefreshError struct {
	StatusCode int
	OAuthError string
	Body       string
}

func (e *TokenRefreshError) Error() string {
	return fmt.Sprintf("token refresh failed (HTTP %d, oauth_error=%q): %s", e.StatusCode, e.OAuthError, e.Body)
}

// IsInvalidGrant returns true when err is (or wraps) a *TokenRefreshError whose
// OAuth error code is "invalid_grant" — i.e. the refresh token has been revoked
// or expired and the user must reauthorize.
func IsInvalidGrant(err error) bool {
	var tre *TokenRefreshError
	if errors.As(err, &tre) {
		return tre.OAuthError == "invalid_grant"
	}
	return false
}
