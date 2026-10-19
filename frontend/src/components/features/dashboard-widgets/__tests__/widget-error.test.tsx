import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { WidgetError } from "@/components/features/dashboard-widgets/widget-error";

describe("WidgetError", () => {
  it("renders the message as an alert with a non-colour icon cue", () => {
    render(<WidgetError message="Failed to load tasks" />);
    const alert = screen.getByRole("alert");
    expect(alert).toHaveTextContent("Failed to load tasks");
    expect(alert.querySelector("svg")).toHaveAttribute("aria-hidden", "true");
  });
});
