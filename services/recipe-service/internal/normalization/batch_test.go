package normalization

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus/hooks/test"
)

func TestGetByRecipeIDs(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	proc := NewProcessor(l, context.Background(), db)

	tenantID := uuid.New()
	householdID := uuid.New()
	r1, r2, r3 := uuid.New(), uuid.New(), uuid.New()

	if _, err := proc.NormalizeIngredients(tenantID, householdID, r1, []ParsedIngredient{
		{Name: "carrots", Quantity: "2", Unit: ""},
		{Name: "onions", Quantity: "1", Unit: ""},
	}); err != nil {
		t.Fatalf("seed r1: %v", err)
	}
	if _, err := proc.NormalizeIngredients(tenantID, householdID, r2, []ParsedIngredient{
		{Name: "garlic", Quantity: "3", Unit: "clove"},
	}); err != nil {
		t.Fatalf("seed r2: %v", err)
	}

	t.Run("empty", func(t *testing.T) {
		m, err := proc.GetByRecipeIDs(nil)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if len(m) != 0 {
			t.Errorf("got %d, want 0", len(m))
		}
	})

	t.Run("single", func(t *testing.T) {
		m, err := proc.GetByRecipeIDs([]uuid.UUID{r2})
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if len(m[r2]) != 1 {
			t.Errorf("r2 got %d ingredients, want 1", len(m[r2]))
		}
	})

	t.Run("multiple preserves position order", func(t *testing.T) {
		m, err := proc.GetByRecipeIDs([]uuid.UUID{r1, r2, r3})
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if len(m[r1]) != 2 {
			t.Fatalf("r1 got %d, want 2", len(m[r1]))
		}
		if m[r1][0].Position() != 0 || m[r1][1].Position() != 1 {
			t.Errorf("r1 position order = [%d, %d], want [0, 1]", m[r1][0].Position(), m[r1][1].Position())
		}
		if _, ok := m[r3]; ok {
			t.Errorf("r3 should not be in result")
		}
	})
}
