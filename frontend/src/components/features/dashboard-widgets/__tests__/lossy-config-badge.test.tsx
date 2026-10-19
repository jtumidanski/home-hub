import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { LossyConfigBadge } from "@/components/features/dashboard-widgets/lossy-config-badge";

describe("LossyConfigBadge", () => {
  it("renders a visible label with the explanation as real text", () => {
    render(<LossyConfigBadge />);
    const badge = screen.getByText("Defaults used");
    expect(badge).toBeInTheDocument();
    const described = badge.closest("[aria-describedby]");
    expect(described).not.toBeNull();
    const id = described!.getAttribute("aria-describedby")!;
    expect(document.getElementById(id)).toHaveTextContent("reduced to defaults");
  });
});
