package normalization

import (
	"testing"
)

func TestLookupUnit_CountFamily(t *testing.T) {
	tests := []struct {
		raw       string
		canonical string
		family    string
	}{
		{"each", "each", "count"},
		{"piece", "piece", "count"},
		{"pcs", "piece", "count"},
		{"count", "each", "count"},
		{"whole", "whole", "count"},
		{"clove", "clove", "count"},
		{"cloves", "clove", "count"},
		{"head", "head", "count"},
		{"heads", "head", "count"},
		{"bunch", "bunch", "count"},
		{"bunches", "bunch", "count"},
		{"sprig", "sprig", "count"},
		{"sprigs", "sprig", "count"},
		{"stalk", "stalk", "count"},
		{"stalks", "stalk", "count"},
		{"slice", "slice", "count"},
		{"slices", "slice", "count"},
		{"pinch", "pinch", "count"},
		{"pinches", "pinch", "count"},
		{"dash", "dash", "count"},
		{"dashes", "dash", "count"},
	}

	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			identity, ok := LookupUnit(tt.raw)
			if !ok {
				t.Fatalf("LookupUnit(%q) returned false, expected true", tt.raw)
			}
			if identity.Canonical != tt.canonical {
				t.Errorf("Canonical = %q, want %q", identity.Canonical, tt.canonical)
			}
			if identity.Family != tt.family {
				t.Errorf("Family = %q, want %q", identity.Family, tt.family)
			}
		})
	}
}

func TestLookupUnit_WeightFamily(t *testing.T) {
	tests := []struct {
		raw       string
		canonical string
		family    string
	}{
		{"g", "gram", "weight"},
		{"gram", "gram", "weight"},
		{"grams", "gram", "weight"},
		{"kg", "kilogram", "weight"},
		{"kilogram", "kilogram", "weight"},
		{"kilograms", "kilogram", "weight"},
		{"oz", "ounce", "weight"},
		{"ounce", "ounce", "weight"},
		{"ounces", "ounce", "weight"},
		{"lb", "pound", "weight"},
		{"pound", "pound", "weight"},
		{"pounds", "pound", "weight"},
	}

	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			identity, ok := LookupUnit(tt.raw)
			if !ok {
				t.Fatalf("LookupUnit(%q) returned false, expected true", tt.raw)
			}
			if identity.Canonical != tt.canonical {
				t.Errorf("Canonical = %q, want %q", identity.Canonical, tt.canonical)
			}
			if identity.Family != tt.family {
				t.Errorf("Family = %q, want %q", identity.Family, tt.family)
			}
		})
	}
}

func TestLookupUnit_VolumeFamily(t *testing.T) {
	tests := []struct {
		raw       string
		canonical string
		family    string
	}{
		{"ml", "milliliter", "volume"},
		{"milliliter", "milliliter", "volume"},
		{"milliliters", "milliliter", "volume"},
		{"l", "liter", "volume"},
		{"liter", "liter", "volume"},
		{"liters", "liter", "volume"},
		{"tsp", "teaspoon", "volume"},
		{"teaspoon", "teaspoon", "volume"},
		{"teaspoons", "teaspoon", "volume"},
		{"tbsp", "tablespoon", "volume"},
		{"tablespoon", "tablespoon", "volume"},
		{"tablespoons", "tablespoon", "volume"},
		{"cup", "cup", "volume"},
		{"cups", "cup", "volume"},
		{"fl oz", "fluid ounce", "volume"},
		{"fluid ounce", "fluid ounce", "volume"},
		{"fluid ounces", "fluid ounce", "volume"},
	}

	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			identity, ok := LookupUnit(tt.raw)
			if !ok {
				t.Fatalf("LookupUnit(%q) returned false, expected true", tt.raw)
			}
			if identity.Canonical != tt.canonical {
				t.Errorf("Canonical = %q, want %q", identity.Canonical, tt.canonical)
			}
			if identity.Family != tt.family {
				t.Errorf("Family = %q, want %q", identity.Family, tt.family)
			}
		})
	}
}

func TestLookupUnit_UnknownUnits(t *testing.T) {
	unknowns := []string{"smidgen", "handful", "barrel", "xyz", ""}

	for _, raw := range unknowns {
		t.Run(raw, func(t *testing.T) {
			_, ok := LookupUnit(raw)
			if ok {
				t.Errorf("LookupUnit(%q) returned true, expected false", raw)
			}
		})
	}
}

func TestLookupUnit_CanonicalNormalization(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		canonical string
	}{
		{"g normalizes to gram", "g", "gram"},
		{"tsp normalizes to teaspoon", "tsp", "teaspoon"},
		{"tbsp normalizes to tablespoon", "tbsp", "tablespoon"},
		{"ml normalizes to milliliter", "ml", "milliliter"},
		{"kg normalizes to kilogram", "kg", "kilogram"},
		{"oz normalizes to ounce", "oz", "ounce"},
		{"lb normalizes to pound", "lb", "pound"},
		{"pcs normalizes to piece", "pcs", "piece"},
		{"cloves normalizes to clove", "cloves", "clove"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			identity, ok := LookupUnit(tt.raw)
			if !ok {
				t.Fatalf("LookupUnit(%q) returned false", tt.raw)
			}
			if identity.Canonical != tt.canonical {
				t.Errorf("Canonical = %q, want %q", identity.Canonical, tt.canonical)
			}
		})
	}
}
