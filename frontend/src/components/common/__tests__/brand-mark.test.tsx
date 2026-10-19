import { describe, it, expect } from "vitest";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { render, screen } from "@testing-library/react";
import { BrandMark } from "../brand-mark";

const colourValues = (root: ParentNode): string[] => {
  const values = new Set<string>();
  root.querySelectorAll("[fill], [stop-color]").forEach((el) => {
    for (const attr of ["fill", "stop-color"]) {
      const v = el.getAttribute(attr);
      if (v) values.add(v.trim().toLowerCase());
    }
  });
  return [...values].sort();
};

describe("BrandMark", () => {
  const faviconSource = readFileSync(resolve(process.cwd(), "public/favicon.svg"), "utf8");
  const favicon = new DOMParser().parseFromString(faviconSource, "image/svg+xml");

  it("exposes an accessible name", () => {
    render(<BrandMark />);
    expect(screen.getByRole("img", { name: "Home Hub" })).toBeInTheDocument();
  });

  it("uses the same fill/stop-color set as public/favicon.svg", () => {
    const { container } = render(<BrandMark />);
    const fromComponent = colourValues(container);
    expect(fromComponent.length).toBeGreaterThan(0);
    expect(fromComponent).toEqual(colourValues(favicon));
  });

  it("no longer uses the indigo mark", () => {
    const { container } = render(<BrandMark />);
    expect(container.innerHTML.toLowerCase()).not.toContain("#6366f1");
    expect(faviconSource.toLowerCase()).not.toContain("#6366f1");
  });
});
