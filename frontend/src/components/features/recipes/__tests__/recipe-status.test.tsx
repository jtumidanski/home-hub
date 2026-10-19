import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { RecipeCard } from "../recipe-card";
import { RecipeIngredients } from "../recipe-ingredients";
import type { RecipeListItem, RecipeIngredient } from "@/types/models/recipe";

vi.mock("@/lib/hooks/use-recipe-image-src", () => ({
  useRecipeImageSrc: () => ({ src: null, isError: false }),
}));

vi.stubGlobal(
  "IntersectionObserver",
  class {
    observe() {}
    disconnect() {}
  },
);

function makeRecipe(overrides: Partial<RecipeListItem["attributes"]> = {}): RecipeListItem {
  return {
    id: "recipe-1",
    type: "recipes",
    attributes: {
      title: "Pasta",
      tags: [],
      plannerReady: true,
      resolvedIngredients: 2,
      totalIngredients: 2,
      createdAt: "2026-03-01T00:00:00Z",
      updatedAt: "2026-03-01T00:00:00Z",
      image: null,
      ...overrides,
    },
  };
}

function renderCard(recipe: RecipeListItem) {
  return render(
    <MemoryRouter>
      <RecipeCard recipe={recipe} onDelete={vi.fn()} />
    </MemoryRouter>,
  );
}

describe("recipe status indicators", () => {
  it("resolved recipe card icon uses text-success and has adjacent text", () => {
    const { container } = renderCard(makeRecipe());
    const icon = container.querySelector("svg.text-success");
    expect(icon).not.toBeNull();
    expect(icon?.nextElementSibling).toHaveTextContent("Planner Ready");
  });

  it("unresolved recipe card icon uses text-warning-text and has adjacent text", () => {
    const { container } = renderCard(makeRecipe({ plannerReady: false, resolvedIngredients: 1 }));
    const icon = container.querySelector("svg.text-warning-text");
    expect(icon).not.toBeNull();
    expect(icon?.nextElementSibling).toHaveTextContent("has unresolved ingredients");
  });

  it("ingredient status icons are named and use status colours", () => {
    const base = { rawQuantity: "1", rawUnit: "", canonicalIngredientId: null };
    const ingredients: RecipeIngredient[] = [
      { ...base, id: "i1", rawName: "eggs", position: 0, normalizationStatus: "matched" },
      { ...base, id: "i2", rawName: "mystery", position: 1, normalizationStatus: "unresolved" },
    ];
    render(<RecipeIngredients ingredients={ingredients} recipeId="r1" readOnly />);
    expect(screen.getByRole("img", { name: "Resolved" })).toHaveClass("text-success");
    expect(screen.getByRole("img", { name: "Unresolved" })).toHaveClass("text-warning-text");
  });
});
