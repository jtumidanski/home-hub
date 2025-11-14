package user

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// contextKey is a private type for context keys to avoid collisions
type contextKey string

const userInfoKey contextKey = "userInfo"

// UserResolverMiddleware creates middleware that resolves the current user information
// by calling the users service /api/me endpoint
func UserResolverMiddleware(logger *logrus.Logger, client *Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract auth headers that need to be forwarded
			authHeaders := make(map[string]string)

			// Headers set by nginx oauth2-proxy
			headerNames := []string{
				"X-Auth-Request-Email",
				"X-Auth-Request-User",
				"X-Auth-Request-Groups",
				"X-Auth-Request-Access-Token",
				"X-Forwarded-By",
			}

			for _, headerName := range headerNames {
				if value := r.Header.Get(headerName); value != "" {
					authHeaders[headerName] = value
				}
			}

			// If no auth headers present, continue without user context
			// (allows health checks and other unauthenticated requests)
			if len(authHeaders) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			// Call users service to resolve user information
			userInfo, err := client.GetCurrentUser(r.Context(), authHeaders)
			if err != nil {
				logger.WithError(err).Warn("Failed to resolve user from users service")
				// Continue without user context - handlers will return 401 if they need auth
				next.ServeHTTP(w, r)
				return
			}

			// Add user info to context
			ctx := context.WithValue(r.Context(), userInfoKey, userInfo)

			logger.WithFields(logrus.Fields{
				"user_id":      userInfo.UserID.String(),
				"household_id": userInfo.HouseholdID.String(),
			}).Debug("User resolved from users service")

			// Call next handler with user info in context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserInfoFromContext extracts the user info from the request context
// Returns nil and false if no user info is present
func UserInfoFromContext(ctx context.Context) (*UserInfo, bool) {
	userInfo, ok := ctx.Value(userInfoKey).(*UserInfo)
	return userInfo, ok
}

// UserIDFromContext extracts the user ID from the request context
// Returns uuid.Nil and false if no user info is present
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userInfo, ok := UserInfoFromContext(ctx)
	if !ok {
		return uuid.Nil, false
	}
	return userInfo.UserID, true
}

// HouseholdIDFromContext extracts the household ID from the request context
// Returns uuid.Nil and false if no user info is present
func HouseholdIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userInfo, ok := UserInfoFromContext(ctx)
	if !ok {
		return uuid.Nil, false
	}
	return userInfo.HouseholdID, true
}

// MustGetUserInfoFromContext extracts the user info from context or panics
// Should only be used after checking UserInfoFromContext or in protected handlers
func MustGetUserInfoFromContext(ctx context.Context) *UserInfo {
	userInfo, ok := UserInfoFromContext(ctx)
	if !ok {
		panic("user info not found in context")
	}
	return userInfo
}
