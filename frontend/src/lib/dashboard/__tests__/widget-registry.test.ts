import { describe, it, expect } from "vitest";
import { WIDGET_TYPES } from "@/lib/dashboard/widget-types";
import { widgetRegistry, findWidget } from "@/lib/dashboard/widget-registry";

describe("widgetRegistry", () => {
  it("has one registry entry per widget type", () => {
    const registryTypes = widgetRegistry.map((w) => w.type).sort();
    expect(registryTypes).toEqual([...WIDGET_TYPES].sort());
  });

  it("has no duplicate types", () => {
    const types = widgetRegistry.map((w) => w.type);
    expect(new Set(types).size).toBe(types.length);
  });

  it("findWidget returns the matching entry", () => {
    const w = findWidget("weather");
    expect(w).toBeDefined();
    expect(w?.type).toBe("weather");
  });

  it("findWidget returns undefined for an unknown type", () => {
    expect(findWidget("nope")).toBeUndefined();
  });

  it("contains the 5 task-046 widgets", () => {
    for (const t of ["tasks-today", "reminders-today", "weather-tomorrow", "calendar-tomorrow", "tasks-tomorrow"] as const) {
      expect(findWidget(t)).toBeDefined();
    }
  });
});
