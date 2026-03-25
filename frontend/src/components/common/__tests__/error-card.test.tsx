import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { ErrorCard } from "../error-card";

describe("ErrorCard", () => {
  it("renders the error message", () => {
    render(<ErrorCard message="Something went wrong" />);
    expect(screen.getByText("Something went wrong")).toBeInTheDocument();
  });

  it("renders with destructive styling", () => {
    const { container } = render(<ErrorCard message="Error!" />);
    expect(container.querySelector(".border-destructive")).toBeInTheDocument();
    expect(container.querySelector(".text-destructive")).toBeInTheDocument();
  });
});
