import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ViewModeToggle } from "../view-mode-toggle";

describe("ViewModeToggle", () => {
  it("renders Week and Month options", () => {
    render(<ViewModeToggle mode="week" onChange={vi.fn()} />);
    expect(screen.getByRole("button", { name: "Week" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Month" })).toBeInTheDocument();
  });

  it("marks the active mode with aria-pressed", () => {
    render(<ViewModeToggle mode="month" onChange={vi.fn()} />);
    expect(screen.getByRole("button", { name: "Month" })).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByRole("button", { name: "Week" })).toHaveAttribute("aria-pressed", "false");
  });

  it("calls onChange with the clicked mode", async () => {
    const onChange = vi.fn();
    render(<ViewModeToggle mode="week" onChange={onChange} />);
    await userEvent.click(screen.getByRole("button", { name: "Month" }));
    expect(onChange).toHaveBeenCalledWith("month");
  });
});
