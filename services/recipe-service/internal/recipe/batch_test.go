package recipe

import (
	"context"
	"testing"

	"github.com/google/uuid"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus/hooks/test"
)

func TestProcessorGetByIDs(t *testing.T) {
	db := setupTestDB(t)
	tenantID := uuid.New()
	householdID := uuid.New()
	userID := uuid.New()
	ctx := tenantctx.WithContext(context.Background(), tenantctx.New(tenantID, householdID, userID))
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, ctx, db)

	a, _, err := p.Create(tenantID, householdID, CreateAttrs{Title: "A", Source: "Boil A."})
	if err != nil {
		t.Fatalf("create a: %v", err)
	}
	b, _, err := p.Create(tenantID, householdID, CreateAttrs{Title: "B", Source: "Boil B."})
	if err != nil {
		t.Fatalf("create b: %v", err)
	}

	t.Run("empty", func(t *testing.T) {
		m, err := p.GetByIDs(nil)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if len(m) != 0 {
			t.Errorf("got %d, want 0", len(m))
		}
	})

	t.Run("single", func(t *testing.T) {
		m, err := p.GetByIDs([]uuid.UUID{a.Id()})
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if got := m[a.Id()].Title(); got != "A" {
			t.Errorf("title = %q, want A", got)
		}
	})

	t.Run("multiple skips deleted and missing", func(t *testing.T) {
		if err := p.Delete(b.Id()); err != nil {
			t.Fatalf("delete b: %v", err)
		}
		missing := uuid.New()
		m, err := p.GetByIDs([]uuid.UUID{a.Id(), b.Id(), missing})
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if len(m) != 1 {
			t.Errorf("got %d, want 1", len(m))
		}
		if _, ok := m[b.Id()]; ok {
			t.Errorf("deleted recipe should not be in result")
		}
	})
}
