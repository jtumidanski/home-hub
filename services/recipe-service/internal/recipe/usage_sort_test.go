package recipe

import "testing"

func TestParseUsageSort(t *testing.T) {
	cases := []struct {
		in   string
		want UsageSort
	}{
		{"usageCount", UsageSortAsc},
		{"-usageCount", UsageSortDesc},
		{"", UsageSortNone},
		{"title", UsageSortNone},
		{"cookCount", UsageSortNone},
	}
	for _, c := range cases {
		if got := parseUsageSort(c.in); got != c.want {
			t.Errorf("parseUsageSort(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}
