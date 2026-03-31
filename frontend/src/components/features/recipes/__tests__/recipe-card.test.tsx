import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { RecipeCard } from "../recipe-card";
import type { RecipeListItem } from "@/types/models/recipe";

const mockNavigate = vi.fn();
vi.mock("react-router-dom", () => ({
  useNavigate: () => mockNavigate,
}));

function makeRecipe(overrides: Partial<RecipeListItem["attributes"]> = {}): RecipeListItem {
  return {
    id: "recipe-1",
    type: "recipes",
    attributes: {
      title: "Pasta Carbonara",
      tags: ["italian", "pasta"],
      plannerReady: false,
      resolvedIngredients: 0,
      totalIngredients: 0,
      createdAt: "2026-03-01T00:00:00Z",
      updatedAt: "2026-03-01T00:00:00Z",
      ...overrides,
    },
  };
}

describe("RecipeCard", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders recipe title and tags", () => {
    render(<RecipeCard recipe={makeRecipe()} onDelete={vi.fn()} />);

    expect(screen.getByText("Pasta Carbonara")).toBeInTheDocument();
    expect(screen.getByText("Italian")).toBeInTheDocument();
    expect(screen.getByText("Pasta")).toBeInTheDocument();
  });

  it("renders description when provided", () => {
    render(<RecipeCard recipe={makeRecipe({ description: "Classic Roman dish" })} onDelete={vi.fn()} />);

    expect(screen.getByText("Classic Roman dish")).toBeInTheDocument();
  });

  it("renders total time when prep and cook time are provided", () => {
    render(
      <RecipeCard
        recipe={makeRecipe({ prepTimeMinutes: 10, cookTimeMinutes: 20 })}
        onDelete={vi.fn()}
      />,
    );

    expect(screen.getByText("30 min")).toBeInTheDocument();
  });

  it("renders ??? min when time is not provided", () => {
    render(<RecipeCard recipe={makeRecipe()} onDelete={vi.fn()} />);

    expect(screen.getByText("??? min")).toBeInTheDocument();
  });

  it("shows classification as a tag and deduplicates it from cooklang tags", () => {
    render(<RecipeCard recipe={makeRecipe({ tags: ["italian", "dinner", "pasta"], classification: "dinner" })} onDelete={vi.fn()} />);

    expect(screen.getByText("Italian")).toBeInTheDocument();
    expect(screen.getByText("Pasta")).toBeInTheDocument();
    // "dinner" appears once (from classification), not twice
    expect(screen.getAllByText("Dinner")).toHaveLength(1);
  });

  it("navigates to detail page on card click", async () => {
    const user = userEvent.setup();
    render(<RecipeCard recipe={makeRecipe()} onDelete={vi.fn()} />);

    await user.click(screen.getByText("Pasta Carbonara"));
    expect(mockNavigate).toHaveBeenCalledWith("/app/recipes/recipe-1");
  });

  it("calls onDelete when Delete action is clicked", async () => {
    const user = userEvent.setup();
    const onDelete = vi.fn();
    render(<RecipeCard recipe={makeRecipe()} onDelete={onDelete} />);

    await user.click(screen.getByRole("button", { name: "Actions" }));
    await user.click(screen.getByText("Delete"));

    expect(onDelete).toHaveBeenCalledWith("recipe-1");
  });

  it("navigates to edit page when Edit action is clicked", async () => {
    const user = userEvent.setup();
    render(<RecipeCard recipe={makeRecipe()} onDelete={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: "Actions" }));
    await user.click(screen.getByText("Edit"));

    expect(mockNavigate).toHaveBeenCalledWith("/app/recipes/recipe-1/edit");
  });
});
