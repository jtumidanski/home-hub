import { describe, it, expect } from "vitest";
import { seedLayout } from "@/lib/dashboard/seed-layout";
import { layoutSchema } from "@/lib/dashboard/schema";

describe("seedLayout", () => {
  it("produces a layout with version 1 and nine widgets", () => {
    const layout = seedLayout();
    expect(layout.version).toBe(1);
    expect(layout.widgets).toHaveLength(9);
  });

  it("places each widget at the expected type/position", () => {
    const layout = seedLayout();
    const rows = layout.widgets.map((w) => ({
      type: w.type,
      x: w.x,
      y: w.y,
      w: w.w,
      h: w.h,
    }));
    expect(rows).toEqual([
      { type: "weather",           x: 0, y: 0, w: 12, h: 3 },
      { type: "tasks-summary",     x: 0, y: 3, w: 4,  h: 2 },
      { type: "reminders-summary", x: 4, y: 3, w: 4,  h: 2 },
      { type: "overdue-summary",   x: 8, y: 3, w: 4,  h: 2 },
      { type: "meal-plan-today",   x: 0, y: 5, w: 4,  h: 3 },
      { type: "habits-today",      x: 4, y: 5, w: 4,  h: 3 },
      { type: "packages-summary",  x: 8, y: 5, w: 4,  h: 3 },
      { type: "calendar-today",    x: 0, y: 8, w: 6,  h: 3 },
      { type: "workout-today",     x: 6, y: 8, w: 6,  h: 3 },
    ]);
  });

  it("embeds the configured status / filter / horizon values", () => {
    const byType = new Map(seedLayout().widgets.map((w) => [w.type, w.config]));
    expect(byType.get("tasks-summary")).toMatchObject({ status: "pending", title: "Pending Tasks" });
    expect(byType.get("reminders-summary")).toMatchObject({ filter: "active", title: "Active Reminders" });
    expect(byType.get("overdue-summary")).toMatchObject({ title: "Overdue" });
    expect(byType.get("meal-plan-today")).toMatchObject({ horizonDays: 1 });
    expect(byType.get("calendar-today")).toMatchObject({ horizonDays: 1, includeAllDay: true });
    expect(byType.get("weather")).toMatchObject({ units: "imperial", location: null });
  });

  it("produces fresh UUIDs on every invocation", () => {
    const first = seedLayout().widgets.map((w) => w.id);
    const second = seedLayout().widgets.map((w) => w.id);
    expect(first).toHaveLength(9);
    expect(second).toHaveLength(9);
    // No id should be shared between the two calls.
    const intersection = first.filter((id) => second.includes(id));
    expect(intersection).toEqual([]);
    // Every id within a single call is also unique.
    expect(new Set(first).size).toBe(first.length);
  });

  it("passes layoutSchema validation", () => {
    const r = layoutSchema.safeParse(seedLayout());
    expect(r.success).toBe(true);
  });
});
