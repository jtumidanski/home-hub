import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";

vi.mock("react-router-dom", async () => {
  const actual = await vi.importActual<typeof import("react-router-dom")>("react-router-dom");
  return { ...actual, useParams: () => ({ id: "ing-1" }), useNavigate: () => vi.fn() };
});

vi.mock("@/lib/hooks/api/use-ingredients", () => ({
  useIngredient: () => ({
    isLoading: false,
    data: {
      data: {
        id: "ing-1",
        attributes: {
          name: "flour",
          displayName: "Flour",
          unitFamily: "weight",
          categoryId: null,
          aliases: [{ id: "a1", name: "plain flour" }],
        },
      },
    },
  }),
  useIngredientRecipes: () => ({
    data: { data: [{ recipeId: "r1", recipeName: "Pancakes" }] },
  }),
  useIngredients: () => ({ data: { data: [] } }),
  useUpdateIngredient: () => ({ mutateAsync: vi.fn() }),
  useDeleteIngredient: () => ({ mutateAsync: vi.fn() }),
  useReassignIngredient: () => ({ mutateAsync: vi.fn() }),
  useAddAlias: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useRemoveAlias: () => ({ mutateAsync: vi.fn() }),
}));

vi.mock("@/lib/hooks/api/use-ingredient-categories", () => ({
  useIngredientCategories: () => ({ data: { data: [] } }),
}));

import { IngredientDetailPage } from "../IngredientDetailPage";

describe("IngredientDetailPage — pointer targets", () => {
  it("gives the alias-remove button a 24px target and 44px on coarse pointers", () => {
    render(
      <MemoryRouter>
        <IngredientDetailPage />
      </MemoryRouter>,
    );
    const remove = screen.getByRole("button", { name: "Remove alias plain flour" });
    expect(remove).toHaveClass("size-6", "coarse:size-11");
  });

  it("gives the recipe-reference button a coarse touch floor", () => {
    render(
      <MemoryRouter>
        <IngredientDetailPage />
      </MemoryRouter>,
    );
    const ref = screen.getByRole("button", { name: "Pancakes" });
    expect(ref).toHaveClass("min-h-6", "coarse:min-h-11");
  });
});
