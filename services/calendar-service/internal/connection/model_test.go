package connection

import (
	"testing"
)

func TestUserColors(t *testing.T) {
	tests := []struct {
		name     string
		validate func(t *testing.T)
	}{
		{
			"palette has eight colors",
			func(t *testing.T) {
				if len(UserColors) != 8 {
					t.Fatalf("expected 8 colors, got %d", len(UserColors))
				}
			},
		},
		{
			"all colors are valid hex",
			func(t *testing.T) {
				for i, color := range UserColors {
					if color == "" {
						t.Fatalf("empty color at index %d", i)
					}
					if color[0] != '#' || len(color) != 7 {
						t.Fatalf("invalid color format at index %d: %q", i, color)
					}
				}
			},
		},
		{
			"colors wrap around after palette length",
			func(t *testing.T) {
				for i := 0; i < len(UserColors)*2; i++ {
					color := UserColors[i%len(UserColors)]
					expected := UserColors[i%len(UserColors)]
					if color != expected {
						t.Fatalf("unexpected color at index %d", i)
					}
				}
				if UserColors[0%len(UserColors)] != UserColors[8%len(UserColors)] {
					t.Fatal("colors should wrap around after palette length")
				}
			},
		},
		{
			"tenth index wraps to third color",
			func(t *testing.T) {
				idx := 10
				color := UserColors[idx%len(UserColors)]
				expected := UserColors[2]
				if color != expected {
					t.Fatalf("expected %q, got %q", expected, color)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.validate(t)
		})
	}
}
