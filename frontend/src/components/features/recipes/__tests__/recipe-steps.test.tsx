import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { RecipeSteps } from "../recipe-steps";
import type { Step } from "@/types/models/recipe";

describe("RecipeSteps", () => {
  it("renders step numbers and text", () => {
    const steps: Step[] = [
      { number: 1, segments: [{ type: "text", value: "Boil water." }] },
      { number: 2, segments: [{ type: "text", value: "Add pasta." }] },
    ];
    render(<RecipeSteps steps={steps} />);

    expect(screen.getByText("1")).toBeInTheDocument();
    expect(screen.getByText("2")).toBeInTheDocument();
    expect(screen.getByText("Boil water.")).toBeInTheDocument();
    expect(screen.getByText("Add pasta.")).toBeInTheDocument();
  });

  it("renders empty state when no steps", () => {
    render(<RecipeSteps steps={[]} />);

    expect(screen.getByText("No steps found.")).toBeInTheDocument();
  });

  it("renders ingredient segments with name", () => {
    const steps: Step[] = [
      {
        number: 1,
        segments: [
          { type: "text", value: "Add " },
          { type: "ingredient", name: "salt", quantity: "1", unit: "tsp" },
          { type: "text", value: "." },
        ],
      },
    ];
    render(<RecipeSteps steps={steps} />);

    expect(screen.getByText("salt")).toBeInTheDocument();
    expect(screen.getByText("(1 tsp)")).toBeInTheDocument();
  });

  it("renders cookware segments", () => {
    const steps: Step[] = [
      {
        number: 1,
        segments: [
          { type: "text", value: "Place in a " },
          { type: "cookware", name: "large pot" },
        ],
      },
    ];
    render(<RecipeSteps steps={steps} />);

    expect(screen.getByText("large pot")).toBeInTheDocument();
  });

  it("renders timer segments", () => {
    const steps: Step[] = [
      {
        number: 1,
        segments: [
          { type: "text", value: "Cook for " },
          { type: "timer", quantity: "8", unit: "minutes" },
        ],
      },
    ];
    render(<RecipeSteps steps={steps} />);

    expect(screen.getByText("8 minutes")).toBeInTheDocument();
  });

  it("renders recipe reference segments", () => {
    const steps: Step[] = [
      {
        number: 1,
        segments: [
          { type: "text", value: "Serve with " },
          { type: "reference", name: "Salsa Verde", path: "./Sauces/Salsa Verde" },
        ],
      },
    ];
    render(<RecipeSteps steps={steps} />);

    expect(screen.getByText("Salsa Verde")).toBeInTheDocument();
  });

  it("renders section headers", () => {
    const steps: Step[] = [
      { number: 1, section: "Filling", segments: [{ type: "text", value: "Cook rice." }] },
      { number: 2, section: "Filling", segments: [{ type: "text", value: "Add beans." }] },
      { number: 3, section: "Assembly", segments: [{ type: "text", value: "Stuff peppers." }] },
    ];
    render(<RecipeSteps steps={steps} />);

    expect(screen.getByText("Filling")).toBeInTheDocument();
    expect(screen.getByText("Assembly")).toBeInTheDocument();
    // Section header should only render once per section
    expect(screen.getAllByText("Filling")).toHaveLength(1);
  });
});
