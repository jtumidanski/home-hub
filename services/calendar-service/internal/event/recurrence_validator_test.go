package event

import (
	"testing"
	"time"
)

func TestParseRRULE(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		hasUntil  bool
		untilUTC  string // RFC3339 UTC
		hasCount  bool
		count     int
		expectErr bool
	}{
		{name: "weekly until date-time", line: "RRULE:FREQ=WEEKLY;UNTIL=20260611T035959Z",
			hasUntil: true, untilUTC: "2026-06-11T03:59:59Z"},
		{name: "weekly count", line: "RRULE:FREQ=WEEKLY;COUNT=10", hasCount: true, count: 10},
		{name: "lowercase tokens", line: "rrule:freq=weekly;until=20260611t035959z",
			hasUntil: true, untilUTC: "2026-06-11T03:59:59Z"},
		{name: "until date-only form", line: "RRULE:FREQ=WEEKLY;UNTIL=20260611",
			hasUntil: true, untilUTC: "2026-06-11T00:00:00Z"},
		{name: "open-ended", line: "RRULE:FREQ=WEEKLY"},
		{name: "malformed until -> err", line: "RRULE:FREQ=WEEKLY;UNTIL=garbage", expectErr: true},
		{name: "malformed count -> err", line: "RRULE:FREQ=WEEKLY;COUNT=abc", expectErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			until, count, hasUntil, hasCount, err := parseRRULE(tc.line)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if hasUntil != tc.hasUntil {
				t.Fatalf("hasUntil = %v, want %v", hasUntil, tc.hasUntil)
			}
			if tc.hasUntil {
				want, _ := time.Parse(time.RFC3339, tc.untilUTC)
				if !until.Equal(want) {
					t.Fatalf("until = %v, want %v", until, want)
				}
			}
			if hasCount != tc.hasCount {
				t.Fatalf("hasCount = %v, want %v", hasCount, tc.hasCount)
			}
			if hasCount && count != tc.count {
				t.Fatalf("count = %d, want %d", count, tc.count)
			}
		})
	}
}
