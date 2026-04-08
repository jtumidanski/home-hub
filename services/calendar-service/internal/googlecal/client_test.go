package googlecal

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/sirupsen/logrus/hooks/test"
)

// newTestClient builds a Client whose token endpoint is overridden to point at
// the supplied httptest server. RefreshToken posts to the package-level
// tokenEndpoint constant, so we wrap the server with a transport-level redirect.
func newTestClient(t *testing.T, server *httptest.Server) *Client {
	t.Helper()
	l, _ := test.NewNullLogger()
	c := NewClient("client-id", "client-secret", l)
	// Replace the http client's transport so any request to the real token
	// endpoint is rewritten to the test server.
	c.httpClient.Transport = rewriteTransport{target: server.URL}
	return c
}

type rewriteTransport struct {
	target string
}

func (r rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	u, err := url.Parse(r.target)
	if err != nil {
		return nil, err
	}
	req.URL.Scheme = u.Scheme
	req.URL.Host = u.Host
	req.Host = u.Host
	return http.DefaultTransport.RoundTrip(req)
}

func TestRefreshToken_InvalidGrantReturnsTypedError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant","error_description":"Token has been expired or revoked."}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.RefreshToken(context.Background(), "stale-refresh")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var tre *TokenRefreshError
	if !errors.As(err, &tre) {
		t.Fatalf("expected *TokenRefreshError, got %T: %v", err, err)
	}
	if tre.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", tre.StatusCode)
	}
	if tre.OAuthError != "invalid_grant" {
		t.Errorf("expected invalid_grant, got %q", tre.OAuthError)
	}
	if !IsInvalidGrant(err) {
		t.Error("IsInvalidGrant should be true")
	}
}

func TestRefreshToken_5xxReturnsTypedErrorWithoutOAuthCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("<html>500</html>"))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.RefreshToken(context.Background(), "anything")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var tre *TokenRefreshError
	if !errors.As(err, &tre) {
		t.Fatalf("expected *TokenRefreshError, got %T: %v", err, err)
	}
	if tre.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", tre.StatusCode)
	}
	if tre.OAuthError != "" {
		t.Errorf("expected empty OAuthError, got %q", tre.OAuthError)
	}
	if IsInvalidGrant(err) {
		t.Error("IsInvalidGrant should be false for 500")
	}
}

func TestIsInvalidGrant_NilFalse(t *testing.T) {
	if IsInvalidGrant(nil) {
		t.Fatal("IsInvalidGrant(nil) should be false")
	}
}
