import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { CarrierIcon } from "../carrier-icon";
import { StatusBadge } from "../status-badge";

describe("CarrierIcon", () => {
  it("renders the carrier label as text with a chart-tinted decorative icon", () => {
    const { container } = render(<CarrierIcon carrier="ups" />);
    expect(screen.getByText("UPS")).toBeInTheDocument();
    const svg = container.querySelector("svg");
    expect(svg).toHaveAttribute("aria-hidden", "true");
    expect(svg?.getAttribute("class")).toContain("text-chart-2");
  });

  it("falls back to muted icon and uppercased label for unknown carriers", () => {
    const { container } = render(<CarrierIcon carrier="dhl" />);
    expect(screen.getByText("DHL")).toBeInTheDocument();
    expect(container.querySelector("svg")?.getAttribute("class")).toContain("text-muted-foreground");
  });
});

describe("StatusBadge", () => {
  function markup(status: string) {
    const { container } = render(<StatusBadge status={status} />);
    const badge = container.firstElementChild as HTMLElement;
    return { cls: badge.className, icon: container.querySelector("svg")?.outerHTML };
  }

  it("gives in_transit/out_for_delivery and pre_transit/stale distinct treatments", () => {
    const a = markup("in_transit");
    const b = markup("out_for_delivery");
    expect(a.cls).not.toBe(b.cls);
    expect(a.icon).not.toBe(b.icon);
    expect(markup("pre_transit").cls).not.toBe(markup("stale").cls);
  });

  it("renders exception as destructive with a decorative icon", () => {
    const { container } = render(<StatusBadge status="exception" />);
    expect(screen.getByText("Exception")).toBeInTheDocument();
    expect(container.firstElementChild?.className).toContain("destructive");
    expect(container.querySelector("svg")).toHaveAttribute("aria-hidden", "true");
  });
});
