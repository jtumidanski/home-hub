import { describe, it, expect } from "vitest";
import { TRACKER_COLORS, SWATCH_CLASSES, swatchClass, type SwatchRole } from "@/lib/constants/swatches";

const ROLES: SwatchRole[] = ["solid", "soft", "borderLeft"];

describe("TRACKER_COLORS", () => {
  it("has 16 unique entries in the TrackerColor order", () => {
    expect(TRACKER_COLORS).toHaveLength(16);
    expect(new Set(TRACKER_COLORS).size).toBe(16);
    expect([...TRACKER_COLORS]).toEqual([
      "red", "orange", "amber", "yellow", "lime", "green", "emerald", "teal",
      "cyan", "blue", "indigo", "violet", "purple", "fuchsia", "pink", "rose",
    ]);
  });
});

describe("SWATCH_CLASSES", () => {
  it.each(ROLES)("maps every colour for role %s to a literal containing swatch-<colour>", (role) => {
    for (const c of TRACKER_COLORS) {
      const cls = SWATCH_CLASSES[role][c];
      expect(cls).toBeTruthy();
      expect(cls).toContain(`swatch-${c}`);
    }
  });

  it("uses the expected utility prefixes per role", () => {
    expect(SWATCH_CLASSES.solid.red).toBe("bg-swatch-red");
    expect(SWATCH_CLASSES.soft.red).toBe("bg-swatch-red-soft");
    expect(SWATCH_CLASSES.borderLeft.red).toBe("border-l-swatch-red");
  });
});

describe("swatchClass", () => {
  it("returns the mapped class for a known colour", () => {
    expect(swatchClass("red", "soft")).toBe("bg-swatch-red-soft");
  });

  it("falls back to neutral for an unknown colour", () => {
    expect(swatchClass("chartreuse", "solid")).toBe("bg-muted-foreground");
    expect(swatchClass("chartreuse", "soft")).toBe("bg-muted");
    expect(swatchClass("chartreuse", "borderLeft")).toBe("border-l-muted-foreground");
  });

  it("falls back for the empty string without containing 'undefined'", () => {
    expect(swatchClass("", "soft")).toBe("bg-muted");
  });

  it.each(["constructor", "toString", "__proto__", "hasOwnProperty"])(
    "treats prototype key %s as unknown rather than throwing",
    (key) => {
      for (const role of ROLES) {
        const cls = swatchClass(key, role);
        expect(cls).not.toContain("undefined");
        expect(cls).not.toContain("[object");
        expect(cls).toMatch(/muted/);
      }
    },
  );
});
