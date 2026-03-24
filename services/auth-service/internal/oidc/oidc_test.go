package oidc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDiscover(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Discovery{
			AuthorizationEndpoint: "https://example.com/auth",
			TokenEndpoint:         "https://example.com/token",
			UserinfoEndpoint:      "https://example.com/userinfo",
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	disc, err := Discover(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if disc.AuthorizationEndpoint != "https://example.com/auth" {
		t.Errorf("expected auth endpoint, got %s", disc.AuthorizationEndpoint)
	}
	if disc.TokenEndpoint != "https://example.com/token" {
		t.Errorf("expected token endpoint, got %s", disc.TokenEndpoint)
	}
	if disc.UserinfoEndpoint != "https://example.com/userinfo" {
		t.Errorf("expected userinfo endpoint, got %s", disc.UserinfoEndpoint)
	}
}

func TestDiscover_InvalidURL(t *testing.T) {
	_, err := Discover("http://127.0.0.1:1")
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

func TestAuthURL(t *testing.T) {
	disc := &Discovery{
		AuthorizationEndpoint: "https://example.com/auth",
	}
	cfg := ProviderConfig{
		ClientID:    "client-123",
		RedirectURL: "https://app.example.com/callback",
	}

	url := AuthURL(disc, cfg, "test-state")
	if url == "" {
		t.Fatal("expected non-empty URL")
	}
	// Verify key parameters are present
	for _, param := range []string{"response_type=code", "client_id=client-123", "state=test-state"} {
		if !contains(url, param) {
			t.Errorf("expected URL to contain %s, got %s", param, url)
		}
	}
}

func TestExchangeCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		json.NewEncoder(w).Encode(TokenResponse{
			AccessToken: "access-123",
			IDToken:     "id-token-123",
			TokenType:   "Bearer",
		})
	}))
	defer server.Close()

	disc := &Discovery{TokenEndpoint: server.URL}
	cfg := ProviderConfig{
		ClientID:     "client-123",
		ClientSecret: "secret-456",
		RedirectURL:  "https://app.example.com/callback",
	}

	resp, err := ExchangeCode(context.Background(), disc, cfg, "auth-code")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AccessToken != "access-123" {
		t.Errorf("expected access-123, got %s", resp.AccessToken)
	}
}

func TestExchangeCode_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	disc := &Discovery{TokenEndpoint: server.URL}
	cfg := ProviderConfig{ClientID: "c", ClientSecret: "s", RedirectURL: "https://r"}

	_, err := ExchangeCode(context.Background(), disc, cfg, "bad-code")
	if err == nil {
		t.Error("expected error for server error response")
	}
}

func TestFetchUserInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("expected Bearer test-token, got %s", auth)
		}
		json.NewEncoder(w).Encode(map[string]string{
			"sub":         "sub-123",
			"email":       "user@example.com",
			"name":        "Test User",
			"given_name":  "Test",
			"family_name": "User",
			"picture":     "https://example.com/pic.jpg",
		})
	}))
	defer server.Close()

	disc := &Discovery{UserinfoEndpoint: server.URL}
	info, err := FetchUserInfo(context.Background(), disc, "test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Subject != "sub-123" {
		t.Errorf("expected sub-123, got %s", info.Subject)
	}
	if info.Email != "user@example.com" {
		t.Errorf("expected user@example.com, got %s", info.Email)
	}
	if info.DisplayName != "Test User" {
		t.Errorf("expected Test User, got %s", info.DisplayName)
	}
	if info.GivenName != "Test" {
		t.Errorf("expected Test, got %s", info.GivenName)
	}
	if info.FamilyName != "User" {
		t.Errorf("expected User, got %s", info.FamilyName)
	}
	if info.AvatarURL != "https://example.com/pic.jpg" {
		t.Errorf("expected avatar URL, got %s", info.AvatarURL)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
