package export

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseQuantity(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want float64
	}{
		{"empty string returns 0", "", 0},
		{"valid integer", "3", 3.0},
		{"valid float", "1.5", 1.5},
		{"zero", "0", 0},
		{"negative", "-1", -1.0},
		{"invalid string returns 0", "abc", 0},
		{"mixed string extracts leading number", "1a", 1},
		{"complex description extracts leading number", "3 small-medium, cut into matchsticks", 3},
		{"fraction string", "1/2", 0.5},
		{"mixed number", "2 1/4", 2.25},
		{"additive expression", "1 + 1", 2},
		{"to taste returns 0", "to taste", 0},
		{"large number", "1000.25", 1000.25},
		{"leading zero", "0.5", 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, parseQuantity(tt.raw))
		})
	}
}

func TestFormatQuantity(t *testing.T) {
	tests := []struct {
		name string
		val  float64
		want string
	}{
		{"whole number shows no decimal", 2.0, "2"},
		{"fraction shows one decimal", 1.5, "1.5"},
		{"zero", 0.0, "0"},
		{"large whole", 100.0, "100"},
		{"one third rounds", 0.3333333, "0.3"},
		{"exactly point five", 2.5, "2.5"},
		{"negative whole", -3.0, "-3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, formatQuantity(tt.val))
		})
	}
}

func TestCapitalizeFirst(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty string", "", ""},
		{"lowercase", "breakfast", "Breakfast"},
		{"already capitalized", "Lunch", "Lunch"},
		{"single char", "d", "D"},
		{"all caps unchanged after first", "dINNER", "DINNER"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, capitalizeFirst(tt.input))
		})
	}
}
