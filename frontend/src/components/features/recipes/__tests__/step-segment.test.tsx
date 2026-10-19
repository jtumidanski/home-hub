import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { StepSegment } from "../step-segment";
import type { Segment } from "@/types/models/recipe";

describe("StepSegment", () => {
  it.each([
    [{ type: "ingredient", name: "flour" }, "Ingredient:", "bg-swatch-amber-soft"],
    [{ type: "cookware", name: "pot" }, "Cookware:", "bg-swatch-blue-soft"],
    [{ type: "timer", quantity: "10", unit: "min" }, "Timer:", "bg-swatch-green-soft"],
    [{ type: "reference", name: "base sauce" }, "Reference:", "bg-swatch-violet-soft"],
  ] as [Segment, string, string][])("labels %j with text and a soft swatch", (segment, label, swatch) => {
    render(<StepSegment segment={segment} />);

    const cue = screen.getByText(label);
    expect(cue.parentElement).toHaveClass(swatch);
  });
});
