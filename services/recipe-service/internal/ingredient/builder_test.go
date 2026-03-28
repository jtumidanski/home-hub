package ingredient

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestBuilder_Build(t *testing.T) {
	t.Run("successful build with valid name", func(t *testing.T) {
		id := uuid.New()
		tenantID := uuid.New()
		now := time.Now().UTC()

		m, err := NewBuilder().
			SetId(id).
			SetTenantID(tenantID).
			SetName("garlic").
			SetDisplayName("Garlic").
			SetUnitFamily("count").
			SetAliasCount(2).
			SetUsageCount(5).
			SetCreatedAt(now).
			SetUpdatedAt(now).
			Build()

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if m.Id() != id {
			t.Errorf("Id() = %v, want %v", m.Id(), id)
		}
		if m.TenantID() != tenantID {
			t.Errorf("TenantID() = %v, want %v", m.TenantID(), tenantID)
		}
		if m.Name() != "garlic" {
			t.Errorf("Name() = %q, want %q", m.Name(), "garlic")
		}
		if m.DisplayName() != "Garlic" {
			t.Errorf("DisplayName() = %q, want %q", m.DisplayName(), "Garlic")
		}
		if m.UnitFamily() != "count" {
			t.Errorf("UnitFamily() = %q, want %q", m.UnitFamily(), "count")
		}
		if m.AliasCount() != 2 {
			t.Errorf("AliasCount() = %d, want %d", m.AliasCount(), 2)
		}
		if m.UsageCount() != 5 {
			t.Errorf("UsageCount() = %d, want %d", m.UsageCount(), 5)
		}
		if !m.CreatedAt().Equal(now) {
			t.Errorf("CreatedAt() = %v, want %v", m.CreatedAt(), now)
		}
		if !m.UpdatedAt().Equal(now) {
			t.Errorf("UpdatedAt() = %v, want %v", m.UpdatedAt(), now)
		}
	})

	t.Run("ErrNameRequired when name is empty", func(t *testing.T) {
		_, err := NewBuilder().
			SetUnitFamily("count").
			Build()

		if err != ErrNameRequired {
			t.Fatalf("expected ErrNameRequired, got %v", err)
		}
	})

	t.Run("ErrInvalidUnitFamily for invalid unit family", func(t *testing.T) {
		_, err := NewBuilder().
			SetName("garlic").
			SetUnitFamily("temperature").
			Build()

		if err != ErrInvalidUnitFamily {
			t.Fatalf("expected ErrInvalidUnitFamily, got %v", err)
		}
	})
}

func TestBuilder_ValidUnitFamilies(t *testing.T) {
	tests := []struct {
		name       string
		unitFamily string
		wantErr    bool
	}{
		{"count is valid", "count", false},
		{"weight is valid", "weight", false},
		{"volume is valid", "volume", false},
		{"empty is valid", "", false},
		{"invalid family", "temperature", true},
		{"invalid family arbitrary", "foo", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewBuilder().
				SetName("test").
				SetUnitFamily(tt.unitFamily).
				Build()

			if tt.wantErr && err != ErrInvalidUnitFamily {
				t.Errorf("expected ErrInvalidUnitFamily, got %v", err)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

func TestValidUnitFamily(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"count", true},
		{"weight", true},
		{"volume", true},
		{"", true},
		{"temperature", false},
		{"foo", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ValidUnitFamily(tt.input); got != tt.want {
				t.Errorf("ValidUnitFamily(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestBuilder_Aliases(t *testing.T) {
	aliases := []Alias{
		{id: uuid.New(), name: "ajo"},
		{id: uuid.New(), name: "knoblauch"},
	}

	m, err := NewBuilder().
		SetName("garlic").
		SetAliases(aliases).
		Build()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(m.Aliases()) != 2 {
		t.Fatalf("expected 2 aliases, got %d", len(m.Aliases()))
	}
	if m.Aliases()[0].Name() != "ajo" {
		t.Errorf("Aliases()[0].Name() = %q, want %q", m.Aliases()[0].Name(), "ajo")
	}
	if m.Aliases()[1].Name() != "knoblauch" {
		t.Errorf("Aliases()[1].Name() = %q, want %q", m.Aliases()[1].Name(), "knoblauch")
	}
}
