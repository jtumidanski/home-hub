import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { LocationProbe } from "@/test/location-probe";

vi.mock("@/lib/hooks/api/use-ingredients", () => ({
  useIngredients: () => ({ data: { data: [], meta: { total: 0 } }, isLoading: false, refetch: vi.fn() }),
  useCreateIngredient: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useDeleteIngredient: () => ({ mutateAsync: vi.fn() }),
}));
vi.mock("@/lib/hooks/api/use-ingredient-categories", () => ({
  useIngredientCategories: () => ({ data: { data: [] } }),
}));
vi.mock("@/components/features/ingredients/category-manager", () => ({
  CategoryManager: () => <div>category-manager-stub</div>,
}));
vi.mock("@/components/features/ingredients/bulk-categorize", () => ({
  BulkCategorize: () => <div>bulk-categorize-stub</div>,
}));

import { IngredientsPage } from "../IngredientsPage";

function renderAt(url: string) {
  return render(
    <MemoryRouter initialEntries={[url]}>
      <Routes>
        <Route path="/app/ingredients/list" element={<IngredientsPage view="list" />} />
        <Route path="/app/ingredients/categories" element={<IngredientsPage view="categories" />} />
        <Route path="/app/ingredients/bulk" element={<IngredientsPage view="bulk" />} />
      </Routes>
      <LocationProbe />
    </MemoryRouter>,
  );
}

const loc = () => screen.getByTestId("location").textContent;

describe("IngredientsPage — views from URL", () => {
  it("pushes categories (keeping the query) and toggles back to list", async () => {
    renderAt("/app/ingredients/list?category=abc");
    await userEvent.click(screen.getByRole("button", { name: /Categories/ }));
    expect(loc()).toBe("/app/ingredients/categories?category=abc");
    expect(screen.getByTestId("nav-type").textContent).toBe("PUSH");
    expect(screen.getByText("category-manager-stub")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: /Categories/ }));
    expect(loc()).toBe("/app/ingredients/list?category=abc");
  });

  it("goes from categories to bulk", async () => {
    renderAt("/app/ingredients/categories");
    await userEvent.click(screen.getByRole("button", { name: /Bulk Edit/ }));
    expect(loc()).toBe("/app/ingredients/bulk");
    expect(screen.getByText("bulk-categorize-stub")).toBeInTheDocument();
  });

  it("New from categories navigates to list and opens the create form", async () => {
    renderAt("/app/ingredients/categories");
    await userEvent.click(screen.getByRole("button", { name: /New/ }));
    expect(loc()).toBe("/app/ingredients/list");
    expect(screen.getByPlaceholderText("Ingredient name...")).toBeInTheDocument();
  });
});
