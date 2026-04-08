package ingredient

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus/hooks/test"
)

func TestProcessorGetByIDs(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	proc := NewProcessor(l, context.Background(), db)
	tenantID := uuid.New()

	a, err := proc.Create(tenantID, "flour", "Flour", "weight", nil)
	if err != nil {
		t.Fatalf("create a: %v", err)
	}
	b, err := proc.Create(tenantID, "sugar", "Sugar", "weight", nil)
	if err != nil {
		t.Fatalf("create b: %v", err)
	}
	if _, err := proc.AddAlias(tenantID, b.Id(), "white sugar"); err != nil {
		t.Fatalf("add alias: %v", err)
	}

	t.Run("empty", func(t *testing.T) {
		m, err := proc.GetByIDs(nil)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if len(m) != 0 {
			t.Errorf("got %d, want 0", len(m))
		}
	})

	t.Run("single", func(t *testing.T) {
		m, err := proc.GetByIDs([]uuid.UUID{a.Id()})
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if got := m[a.Id()].Name(); got != "flour" {
			t.Errorf("name = %q, want flour", got)
		}
	})

	t.Run("multiple with aliases preloaded", func(t *testing.T) {
		missing := uuid.New()
		m, err := proc.GetByIDs([]uuid.UUID{a.Id(), b.Id(), missing})
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if len(m) != 2 {
			t.Fatalf("got %d, want 2", len(m))
		}
		bm := m[b.Id()]
		if len(bm.Aliases()) != 1 {
			t.Errorf("b aliases = %d, want 1", len(bm.Aliases()))
		}
	})
}
