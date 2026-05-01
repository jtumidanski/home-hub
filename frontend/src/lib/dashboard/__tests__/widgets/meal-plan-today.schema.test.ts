import { describe, it, expect } from "vitest";
import { mealPlanTodayWidget } from "@/lib/dashboard/widgets/meal-plan-today";

describe("meal-plan-today widget extension", () => {
  it("schema accepts view: today-detail", () => {
    expect(mealPlanTodayWidget.configSchema.safeParse({ horizonDays: 3, view: "today-detail" }).success).toBe(true);
  });

  it("schema accepts view: list", () => {
    expect(mealPlanTodayWidget.configSchema.safeParse({ horizonDays: 1, view: "list" }).success).toBe(true);
  });

  it("schema rejects unknown view", () => {
    expect(mealPlanTodayWidget.configSchema.safeParse({ horizonDays: 1, view: "grid" }).success).toBe(false);
  });

  it("default view is list when omitted", () => {
    const parsed = mealPlanTodayWidget.configSchema.parse({ horizonDays: 1 });
    expect(parsed.view).toBe("list");
  });

  it("default config still has horizonDays:1 and view:list", () => {
    expect(mealPlanTodayWidget.defaultConfig).toEqual({ horizonDays: 1, view: "list" });
  });
});
