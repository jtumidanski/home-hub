import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { BookOpen } from "lucide-react";
import { EmptyState } from "../empty-state";

describe("EmptyState", () => {
  it("renders the title as an h2 in the heading face", () => {
    render(<EmptyState title="No recipes yet" />);
    const h2 = screen.getByRole("heading", { level: 2, name: "No recipes yet" });
    expect(h2).toHaveClass("font-heading");
    expect(h2).toHaveClass("text-lg");
  });

  it("renders the title at the requested heading level", () => {
    render(<EmptyState title="Oops" headingLevel="h1" />);
    expect(screen.getByRole("heading", { level: 1, name: "Oops" })).toBeInTheDocument();
  });

  it("hides the icon from assistive tech", () => {
    const { container } = render(<EmptyState icon={BookOpen} title="No recipes yet" />);
    expect(container.querySelector("svg")).toHaveAttribute("aria-hidden", "true");
  });

  it("renders description and action", () => {
    render(
      <EmptyState
        title="No recipes yet"
        description="Add your first one."
        action={<button type="button">Add recipe</button>}
      />,
    );
    expect(screen.getByText("Add your first one.")).toHaveClass("text-muted-foreground");
    expect(screen.getByRole("button", { name: "Add recipe" })).toBeInTheDocument();
  });
});
