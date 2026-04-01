package planner

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func intPtr(v int) *int { return &v }

func TestBuilder_Build(t *testing.T) {
	t.Run("successful build with all fields", func(t *testing.T) {
		id := uuid.New()
		recipeID := uuid.New()
		now := time.Now().UTC()

		m, err := NewBuilder().
			SetId(id).
			SetRecipeID(recipeID).
			SetClassification("dinner").
			SetServingsYield(intPtr(4)).
			SetEatWithinDays(intPtr(3)).
			SetMinGapDays(intPtr(7)).
			SetMaxConsecutiveDays(intPtr(2)).
			SetCreatedAt(now).
			SetUpdatedAt(now).
			Build()

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if m.Id() != id {
			t.Errorf("Id() = %v, want %v", m.Id(), id)
		}
		if m.RecipeID() != recipeID {
			t.Errorf("RecipeID() = %v, want %v", m.RecipeID(), recipeID)
		}
		if m.Classification() != "dinner" {
			t.Errorf("Classification() = %q, want %q", m.Classification(), "dinner")
		}
		if m.ServingsYield() == nil || *m.ServingsYield() != 4 {
			t.Errorf("ServingsYield() = %v, want 4", m.ServingsYield())
		}
		if m.EatWithinDays() == nil || *m.EatWithinDays() != 3 {
			t.Errorf("EatWithinDays() = %v, want 3", m.EatWithinDays())
		}
		if m.MinGapDays() == nil || *m.MinGapDays() != 7 {
			t.Errorf("MinGapDays() = %v, want 7", m.MinGapDays())
		}
		if m.MaxConsecutiveDays() == nil || *m.MaxConsecutiveDays() != 2 {
			t.Errorf("MaxConsecutiveDays() = %v, want 2", m.MaxConsecutiveDays())
		}
		if !m.CreatedAt().Equal(now) {
			t.Errorf("CreatedAt() = %v, want %v", m.CreatedAt(), now)
		}
		if !m.UpdatedAt().Equal(now) {
			t.Errorf("UpdatedAt() = %v, want %v", m.UpdatedAt(), now)
		}
	})

	t.Run("build requires recipe ID", func(t *testing.T) {
		_, err := NewBuilder().Build()
		if err == nil {
			t.Fatal("expected error from empty builder, got nil")
		}
		if err != ErrRecipeIDRequired {
			t.Fatalf("expected ErrRecipeIDRequired, got %v", err)
		}
	})

	t.Run("nil optional fields", func(t *testing.T) {
		m, err := NewBuilder().
			SetRecipeID(uuid.New()).
			SetClassification("lunch").
			Build()

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if m.ServingsYield() != nil {
			t.Errorf("ServingsYield() = %v, want nil", m.ServingsYield())
		}
		if m.EatWithinDays() != nil {
			t.Errorf("EatWithinDays() = %v, want nil", m.EatWithinDays())
		}
		if m.MinGapDays() != nil {
			t.Errorf("MinGapDays() = %v, want nil", m.MinGapDays())
		}
		if m.MaxConsecutiveDays() != nil {
			t.Errorf("MaxConsecutiveDays() = %v, want nil", m.MaxConsecutiveDays())
		}
	})
}
