import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { LossyConfigBadge } from "@/components/features/dashboard-widgets/lossy-config-badge";

describe("LossyConfigBadge", () => {
  it("renders with a tooltip title and label", () => {
    render(<LossyConfigBadge />);
    const badge = screen.getByText("reduced to defaults");
    expect(badge).toBeInTheDocument();
    expect(badge.closest("[title]")).toHaveAttribute(
      "title",
      expect.stringContaining("reduced to defaults"),
    );
  });
});
