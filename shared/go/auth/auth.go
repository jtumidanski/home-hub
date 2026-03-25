// Package auth provides JWT validation, JWKS-based key resolution,
// and authentication middleware for downstream services.
package auth

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/logging"
	"github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus"
)

// Claims represents the JWT claims used by Home Hub.
type Claims struct {
	jwt.RegisteredClaims
	UserID      uuid.UUID `json:"user_id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	HouseholdID uuid.UUID `json:"household_id"`
}

// Validator validates JWTs against a JWKS endpoint.
type Validator struct {
	jwksURL    string
	keySet     *cachedKeySet
	logger     *logrus.Logger
}

type cachedKeySet struct {
	mu      sync.RWMutex
	keys    map[string]interface{}
	fetched time.Time
	ttl     time.Duration
}

// NewValidator creates a JWT validator that fetches keys from the given JWKS URL.
func NewValidator(l *logrus.Logger, jwksURL string) *Validator {
	return &Validator{
		jwksURL: jwksURL,
		logger:  l,
		keySet: &cachedKeySet{
			keys: make(map[string]interface{}),
			ttl:  5 * time.Minute,
		},
	}
}

// Validate parses and validates a JWT string, returning the claims.
func (v *Validator) Validate(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, v.keyFunc)
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func (v *Validator) keyFunc(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
		return nil, errors.New("unexpected signing method")
	}

	kid, ok := token.Header["kid"].(string)
	if !ok {
		return nil, errors.New("missing kid in token header")
	}

	key, err := v.getKey(kid)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (v *Validator) getKey(kid string) (interface{}, error) {
	v.keySet.mu.RLock()
	if key, ok := v.keySet.keys[kid]; ok && time.Since(v.keySet.fetched) < v.keySet.ttl {
		v.keySet.mu.RUnlock()
		return key, nil
	}
	v.keySet.mu.RUnlock()

	if err := v.refreshKeys(); err != nil {
		return nil, err
	}

	v.keySet.mu.RLock()
	defer v.keySet.mu.RUnlock()
	key, ok := v.keySet.keys[kid]
	if !ok {
		return nil, errors.New("key not found in JWKS")
	}
	return key, nil
}

func (v *Validator) refreshKeys() error {
	return FetchJWKS(v.jwksURL, v.keySet)
}

// Middleware returns HTTP middleware that validates JWT from cookies
// and injects tenant context.
func Middleware(l *logrus.Logger, v *Validator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("access_token")
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			claims, err := v.Validate(cookie.Value)
			if err != nil {
				l.WithError(err).Warn("invalid JWT")
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			tenantID := claims.TenantID
			householdID := claims.HouseholdID

			// Allow header overrides when JWT claims have nil UUIDs (e.g., during onboarding).
			if tenantID == uuid.Nil {
				if h := r.Header.Get("X-Tenant-ID"); h != "" {
					if parsed, err := uuid.Parse(h); err == nil {
						tenantID = parsed
					}
				}
			}
			if householdID == uuid.Nil {
				if h := r.Header.Get("X-Household-ID"); h != "" {
					if parsed, err := uuid.Parse(h); err == nil {
						householdID = parsed
					}
				}
			}

			ctx := r.Context()
			ctx = tenant.WithContext(ctx, tenant.New(tenantID, householdID, claims.UserID))
			ctx = logging.WithField(ctx, "user_id", claims.UserID.String())
			ctx = logging.WithField(ctx, "tenant_id", tenantID.String())
			ctx = logging.WithField(ctx, "household_id", householdID.String())
			ctx = context.WithValue(ctx, claimsKey, claims)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type ctxKey string

const claimsKey ctxKey = "claims"

// ClaimsFromContext retrieves JWT claims from the request context.
func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	c, ok := ctx.Value(claimsKey).(*Claims)
	return c, ok
}
