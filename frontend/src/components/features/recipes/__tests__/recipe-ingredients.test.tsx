import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { RecipeIngredients } from "../recipe-ingredients";
import type { Ingredient } from "@/types/models/recipe";

describe("RecipeIngredients", () => {
  it("renders ingredient list with quantity, unit, and name", () => {
    const ingredients: Ingredient[] = [
      { name: "eggs", quantity: "3", unit: "" },
      { name: "flour", quantity: "125", unit: "g" },
      { name: "salt", quantity: "1", unit: "tsp" },
    ];
    render(<RecipeIngredients ingredients={ingredients} />);

    expect(screen.getByText("eggs")).toBeInTheDocument();
    expect(screen.getByText("3")).toBeInTheDocument();
    expect(screen.getByText("125 g")).toBeInTheDocument();
    expect(screen.getByText("1 tsp")).toBeInTheDocument();
    expect(screen.getByText("salt")).toBeInTheDocument();
  });

  it("renders empty state when no ingredients", () => {
    render(<RecipeIngredients ingredients={[]} />);

    expect(screen.getByText("No ingredients found.")).toBeInTheDocument();
  });

  it("renders as a bulleted list", () => {
    const ingredients: Ingredient[] = [
      { name: "salt", quantity: "1", unit: "tsp" },
    ];
    const { container } = render(<RecipeIngredients ingredients={ingredients} />);

    expect(container.querySelector("ul")).toBeInTheDocument();
    expect(container.querySelectorAll("li")).toHaveLength(1);
  });

  it("handles ingredient without quantity", () => {
    const ingredients: Ingredient[] = [
      { name: "salt", quantity: "", unit: "" },
    ];
    render(<RecipeIngredients ingredients={ingredients} />);

    expect(screen.getByText("salt")).toBeInTheDocument();
  });
});
