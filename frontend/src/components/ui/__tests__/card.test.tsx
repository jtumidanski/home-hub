import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { Card, CardTitle } from "../card";

describe("Card", () => {
  it("root uses rounded-xl and shadow-card", () => {
    render(<Card data-testid="card">x</Card>);
    const cls = screen.getByTestId("card").className;
    expect(cls).toContain("rounded-xl");
    expect(cls).toContain("shadow-card");
    expect(cls).toContain("dark:border");
  });

  it("CardTitle uses the heading font at no smaller than text-lg", () => {
    render(<CardTitle>Title</CardTitle>);
    const cls = screen.getByText("Title").className;
    expect(cls).toContain("font-heading");
    expect(cls).toContain("text-lg");
    expect(cls).not.toContain("text-sm");
  });
});
