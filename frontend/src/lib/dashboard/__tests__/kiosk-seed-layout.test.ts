import { describe, it, expect } from "vitest";
import { kioskSeedLayout } from "@/lib/dashboard/kiosk-seed-layout";
import { layoutSchema } from "@/lib/dashboard/schema";
import { findWidget } from "@/lib/dashboard/widget-registry";
import { GRID_COLUMNS } from "@/lib/dashboard/widget-types";

describe("kioskSeedLayout", () => {
  it("passes layoutSchema validation", () => {
    expect(layoutSchema.safeParse(kioskSeedLayout()).success).toBe(true);
  });

  it("every widget type is in the registry", () => {
    for (const w of kioskSeedLayout().widgets) {
      expect(findWidget(w.type), `missing registry entry for ${w.type}`).toBeDefined();
    }
  });

  it("every widget satisfies x + w <= GRID_COLUMNS", () => {
    for (const w of kioskSeedLayout().widgets) {
      expect(w.x + w.w, `widget ${w.type} overflows`).toBeLessThanOrEqual(GRID_COLUMNS);
    }
  });

  it("each invocation produces fresh UUIDs", () => {
    const a = kioskSeedLayout().widgets.map((w) => w.id);
    const b = kioskSeedLayout().widgets.map((w) => w.id);
    expect(a.filter((id) => b.includes(id))).toEqual([]);
  });

  it("places the expected widgets in the four columns", () => {
    const types = kioskSeedLayout().widgets.map((w) => w.type);
    expect(types).toEqual([
      "weather", "meal-plan-today", "tasks-today",
      "calendar-today",
      "weather-tomorrow", "calendar-tomorrow", "tasks-tomorrow",
      "reminders-today",
    ]);
  });

  it("meal-plan-today seeded with view: today-detail and horizonDays: 3", () => {
    const meal = kioskSeedLayout().widgets.find((w) => w.type === "meal-plan-today")!;
    expect(meal.config).toMatchObject({ view: "today-detail", horizonDays: 3 });
  });
});
