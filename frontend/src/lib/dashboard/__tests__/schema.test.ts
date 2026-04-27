import { describe, it, expect } from "vitest";
import { layoutSchema } from "@/lib/dashboard/schema";
import { MAX_WIDGETS } from "@/lib/dashboard/widget-types";

function uuid(n: number): string {
  const hex = n.toString(16).padStart(12, "0");
  return `00000000-0000-4000-8000-${hex}`;
}

describe("layoutSchema", () => {
  it("parses a valid layout", () => {
    const r = layoutSchema.safeParse({
      version: 1,
      widgets: [
        { id: uuid(1), type: "weather", x: 0, y: 0, w: 12, h: 3, config: {} },
        { id: uuid(2), type: "tasks-summary", x: 0, y: 3, w: 4, h: 2, config: {} },
      ],
    });
    expect(r.success).toBe(true);
  });

  it("rejects an invalid schema version", () => {
    const r = layoutSchema.safeParse({
      version: 2,
      widgets: [],
    });
    expect(r.success).toBe(false);
  });

  it("rejects more than MAX_WIDGETS widgets", () => {
    const widgets = Array.from({ length: MAX_WIDGETS + 1 }, (_, i) => ({
      id: uuid(i + 1),
      type: "weather",
      x: 0,
      y: i,
      w: 1,
      h: 1,
      config: {},
    }));
    const r = layoutSchema.safeParse({ version: 1, widgets });
    expect(r.success).toBe(false);
  });

  it("rejects widgets with geometry that exceeds the grid width", () => {
    const r = layoutSchema.safeParse({
      version: 1,
      widgets: [
        { id: uuid(1), type: "weather", x: 10, y: 0, w: 6, h: 2, config: {} },
      ],
    });
    expect(r.success).toBe(false);
  });

  it("rejects widgets with negative coordinates", () => {
    const r = layoutSchema.safeParse({
      version: 1,
      widgets: [
        { id: uuid(1), type: "weather", x: -1, y: 0, w: 4, h: 2, config: {} },
      ],
    });
    expect(r.success).toBe(false);
  });
});
