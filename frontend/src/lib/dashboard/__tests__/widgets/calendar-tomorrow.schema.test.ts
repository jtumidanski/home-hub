import { describe, it, expect } from "vitest";
import { calendarTomorrowWidget } from "@/lib/dashboard/widgets/calendar-tomorrow";

describe("calendar-tomorrow widget definition", () => {
  it("declares metadata", () => {
    expect(calendarTomorrowWidget.type).toBe("calendar-tomorrow");
    expect(calendarTomorrowWidget.dataScope).toBe("household");
    expect(calendarTomorrowWidget.defaultSize).toEqual({ w: 4, h: 3 });
    expect(calendarTomorrowWidget.defaultConfig).toEqual({ includeAllDay: true, limit: 5 });
  });

  it("schema enforces limit bounds", () => {
    expect(calendarTomorrowWidget.configSchema.safeParse({ limit: 0 }).success).toBe(false);
    expect(calendarTomorrowWidget.configSchema.safeParse({ limit: 11 }).success).toBe(false);
  });
});
