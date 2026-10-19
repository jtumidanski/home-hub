import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { Sparkles } from "lucide-react";
import { StatusBanner } from "../status-banner";

describe("StatusBanner", () => {
  it("renders destructive as an alert with an aria-hidden icon and text", () => {
    render(<StatusBanner variant="destructive">Save failed</StatusBanner>);
    const alert = screen.getByRole("alert");
    expect(alert).toHaveTextContent("Save failed");
    expect(alert).toHaveClass("border-destructive");
    const icon = alert.querySelector("svg");
    expect(icon).toHaveAttribute("aria-hidden", "true");
  });

  it.each([
    ["success", "border-success"],
    ["warning", "border-warning"],
    ["info", "border-info"],
  ] as const)("renders %s as a polite status", (variant, cls) => {
    render(<StatusBanner variant={variant}>Heads up</StatusBanner>);
    const status = screen.getByRole("status");
    expect(status).toHaveClass(cls);
    expect(status.querySelector("svg")).toBeInTheDocument();
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("defaults to info", () => {
    render(<StatusBanner>Note</StatusBanner>);
    expect(screen.getByRole("status")).toHaveAttribute("data-variant", "info");
  });

  it("accepts a custom icon", () => {
    render(
      <StatusBanner icon={Sparkles} variant="success">
        New
      </StatusBanner>,
    );
    expect(screen.getByRole("status").querySelector("svg")).toHaveClass("lucide-sparkles");
  });
});
