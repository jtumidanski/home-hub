package weathercode

import "testing"

func TestLookup(t *testing.T) {
	tests := []struct {
		code        int
		wantSummary string
		wantIcon    string
	}{
		{0, "Clear", "sun"},
		{1, "Mostly Clear", "sun"},
		{2, "Partly Cloudy", "cloud-sun"},
		{3, "Overcast", "cloud"},
		{45, "Fog", "cloud-fog"},
		{51, "Drizzle", "cloud-drizzle"},
		{61, "Rain", "cloud-rain"},
		{66, "Freezing Rain", "cloud-rain"},
		{71, "Snow", "snowflake"},
		{77, "Snow Grains", "snowflake"},
		{80, "Rain Showers", "cloud-rain"},
		{85, "Snow Showers", "snowflake"},
		{95, "Thunderstorm", "cloud-lightning"},
		{96, "Thunderstorm with Hail", "cloud-lightning"},
		{99, "Thunderstorm with Hail", "cloud-lightning"},
	}

	for _, tt := range tests {
		t.Run(tt.wantSummary, func(t *testing.T) {
			summary, icon := Lookup(tt.code)
			if summary != tt.wantSummary {
				t.Errorf("Lookup(%d) summary = %q, want %q", tt.code, summary, tt.wantSummary)
			}
			if icon != tt.wantIcon {
				t.Errorf("Lookup(%d) icon = %q, want %q", tt.code, icon, tt.wantIcon)
			}
		})
	}
}

func TestLookupUnknown(t *testing.T) {
	summary, icon := Lookup(999)
	if summary != "Unknown" {
		t.Errorf("expected Unknown, got %q", summary)
	}
	if icon != "cloud" {
		t.Errorf("expected cloud, got %q", icon)
	}
}
