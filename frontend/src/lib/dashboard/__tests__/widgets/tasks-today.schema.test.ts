import { describe, it, expect } from "vitest";
import { tasksTodayWidget } from "@/lib/dashboard/widgets/tasks-today";

describe("tasks-today widget definition", () => {
  it("declares the registry metadata", () => {
    expect(tasksTodayWidget.type).toBe("tasks-today");
    expect(tasksTodayWidget.displayName).toBe("Today's Tasks");
    expect(tasksTodayWidget.dataScope).toBe("household");
    expect(tasksTodayWidget.defaultSize).toEqual({ w: 4, h: 4 });
    expect(tasksTodayWidget.minSize).toEqual({ w: 3, h: 2 });
    expect(tasksTodayWidget.maxSize).toEqual({ w: 6, h: 8 });
  });

  it("default config sets includeCompleted to true", () => {
    expect(tasksTodayWidget.defaultConfig).toEqual({ includeCompleted: true });
  });

  it("schema accepts a custom title and rejects long titles", () => {
    expect(tasksTodayWidget.configSchema.safeParse({ title: "Custom" }).success).toBe(true);
    expect(tasksTodayWidget.configSchema.safeParse({ title: "x".repeat(81) }).success).toBe(false);
  });

  it("schema applies includeCompleted default", () => {
    const parsed = tasksTodayWidget.configSchema.parse({});
    expect(parsed.includeCompleted).toBe(true);
  });
});
