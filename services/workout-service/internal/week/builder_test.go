package week

import (
	"testing"
	"time"
)

func TestNormalizeToMonday(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{"2026-04-06", "2026-04-06"}, // Mon stays Mon
		{"2026-04-08", "2026-04-06"}, // Wed → Mon
		{"2026-04-12", "2026-04-06"}, // Sun → Mon
		{"2026-04-13", "2026-04-13"}, // next Mon
	}
	for _, tc := range cases {
		in, _ := time.Parse("2006-01-02", tc.in)
		got := NormalizeToMonday(in).Format("2006-01-02")
		if got != tc.out {
			t.Fatalf("NormalizeToMonday(%s) = %s, want %s", tc.in, got, tc.out)
		}
	}
}

func TestValidateRestDayFlags(t *testing.T) {
	if err := ValidateRestDayFlags([]int{0, 1, 6}); err != nil {
		t.Fatalf("expected valid flags to pass, got %v", err)
	}
	if err := ValidateRestDayFlags([]int{7}); err == nil {
		t.Fatal("expected day 7 to fail")
	}
	if err := ValidateRestDayFlags([]int{-1}); err == nil {
		t.Fatal("expected day -1 to fail")
	}
}
