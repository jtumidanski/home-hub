package http

import (
	"fmt"
	nethttp "net/http"
	"strings"
	"time"
)

// ParseDateParam parses a YYYY-MM-DD query parameter and returns the date
// anchored to midnight UTC. The returned time.Time represents a calendar day,
// not an instant in time — downstream queries against `type:date` GORM columns
// compare by date irrespective of timezone.
//
// Returns an error naming the parameter and the raw value when the parameter
// is missing, empty, or not a valid YYYY-MM-DD calendar date. Handlers should
// translate the error into a 400 response.
func ParseDateParam(r *nethttp.Request, name string) (time.Time, error) {
	raw := strings.TrimSpace(r.URL.Query().Get(name))
	if raw == "" {
		return time.Time{}, fmt.Errorf("query parameter %q is required and must be YYYY-MM-DD (got: %q)", name, raw)
	}
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return time.Time{}, fmt.Errorf("query parameter %q must be YYYY-MM-DD (got: %q)", name, raw)
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
}
