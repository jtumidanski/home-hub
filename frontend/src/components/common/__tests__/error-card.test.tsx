import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { ErrorCard } from "../error-card";

describe("ErrorCard", () => {
  it("renders the error message in an alert", () => {
    render(<ErrorCard message="Something went wrong" />);
    expect(screen.getByRole("alert")).toHaveTextContent("Something went wrong");
  });

  it("carries an icon so status is not colour-only", () => {
    render(<ErrorCard message="Error!" />);
    const icon = screen.getByRole("alert").querySelector("svg");
    expect(icon).toHaveAttribute("aria-hidden", "true");
  });

  it("renders with destructive styling", () => {
    render(<ErrorCard message="Error!" />);
    expect(screen.getByRole("alert")).toHaveClass("border-destructive");
  });
});
