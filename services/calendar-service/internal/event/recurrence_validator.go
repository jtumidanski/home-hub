package event

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

const (
	maxUntilWindow = 5*365*24*time.Hour + 24*time.Hour
	minOccurrences = 1
	maxOccurrences = 730

	codeUnbounded  = "recurrence_unbounded"
	codeTooLong    = "recurrence_too_long"
	codeCountRange = "recurrence_count_out_of_range"
)

type RecurrenceError struct {
	Code    string
	Detail  string
	RuleRaw string
}

func (e *RecurrenceError) Error() string { return e.Code + ": " + e.Detail }

// parseRRULE extracts UNTIL and COUNT from a single "RRULE:..." line.
// Component names are matched case-insensitively per RFC 5545. UNTIL
// accepts both the date form (YYYYMMDD) and the date-time UTC form
// (YYYYMMDDTHHMMSSZ).
func parseRRULE(line string) (until time.Time, count int, hasUntil, hasCount bool, err error) {
	upper := strings.ToUpper(line)
	if !strings.HasPrefix(upper, "RRULE:") {
		return time.Time{}, 0, false, false, errors.New("not an RRULE line")
	}
	body := line[len("RRULE:"):]
	for _, kv := range strings.Split(body, ";") {
		eq := strings.IndexByte(kv, '=')
		if eq < 0 {
			continue
		}
		key := strings.ToUpper(strings.TrimSpace(kv[:eq]))
		val := strings.TrimSpace(kv[eq+1:])
		switch key {
		case "UNTIL":
			t, perr := parseUntil(val)
			if perr != nil {
				return time.Time{}, 0, false, false, perr
			}
			until, hasUntil = t, true
		case "COUNT":
			n, perr := strconv.Atoi(val)
			if perr != nil {
				return time.Time{}, 0, false, false, perr
			}
			count, hasCount = n, true
		}
	}
	return
}

func parseUntil(s string) (time.Time, error) {
	upper := strings.ToUpper(strings.TrimSpace(s))
	if t, err := time.Parse("20060102T150405Z", upper); err == nil {
		return t, nil
	}
	if t, err := time.Parse("20060102", upper); err == nil {
		return t, nil
	}
	return time.Time{}, errors.New("invalid UNTIL value: " + s)
}
