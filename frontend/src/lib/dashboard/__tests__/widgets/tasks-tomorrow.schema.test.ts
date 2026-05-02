import { describe, it, expect } from "vitest";
import { tasksTomorrowWidget } from "@/lib/dashboard/widgets/tasks-tomorrow";

describe("tasks-tomorrow widget definition", () => {
  it("declares metadata", () => {
    expect(tasksTomorrowWidget.type).toBe("tasks-tomorrow");
    expect(tasksTomorrowWidget.dataScope).toBe("household");
    expect(tasksTomorrowWidget.defaultSize).toEqual({ w: 3, h: 3 });
    expect(tasksTomorrowWidget.defaultConfig).toEqual({ limit: 5 });
  });

  it("schema enforces limit bounds", () => {
    expect(tasksTomorrowWidget.configSchema.safeParse({ limit: 0 }).success).toBe(false);
    expect(tasksTomorrowWidget.configSchema.safeParse({ limit: 11 }).success).toBe(false);
  });
});
