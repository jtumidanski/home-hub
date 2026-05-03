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

// ValidateRecurrence enforces the §4.6 PRD checks on every "RRULE:" entry of
// the slice. Non-RRULE components (EXDATE, RDATE, etc.) are ignored. Returns
// the first failure, or nil if every RRULE line is bounded and within range.
// If eventStart.IsZero(), the 5-year window check is skipped (the count and
// unbounded checks still run); this is used by the update handler where the
// start time may not be supplied.
func ValidateRecurrence(recurrence []string, eventStart time.Time) *RecurrenceError {
	for _, line := range recurrence {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(strings.ToUpper(trimmed), "RRULE:") {
			continue
		}
		until, count, hasUntil, hasCount, err := parseRRULE(trimmed)
		if err != nil || (!hasUntil && !hasCount) {
			return &RecurrenceError{
				Code:    codeUnbounded,
				Detail:  "Recurring events must specify an end date (UNTIL=) or occurrence count (COUNT=)",
				RuleRaw: trimmed,
			}
		}
		if hasCount && (count < minOccurrences || count > maxOccurrences) {
			return &RecurrenceError{
				Code:    codeCountRange,
				Detail:  "COUNT must be between 1 and 730",
				RuleRaw: trimmed,
			}
		}
		if hasUntil && !eventStart.IsZero() {
			if until.Sub(eventStart) > maxUntilWindow {
				return &RecurrenceError{
					Code:    codeTooLong,
					Detail:  "UNTIL must be no more than 5 years after the event start",
					RuleRaw: trimmed,
				}
			}
		}
	}
	return nil
}
