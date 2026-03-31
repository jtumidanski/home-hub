package category

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestBuilder_Build(t *testing.T) {
	t.Run("successful build with all fields", func(t *testing.T) {
		id := uuid.New()
		tenantID := uuid.New()
		now := time.Now().UTC()

		m, err := NewBuilder().
			SetId(id).
			SetTenantID(tenantID).
			SetName("Produce").
			SetSortOrder(1).
			SetIngredientCount(5).
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
		if m.Name() != "Produce" {
			t.Errorf("Name() = %q, want %q", m.Name(), "Produce")
		}
		if m.SortOrder() != 1 {
			t.Errorf("SortOrder() = %d, want %d", m.SortOrder(), 1)
		}
		if m.IngredientCount() != 5 {
			t.Errorf("IngredientCount() = %d, want %d", m.IngredientCount(), 5)
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
			SetSortOrder(1).
			Build()

		if err != ErrNameRequired {
			t.Fatalf("expected ErrNameRequired, got %v", err)
		}
	})

	t.Run("ErrNameTooLong when name exceeds 100 chars", func(t *testing.T) {
		longName := strings.Repeat("a", 101)
		_, err := NewBuilder().
			SetName(longName).
			Build()

		if err != ErrNameTooLong {
			t.Fatalf("expected ErrNameTooLong, got %v", err)
		}
	})

	t.Run("name exactly 100 chars is valid", func(t *testing.T) {
		name := strings.Repeat("a", 100)
		_, err := NewBuilder().
			SetName(name).
			Build()

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("ErrInvalidSortOrder when sort order is negative", func(t *testing.T) {
		_, err := NewBuilder().
			SetName("Produce").
			SetSortOrder(-1).
			Build()

		if err != ErrInvalidSortOrder {
			t.Fatalf("expected ErrInvalidSortOrder, got %v", err)
		}
	})

	t.Run("sort order zero is valid", func(t *testing.T) {
		_, err := NewBuilder().
			SetName("Produce").
			SetSortOrder(0).
			Build()

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

func TestBuilder_Validation(t *testing.T) {
	tests := []struct {
		name      string
		catName   string
		sortOrder int
		wantErr   error
	}{
		{"valid category", "Produce", 1, nil},
		{"empty name", "", 0, ErrNameRequired},
		{"name too long", strings.Repeat("x", 101), 0, ErrNameTooLong},
		{"negative sort order", "Dairy", -5, ErrInvalidSortOrder},
		{"zero sort order is valid", "Other", 0, nil},
		{"large sort order is valid", "Custom", 9999, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewBuilder().
				SetName(tt.catName).
				SetSortOrder(tt.sortOrder).
				Build()

			if err != tt.wantErr {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestModel_WithIngredientCount(t *testing.T) {
	m, err := NewBuilder().
		SetName("Produce").
		SetIngredientCount(0).
		Build()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	updated := m.WithIngredientCount(10)

	if updated.IngredientCount() != 10 {
		t.Errorf("WithIngredientCount(10).IngredientCount() = %d, want 10", updated.IngredientCount())
	}
	// Original is unchanged (value receiver copy)
	if m.IngredientCount() != 0 {
		t.Errorf("original IngredientCount() = %d, want 0", m.IngredientCount())
	}
}

func TestModel_Immutability(t *testing.T) {
	id := uuid.New()
	tenantID := uuid.New()
	now := time.Now().UTC()

	m, err := NewBuilder().
		SetId(id).
		SetTenantID(tenantID).
		SetName("Dairy & Eggs").
		SetSortOrder(3).
		SetCreatedAt(now).
		SetUpdatedAt(now).
		Build()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// WithIngredientCount returns a copy, not mutating original
	m2 := m.WithIngredientCount(42)
	if m.IngredientCount() != 0 {
		t.Error("original model was mutated by WithIngredientCount")
	}
	if m2.IngredientCount() != 42 {
		t.Errorf("copy IngredientCount() = %d, want 42", m2.IngredientCount())
	}
	// Other fields preserved on copy
	if m2.Name() != "Dairy & Eggs" {
		t.Errorf("copy Name() = %q, want %q", m2.Name(), "Dairy & Eggs")
	}
}
