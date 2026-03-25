package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ProviderConfig holds the configuration for an OIDC provider.
type ProviderConfig struct {
	Name         string
	IssuerURL    string
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// UserInfo holds the user information returned from the OIDC provider.
type UserInfo struct {
	Subject     string
	Email       string
	DisplayName string
	GivenName   string
	FamilyName  string
	AvatarURL   string
}

// Discovery holds the OIDC discovery document endpoints.
type Discovery struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserinfoEndpoint      string `json:"userinfo_endpoint"`
}

// Discover fetches the OIDC discovery document.
func Discover(issuerURL string) (*Discovery, error) {
	wellKnown := strings.TrimSuffix(issuerURL, "/") + "/.well-known/openid-configuration"
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(wellKnown)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch discovery: %w", err)
	}
	defer resp.Body.Close()

	var doc Discovery
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, fmt.Errorf("failed to parse discovery: %w", err)
	}
	return &doc, nil
}

// AuthURL builds the authorization redirect URL.
func AuthURL(disc *Discovery, cfg ProviderConfig, state string) string {
	params := url.Values{
		"response_type": {"code"},
		"client_id":     {cfg.ClientID},
		"redirect_uri":  {cfg.RedirectURL},
		"scope":         {"openid email profile"},
		"state":         {state},
	}
	return disc.AuthorizationEndpoint + "?" + params.Encode()
}

// TokenResponse holds the token endpoint response.
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	IDToken     string `json:"id_token"`
	TokenType   string `json:"token_type"`
}

// ExchangeCode exchanges an authorization code for tokens.
func ExchangeCode(ctx context.Context, disc *Discovery, cfg ProviderConfig, code string) (*TokenResponse, error) {
	data := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {cfg.RedirectURL},
		"client_id":     {cfg.ClientID},
		"client_secret": {cfg.ClientSecret},
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.PostForm(disc.TokenEndpoint, data)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token endpoint returned %d", resp.StatusCode)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}
	return &tokenResp, nil
}

// FetchUserInfo fetches user information from the userinfo endpoint.
func FetchUserInfo(ctx context.Context, disc *Discovery, accessToken string) (*UserInfo, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, disc.UserinfoEndpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("userinfo request failed: %w", err)
	}
	defer resp.Body.Close()

	var raw struct {
		Sub        string `json:"sub"`
		Email      string `json:"email"`
		Name       string `json:"name"`
		GivenName  string `json:"given_name"`
		FamilyName string `json:"family_name"`
		Picture    string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("failed to parse userinfo: %w", err)
	}

	return &UserInfo{
		Subject:     raw.Sub,
		Email:       raw.Email,
		DisplayName: raw.Name,
		GivenName:   raw.GivenName,
		FamilyName:  raw.FamilyName,
		AvatarURL:   raw.Picture,
	}, nil
}
