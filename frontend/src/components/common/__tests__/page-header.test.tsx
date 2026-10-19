import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { PageHeader } from "../page-header";

describe("PageHeader", () => {
  it("renders the title as a Fraunces h1 that can wrap", () => {
    render(<PageHeader title="Recipes" />);
    const h1 = screen.getByRole("heading", { level: 1, name: "Recipes" });
    expect(h1).toHaveClass("font-heading");
    expect(h1).toHaveClass("min-w-0");
    expect(h1).toHaveClass("break-words");
  });

  it("wraps the title in a header element", () => {
    render(<PageHeader title="Recipes" />);
    expect(screen.getByRole("heading", { level: 1 }).closest("header")).toBeInTheDocument();
  });

  it("renders description in muted text", () => {
    render(<PageHeader title="Recipes" description="Your collection" />);
    expect(screen.getByText("Your collection")).toHaveClass("text-muted-foreground");
  });

  it("renders actions in a wrapping container", () => {
    render(<PageHeader title="Recipes" actions={<button type="button">Add</button>} />);
    const button = screen.getByRole("button", { name: "Add" });
    expect(button.parentElement).toHaveClass("flex-wrap");
  });

  it("omits description and actions when not given", () => {
    render(<PageHeader title="Recipes" />);
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
  });
});
