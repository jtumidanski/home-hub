// Package tz resolves the effective IANA timezone for an incoming request and
// caches the result on the request context so downstream callers within the
// same request share one *time.Location.
//
// Resolution order:
//  1. X-Timezone header (parsed via time.LoadLocation).
//  2. Household timezone looked up via the injected HouseholdLookup callback.
//  3. UTC fallback, with a warn-level log.
package tz

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type contextKey struct{}

var locKey = contextKey{}

// HouseholdLookup resolves a household's configured timezone (IANA identifier)
// given its id. Implementations typically call account-service.
type HouseholdLookup func(ctx context.Context, householdID uuid.UUID) (string, error)

// WithLocation stores the resolved location on the context.
func WithLocation(ctx context.Context, loc *time.Location) context.Context {
	return context.WithValue(ctx, locKey, loc)
}

// FromContext returns the cached location, if any.
func FromContext(ctx context.Context) (*time.Location, bool) {
	loc, ok := ctx.Value(locKey).(*time.Location)
	return loc, ok
}

// Resolve determines the effective *time.Location for a request. It never
// returns nil and never errors — on all failure paths it falls back to UTC.
func Resolve(ctx context.Context, l logrus.FieldLogger, headers http.Header, householdID uuid.UUID, lookup HouseholdLookup) *time.Location {
	if loc, ok := FromContext(ctx); ok && loc != nil {
		return loc
	}

	if hdr := headers.Get("X-Timezone"); hdr != "" {
		if loc, err := time.LoadLocation(hdr); err == nil {
			return loc
		}
		l.WithField("header", hdr).Warn("invalid X-Timezone header, trying household fallback")
	}

	if lookup != nil {
		tz, err := lookup(ctx, householdID)
		switch {
		case err != nil:
			l.WithError(err).Warn("household timezone lookup failed, falling back to UTC")
		case tz == "":
			l.Warn("household timezone empty, falling back to UTC")
		default:
			if loc, err := time.LoadLocation(tz); err == nil {
				return loc
			}
			l.WithField("household_tz", tz).Warn("invalid household timezone, falling back to UTC")
		}
	} else {
		l.Warn("no household lookup configured, falling back to UTC")
	}

	return time.UTC
}
