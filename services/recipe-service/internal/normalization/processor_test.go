package normalization

import (
	"testing"
)

func TestNormalizeText(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"strips leading 'the'", "the garlic", "garlic"},
		{"strips leading 'a'", "a tomato", "tomato"},
		{"strips leading 'an'", "an onion", "onion"},
		{"removes trailing s", "tomatoes", "tomatoe"},
		{"does not remove trailing s from double-s", "grass", "grass"},
		{"collapses whitespace", "red   bell   pepper", "red bell pepper"},
		{"strips article and collapses whitespace", "the  fresh  tomatoes", "fresh tomatoe"},
		{"no change for simple word", "garlic", "garlic"},
		{"empty string remains empty", "", ""},
		{"single character a preserved", "a", "a"},
		{"single s not stripped from short word", "ss", "ss"},
		{"trims trailing space", "garlic ", "garlic"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeText(tt.input)
			if got != tt.want {
				t.Errorf("normalizeText(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidStatus(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"matched", true},
		{"alias_matched", true},
		{"unresolved", true},
		{"manually_confirmed", true},
		{"", false},
		{"pending", false},
		{"resolved", false},
		{"MATCHED", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ValidStatus(tt.input)
			if got != tt.want {
				t.Errorf("ValidStatus(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
