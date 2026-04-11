package tz

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func TestResolve_HeaderValidWins(t *testing.T) {
	l, _ := test.NewNullLogger()
	h := http.Header{}
	h.Set("X-Timezone", "America/New_York")

	loc := Resolve(context.Background(), l, h, uuid.Nil, func(ctx context.Context, id uuid.UUID) (string, error) {
		t.Fatal("lookup should not be called when header is valid")
		return "", nil
	})

	assert.Equal(t, "America/New_York", loc.String())
}

func TestResolve_HeaderInvalidFallsThroughToHousehold(t *testing.T) {
	l, _ := test.NewNullLogger()
	h := http.Header{}
	h.Set("X-Timezone", "Mars/Olympus")

	loc := Resolve(context.Background(), l, h, uuid.New(), func(ctx context.Context, id uuid.UUID) (string, error) {
		return "Europe/Berlin", nil
	})

	assert.Equal(t, "Europe/Berlin", loc.String())
}

func TestResolve_HouseholdFallback(t *testing.T) {
	l, _ := test.NewNullLogger()

	loc := Resolve(context.Background(), l, http.Header{}, uuid.New(), func(ctx context.Context, id uuid.UUID) (string, error) {
		return "Asia/Tokyo", nil
	})

	assert.Equal(t, "Asia/Tokyo", loc.String())
}

func TestResolve_UTCFallbackWhenLookupErrors(t *testing.T) {
	l, _ := test.NewNullLogger()

	loc := Resolve(context.Background(), l, http.Header{}, uuid.New(), func(ctx context.Context, id uuid.UUID) (string, error) {
		return "", errors.New("boom")
	})

	assert.Equal(t, time.UTC, loc)
}

func TestResolve_UTCFallbackWhenNoLookup(t *testing.T) {
	l, _ := test.NewNullLogger()

	loc := Resolve(context.Background(), l, http.Header{}, uuid.Nil, nil)

	assert.Equal(t, time.UTC, loc)
}

func TestResolve_CachesOnContext(t *testing.T) {
	l, _ := test.NewNullLogger()
	fixed, _ := time.LoadLocation("Australia/Sydney")
	ctx := WithLocation(context.Background(), fixed)

	h := http.Header{}
	h.Set("X-Timezone", "America/New_York")
	loc := Resolve(ctx, l, h, uuid.Nil, nil)

	assert.Equal(t, fixed, loc, "cached context location must take precedence")
}
