package carrier

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// OAuthConfig holds the credentials for a carrier's OAuth 2.0 client credentials flow.
type OAuthConfig struct {
	TokenURL     string
	ClientID     string
	ClientSecret string
}

// OAuthToken represents a cached OAuth 2.0 access token.
type OAuthToken struct {
	AccessToken string
	ExpiresAt   time.Time
}

// IsExpired returns true if the token is expired or about to expire (within 60s buffer).
func (t *OAuthToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt.Add(-60 * time.Second))
}

// OAuthTokenManager manages OAuth tokens for carrier APIs with thread-safe caching.
type OAuthTokenManager struct {
	mu     sync.RWMutex
	tokens map[string]*OAuthToken
	l      logrus.FieldLogger
}

// NewOAuthTokenManager creates a new token manager.
func NewOAuthTokenManager(l logrus.FieldLogger) *OAuthTokenManager {
	return &OAuthTokenManager{
		tokens: make(map[string]*OAuthToken),
		l:      l,
	}
}

// GetToken returns a valid access token for the given carrier, refreshing if needed.
func (m *OAuthTokenManager) GetToken(ctx context.Context, carrierName string, cfg OAuthConfig) (string, error) {
	m.mu.RLock()
	if tok, ok := m.tokens[carrierName]; ok && !tok.IsExpired() {
		m.mu.RUnlock()
		return tok.AccessToken, nil
	}
	m.mu.RUnlock()

	return m.refreshToken(ctx, carrierName, cfg)
}

func (m *OAuthTokenManager) refreshToken(ctx context.Context, carrierName string, cfg OAuthConfig) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if tok, ok := m.tokens[carrierName]; ok && !tok.IsExpired() {
		return tok.AccessToken, nil
	}

	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		return "", fmt.Errorf("missing OAuth credentials for carrier %s", carrierName)
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", cfg.ClientID)
	data.Set("client_secret", cfg.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create token request for %s: %w", carrierName, err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request failed for %s: %w", carrierName, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read token response for %s: %w", carrierName, err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request for %s returned %d: %s", carrierName, resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse token response for %s: %w", carrierName, err)
	}

	m.tokens[carrierName] = &OAuthToken{
		AccessToken: tokenResp.AccessToken,
		ExpiresAt:   time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}

	m.l.WithField("carrier", carrierName).Info("refreshed OAuth token")
	return tokenResp.AccessToken, nil
}
