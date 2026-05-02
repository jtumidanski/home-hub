import { describe, it, expect } from "vitest";
import { remindersTodayWidget } from "@/lib/dashboard/widgets/reminders-today";

describe("reminders-today widget definition", () => {
  it("declares the registry metadata", () => {
    expect(remindersTodayWidget.type).toBe("reminders-today");
    expect(remindersTodayWidget.dataScope).toBe("household");
    expect(remindersTodayWidget.defaultSize).toEqual({ w: 3, h: 4 });
    expect(remindersTodayWidget.minSize).toEqual({ w: 3, h: 2 });
    expect(remindersTodayWidget.maxSize).toEqual({ w: 6, h: 8 });
    expect(remindersTodayWidget.defaultConfig).toEqual({ limit: 5 });
  });

  it("schema enforces limit bounds", () => {
    expect(remindersTodayWidget.configSchema.safeParse({ limit: 0 }).success).toBe(false);
    expect(remindersTodayWidget.configSchema.safeParse({ limit: 11 }).success).toBe(false);
    expect(remindersTodayWidget.configSchema.safeParse({ limit: 5 }).success).toBe(true);
  });
});
