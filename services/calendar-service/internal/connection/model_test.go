package connection

import (
	"testing"
)

func TestColorAssignment(t *testing.T) {
	for i := 0; i < len(UserColors)*2; i++ {
		color := UserColors[i%len(UserColors)]
		if color == "" {
			t.Fatalf("empty color at index %d", i)
		}
		if color[0] != '#' || len(color) != 7 {
			t.Fatalf("invalid color format at index %d: %q", i, color)
		}
	}
}

func TestColorWraps(t *testing.T) {
	idx := 10
	color := UserColors[idx%len(UserColors)]
	expected := UserColors[2] // 10 % 8 = 2
	if color != expected {
		t.Fatalf("expected %q, got %q", expected, color)
	}

	// First and ninth connections should get the same color
	if UserColors[0%len(UserColors)] != UserColors[8%len(UserColors)] {
		t.Fatal("colors should wrap around after palette length")
	}
}

func TestPaletteHasEightColors(t *testing.T) {
	if len(UserColors) != 8 {
		t.Fatalf("expected 8 colors, got %d", len(UserColors))
	}
}
