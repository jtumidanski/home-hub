package performance

import (
	"testing"
)

// Table-driven coverage of the §4.4.1 status state machine. The processor
// implements the transitions in two helpers (`applyExplicitStatus` for
// explicit requests, `deriveStatusFromActuals` for the auto path); we test
// both directly so the rules don't have to round-trip through a real DB.
func TestApplyExplicitStatus(t *testing.T) {
	cases := []struct {
		name       string
		prev       string
		requested  string
		hasActuals bool
		want       string
	}{
		{"pending->done", StatusPending, StatusDone, false, StatusDone},
		{"pending->skipped", StatusPending, StatusSkipped, false, StatusSkipped},
		{"partial->done", StatusPartial, StatusDone, true, StatusDone},
		{"partial->skipped clears actuals", StatusPartial, StatusSkipped, true, StatusSkipped},
		{"done->skipped", StatusDone, StatusSkipped, true, StatusSkipped},
		{"done->pending with actuals -> partial", StatusDone, StatusPending, true, StatusPartial},
		{"done->pending without actuals -> pending", StatusDone, StatusPending, false, StatusPending},
		{"skipped->pending unskip", StatusSkipped, StatusPending, false, StatusPending},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := applyExplicitStatus(tc.prev, tc.requested, tc.hasActuals)
			if got != tc.want {
				t.Fatalf("applyExplicitStatus(%s,%s,%v) = %s, want %s", tc.prev, tc.requested, tc.hasActuals, got, tc.want)
			}
		})
	}
}

func TestDeriveStatusFromActuals(t *testing.T) {
	cases := []struct {
		prev string
		want string
	}{
		{StatusPending, StatusPartial}, // pending → partial
		{StatusSkipped, StatusPartial}, // skip is cleared when actuals arrive
		{StatusPartial, StatusPartial}, // stays
		{StatusDone, StatusDone},       // user is correcting, not retracting
	}
	for _, tc := range cases {
		got := deriveStatusFromActuals(tc.prev)
		if got != tc.want {
			t.Fatalf("deriveStatusFromActuals(%s) = %s, want %s", tc.prev, got, tc.want)
		}
	}
}

func TestHasActuals(t *testing.T) {
	v := 3
	in := PatchInput{ActualSets: &v}
	if !in.hasActuals() {
		t.Fatal("expected hasActuals true with sets set")
	}
	if (PatchInput{}).hasActuals() {
		t.Fatal("expected hasActuals false on empty input")
	}
}
