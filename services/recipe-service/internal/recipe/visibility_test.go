package recipe

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus/hooks/test"

	"github.com/jtumidanski/home-hub/services/recipe-service/internal/recipeimage"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
)

func TestImageVisibility(t *testing.T) {
	db := setupTestDB(t)
	tid, hidA, hidB := uuid.New(), uuid.New(), uuid.New()
	ctxA := tenantctx.WithContext(context.Background(), tenantctx.New(tid, hidA, uuid.New()))
	ctxB := tenantctx.WithContext(context.Background(), tenantctx.New(tid, hidB, uuid.New()))
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, ctxA, db)
	m, _, err := p.Create(tid, hidA, CreateAttrs{Title: "Soup", Source: "Boil water."})
	if err != nil {
		t.Fatal(err)
	}
	visible := ImageVisibility(db)

	if err := visible(ctxA, m.Id()); err != nil {
		t.Fatalf("own household: %v", err)
	}
	if err := visible(ctxB, m.Id()); !errors.Is(err, recipeimage.ErrRecipeNotFound) {
		t.Fatalf("other household err = %v", err)
	}
	if err := visible(ctxA, uuid.New()); !errors.Is(err, recipeimage.ErrRecipeNotFound) {
		t.Fatalf("missing err = %v", err)
	}
	if err := visible(context.Background(), m.Id()); !errors.Is(err, recipeimage.ErrRecipeNotFound) {
		t.Fatalf("no tenant err = %v", err)
	}
	if err := p.Delete(m.Id()); err != nil {
		t.Fatal(err)
	}
	if err := visible(ctxA, m.Id()); !errors.Is(err, recipeimage.ErrRecipeNotFound) {
		t.Fatalf("soft-deleted err = %v", err)
	}
}
