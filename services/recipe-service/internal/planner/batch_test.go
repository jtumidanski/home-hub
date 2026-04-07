package planner

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

	r1, r2, r3 := uuid.New(), uuid.New(), uuid.New()
	if _, err := proc.CreateOrUpdate(r1, ConfigAttrs{Classification: strPtr("dinner"), ServingsYield: ptrInt(4)}); err != nil {
		t.Fatalf("seed r1: %v", err)
	}
	if _, err := proc.CreateOrUpdate(r2, ConfigAttrs{Classification: strPtr("lunch"), ServingsYield: ptrInt(2)}); err != nil {
		t.Fatalf("seed r2: %v", err)
	}

	t.Run("empty input", func(t *testing.T) {
		m, err := proc.GetByRecipeIDs(nil)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if len(m) != 0 {
			t.Errorf("got %d entries, want 0", len(m))
		}
	})

	t.Run("single id", func(t *testing.T) {
		m, err := proc.GetByRecipeIDs([]uuid.UUID{r1})
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if len(m) != 1 {
			t.Fatalf("got %d entries, want 1", len(m))
		}
		if got := m[r1].Classification(); got != "dinner" {
			t.Errorf("classification = %q, want dinner", got)
		}
	})

	t.Run("multiple ids with one missing", func(t *testing.T) {
		m, err := proc.GetByRecipeIDs([]uuid.UUID{r1, r2, r3})
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if len(m) != 2 {
			t.Fatalf("got %d entries, want 2", len(m))
		}
		if _, ok := m[r3]; ok {
			t.Errorf("r3 should not be in result")
		}
		if m[r2].Classification() != "lunch" {
			t.Errorf("r2 classification = %q, want lunch", m[r2].Classification())
		}
	})
}
