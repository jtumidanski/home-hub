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

func mustTime(t *testing.T, s string) time.Time {
	t.Helper()
	v, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatalf("bad time %q: %v", s, err)
	}
	return v
}

func TestValidateRecurrence(t *testing.T) {
	start := mustTime(t, "2026-05-06T09:00:00Z")

	tests := []struct {
		name    string
		input   []string
		start   time.Time
		wantNil bool
		wantCode string
	}{
		{name: "nil slice", input: nil, start: start, wantNil: true},
		{name: "empty slice", input: []string{}, start: start, wantNil: true},
		{name: "valid until", input: []string{"RRULE:FREQ=WEEKLY;UNTIL=20260611T035959Z"}, start: start, wantNil: true},
		{name: "valid count", input: []string{"RRULE:FREQ=DAILY;COUNT=5"}, start: start, wantNil: true},
		{name: "open-ended", input: []string{"RRULE:FREQ=WEEKLY"}, start: start, wantCode: codeUnbounded},
		{name: "until > 5y", input: []string{"RRULE:FREQ=WEEKLY;UNTIL=20320101T000000Z"}, start: start, wantCode: codeTooLong},
		{name: "count zero", input: []string{"RRULE:FREQ=WEEKLY;COUNT=0"}, start: start, wantCode: codeCountRange},
		{name: "count 731", input: []string{"RRULE:FREQ=WEEKLY;COUNT=731"}, start: start, wantCode: codeCountRange},
		{name: "case-insensitive", input: []string{"rrule:freq=weekly;until=20260601t000000z"}, start: start, wantNil: true},
		{name: "until date-only ok", input: []string{"RRULE:FREQ=WEEKLY;UNTIL=20260601"}, start: start, wantNil: true},
		{name: "EXDATE alongside RRULE only", input: []string{"EXDATE:20260513T090000Z", "RRULE:FREQ=WEEKLY;COUNT=5"}, start: start, wantNil: true},
		{name: "malformed UNTIL -> unbounded", input: []string{"RRULE:FREQ=WEEKLY;UNTIL=garbage"}, start: start, wantCode: codeUnbounded},
		{name: "zero start skips too-long check", input: []string{"RRULE:FREQ=WEEKLY;UNTIL=20990101T000000Z"}, start: time.Time{}, wantNil: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateRecurrence(tc.input, tc.start)
			if tc.wantNil {
				if err != nil {
					t.Fatalf("expected nil, got %+v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error code %q, got nil", tc.wantCode)
			}
			if err.Code != tc.wantCode {
				t.Fatalf("code = %q, want %q", err.Code, tc.wantCode)
			}
		})
	}
}

func TestValidateRecurrence_HandlerScenarios(t *testing.T) {
	start := mustTime(t, "2026-05-06T09:00:00-04:00")

	cases := []struct {
		name string
		rule string
		want string
	}{
		{"weekly may6 to jun10 inclusive", "RRULE:FREQ=WEEKLY;UNTIL=20260611T035959Z", ""},
		{"daily count 5", "RRULE:FREQ=DAILY;COUNT=5", ""},
		{"open-ended weekly", "RRULE:FREQ=WEEKLY", codeUnbounded},
		{"until 5y+2d", "RRULE:FREQ=WEEKLY;UNTIL=20310509T000000Z", codeTooLong},
		{"count 731", "RRULE:FREQ=WEEKLY;COUNT=731", codeCountRange},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateRecurrence([]string{tc.rule}, start)
			if tc.want == "" {
				if err != nil {
					t.Fatalf("expected nil, got %+v", err)
				}
				return
			}
			if err == nil || err.Code != tc.want {
				t.Fatalf("got %+v, want code %q", err, tc.want)
			}
		})
	}
}
