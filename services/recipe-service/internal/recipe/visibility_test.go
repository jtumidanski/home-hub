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

	tests := []struct {
		name   string
		before func(t *testing.T)
		ctx    context.Context
		id     uuid.UUID
		want   error
	}{
		{name: "own household", ctx: ctxA, id: m.Id()},
		{name: "other household", ctx: ctxB, id: m.Id(), want: recipeimage.ErrRecipeNotFound},
		{name: "missing recipe", ctx: ctxA, id: uuid.New(), want: recipeimage.ErrRecipeNotFound},
		{name: "no tenant", ctx: context.Background(), id: m.Id(), want: recipeimage.ErrRecipeNotFound},
		{
			name: "soft-deleted",
			before: func(t *testing.T) {
				if err := p.Delete(m.Id()); err != nil {
					t.Fatal(err)
				}
			},
			ctx: ctxA, id: m.Id(), want: recipeimage.ErrRecipeNotFound,
		},
	}
	// Cases run in order: the soft-delete case mutates the shared recipe.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.before != nil {
				tc.before(t)
			}
			err := visible(tc.ctx, tc.id)
			if tc.want == nil {
				if err != nil {
					t.Fatalf("err = %v", err)
				}
				return
			}
			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
		})
	}
}
