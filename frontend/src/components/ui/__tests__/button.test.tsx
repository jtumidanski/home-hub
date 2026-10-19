import { join } from "node:path";
import { readdirSync, readFileSync } from "node:fs";
import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { Button } from "../button";

describe("Button", () => {
  it("uses the outline focus pattern and a coarse-pointer target", () => {
    render(<Button>Save</Button>);
    const cls = screen.getByRole("button").className;
    expect(cls).toContain("focus-visible:outline-ring");
    expect(cls).toContain("coarse:min-h-11");
    expect(cls).not.toContain("ring-ring/50");
    expect(cls).not.toContain("outline-none");
  });

  it("sizes icon buttons for coarse pointers", () => {
    render(<Button size="icon">x</Button>);
    expect(screen.getByRole("button").className).toContain("coarse:size-11");
  });

  it("keeps icon-xs compact on fine pointers", () => {
    render(<Button size="icon-xs">x</Button>);
    expect(screen.getByRole("button").className).toContain("size-6");
  });
});

describe("ui source", () => {
  it("contains no ring-ring/50 focus pattern", () => {
    const dir = join(process.cwd(), "src/components/ui");
    const offenders = readdirSync(dir)
      .filter((f) => f.endsWith(".tsx"))
      .filter((f) => readFileSync(join(dir, f), "utf8").includes("ring-ring/50"));
    expect(offenders).toEqual([]);
  });
});
