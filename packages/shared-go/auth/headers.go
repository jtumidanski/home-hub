package auth

import (
	"errors"
	"net/http"
	"strings"
)

// Header names passed by oauth2-proxy
const (
	HeaderEmail       = "X-Auth-Request-Email"
	HeaderUser        = "X-Auth-Request-User"
	HeaderGroups      = "X-Auth-Request-Groups"
	HeaderAccessToken = "X-Auth-Request-Access-Token"
	HeaderForwardedBy = "X-Forwarded-By"
)

var (
	ErrMissingEmailHeader = errors.New("missing X-Auth-Request-Email header")
	ErrEmptyEmail         = errors.New("empty email in X-Auth-Request-Email header")
	ErrInvalidSource      = errors.New("request not from trusted ingress")
)

// Headers contains the extracted authentication headers
type Headers struct {
	Email       string
	User        string
	Groups      []string
	AccessToken string
}

// ExtractHeaders extracts authentication headers from the HTTP request
func ExtractHeaders(r *http.Request) (Headers, error) {
	email := strings.TrimSpace(r.Header.Get(HeaderEmail))
	if email == "" {
		return Headers{}, ErrMissingEmailHeader
	}

	user := strings.TrimSpace(r.Header.Get(HeaderUser))
	if user == "" {
		// Default to email if user header is missing
		user = email
	}

	// Parse groups (comma-separated)
	groupsHeader := strings.TrimSpace(r.Header.Get(HeaderGroups))
	var groups []string
	if groupsHeader != "" {
		for _, group := range strings.Split(groupsHeader, ",") {
			trimmed := strings.TrimSpace(group)
			if trimmed != "" {
				groups = append(groups, trimmed)
			}
		}
	}

	accessToken := strings.TrimSpace(r.Header.Get(HeaderAccessToken))

	return Headers{
		Email:       email,
		User:        user,
		Groups:      groups,
		AccessToken: accessToken,
	}, nil
}

// ValidateSource checks if the request came from a trusted ingress
// This prevents clients from directly sending auth headers to bypass authentication
func ValidateSource(r *http.Request) error {
	// Check for the X-Forwarded-By header set by nginx
	forwardedBy := r.Header.Get(HeaderForwardedBy)
	if forwardedBy != "nginx-ingress" {
		return ErrInvalidSource
	}

	// Additional checks can be added here:
	// - Check X-Forwarded-For for trusted IPs
	// - Check for internal network ranges
	// - Verify TLS client certificates

	return nil
}

// InferProvider attempts to determine the OAuth provider from email or other headers
// Returns "google" or "github" based on heuristics, defaults to "google"
func InferProvider(headers Headers) string {
	// Check if email domain suggests a provider
	if strings.HasSuffix(strings.ToLower(headers.Email), "@gmail.com") ||
		strings.HasSuffix(strings.ToLower(headers.Email), "@googlemail.com") {
		return "google"
	}

	// GitHub emails often use noreply addresses
	if strings.Contains(strings.ToLower(headers.Email), "@users.noreply.github.com") {
		return "github"
	}

	// Check groups for provider hints
	for _, group := range headers.Groups {
		if strings.HasPrefix(strings.ToLower(group), "github:") {
			return "github"
		}
	}

	// Default to google
	// In a production system, you might track this more explicitly
	// by having separate oauth2-proxy instances report provider info
	return "google"
}
