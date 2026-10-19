import { render, screen } from "@testing-library/react";
import { Check } from "lucide-react";
import { describe, expect, it } from "vitest";
import { Badge } from "../badge";

describe("Badge", () => {
  it("renders a decorative icon before the text for status variants", () => {
    render(
      <Badge variant="success" icon={Check}>
        Done
      </Badge>,
    );
    const badge = screen.getByText("Done");
    const svg = badge.querySelector("svg");
    expect(svg).not.toBeNull();
    expect(svg).toHaveAttribute("aria-hidden", "true");
    expect(badge.firstElementChild).toBe(svg);
    expect(badge.className).toContain("bg-success");
  });

  it("uses a solid warning fill with its foreground", () => {
    render(<Badge variant="warning">Heads up</Badge>);
    const cls = screen.getByText("Heads up").className;
    expect(cls).toContain("bg-warning");
    expect(cls).toContain("text-warning-foreground");
  });

  it("uses a solid destructive fill with its foreground", () => {
    render(<Badge variant="destructive">Failed</Badge>);
    const cls = screen.getByText("Failed").className;
    expect(cls).toContain("bg-destructive");
    expect(cls).toContain("text-destructive-foreground");
  });
});
