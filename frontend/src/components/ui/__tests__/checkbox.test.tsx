import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { Checkbox } from "../checkbox";

describe("Checkbox", () => {
  it("exposes the checkbox role with an accessible name and toggles", async () => {
    const onCheckedChange = vi.fn();
    render(<Checkbox aria-label="Select all" onCheckedChange={onCheckedChange} />);
    const box = screen.getByRole("checkbox", { name: "Select all" });
    expect(box).toHaveAttribute("aria-checked", "false");
    await userEvent.click(box);
    expect(onCheckedChange).toHaveBeenCalledWith(true, expect.anything());
    expect(box).toHaveAttribute("aria-checked", "true");
  });

  it("uses the outline focus pattern, a 24px target and a coarse hit area", () => {
    render(<Checkbox aria-label="x" />);
    const cls = screen.getByRole("checkbox").className;
    expect(cls).toContain("size-6");
    expect(cls).toContain("coarse:before:-inset-2.5");
    expect(cls).toContain("focus-visible:outline-ring");
    expect(cls).toContain("border-input");
    expect(cls).not.toContain("ring-ring/50");
    expect(cls).not.toContain("outline-none");
  });

  it("does not toggle when disabled", async () => {
    render(<Checkbox aria-label="x" disabled />);
    const box = screen.getByRole("checkbox");
    await userEvent.click(box);
    expect(box).toHaveAttribute("aria-checked", "false");
  });
});
