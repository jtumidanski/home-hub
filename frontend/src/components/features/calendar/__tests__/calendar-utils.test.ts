import { describe, it, expect } from "vitest";
import type { CalendarEvent } from "@/types/models/calendar";
import {
  HOUR_HEIGHT,
  getStartOfWeek,
  getWeekDays,
  isSameDay,
  formatHour,
  formatDateRange,
  getEventPosition,
  layoutOverlappingEvents,
  getEventsForDay,
  getTimeInZone,
  getDateInZone,
} from "../calendar-utils";

function makeEvent(overrides: Partial<CalendarEvent["attributes"]> & { id?: string } = {}): CalendarEvent {
  const { id = "evt-1", ...attrs } = overrides;
  return {
    id,
    type: "calendar-events",
    attributes: {
      title: "Test Event",
      description: null,
      startTime: "2026-03-26T10:00:00Z",
      endTime: "2026-03-26T11:00:00Z",
      allDay: false,
      location: null,
      visibility: "default",
      isOwner: true,
      userDisplayName: "Test User",
      userColor: "#4285F4",
      ...attrs,
    },
  };
}

describe("getStartOfWeek", () => {
  it("returns Sunday for a Wednesday", () => {
    // March 25, 2026 is a Wednesday
    const wed = new Date(2026, 2, 25);
    const start = getStartOfWeek(wed);
    expect(start.getDay()).toBe(0); // Sunday
    expect(start.getDate()).toBe(22);
  });

  it("returns the same day for a Sunday", () => {
    const sun = new Date(2026, 2, 22);
    const start = getStartOfWeek(sun);
    expect(start.getDay()).toBe(0);
    expect(start.getDate()).toBe(22);
  });

  it("sets time to midnight", () => {
    const date = new Date(2026, 2, 25, 14, 30);
    const start = getStartOfWeek(date);
    expect(start.getHours()).toBe(0);
    expect(start.getMinutes()).toBe(0);
    expect(start.getSeconds()).toBe(0);
  });
});

describe("getWeekDays", () => {
  it("returns 7 consecutive days", () => {
    const start = new Date(2026, 2, 22);
    const days = getWeekDays(start);
    expect(days).toHaveLength(7);
    expect(days[0]!.getDate()).toBe(22);
    expect(days[6]!.getDate()).toBe(28);
  });
});

describe("isSameDay", () => {
  it("returns true for same date", () => {
    const a = new Date(2026, 2, 25, 10, 0);
    const b = new Date(2026, 2, 25, 22, 30);
    expect(isSameDay(a, b)).toBe(true);
  });

  it("returns false for different dates", () => {
    const a = new Date(2026, 2, 25);
    const b = new Date(2026, 2, 26);
    expect(isSameDay(a, b)).toBe(false);
  });
});

describe("formatHour", () => {
  it("formats midnight as 12 AM", () => {
    expect(formatHour(0)).toBe("12 AM");
  });

  it("formats morning hours", () => {
    expect(formatHour(9)).toBe("9 AM");
  });

  it("formats noon as 12 PM", () => {
    expect(formatHour(12)).toBe("12 PM");
  });

  it("formats afternoon hours", () => {
    expect(formatHour(15)).toBe("3 PM");
  });
});

describe("formatDateRange", () => {
  it("formats range within same year", () => {
    const start = new Date(2026, 2, 22);
    const end = new Date(2026, 2, 29);
    const result = formatDateRange(start, end);
    expect(result).toContain("March");
    expect(result).toContain("22");
    expect(result).toContain("28");
    expect(result).toContain("2026");
  });
});

describe("getEventPosition", () => {
  it("positions an event at the correct top offset", () => {
    // Event at 10 AM, START_HOUR is 6, so 4 hours down
    const event = makeEvent({ startTime: "2026-03-26T10:00:00", endTime: "2026-03-26T11:00:00" });
    const day = new Date(2026, 2, 26);
    const { top, height } = getEventPosition(event, day);
    expect(top).toBe(4 * HOUR_HEIGHT);
    expect(height).toBe(HOUR_HEIGHT);
  });

  it("enforces minimum height of 15px", () => {
    // 10 minute event
    const event = makeEvent({ startTime: "2026-03-26T10:00:00", endTime: "2026-03-26T10:10:00" });
    const day = new Date(2026, 2, 26);
    const { height } = getEventPosition(event, day);
    expect(height).toBe(HOUR_HEIGHT / 4); // 15px minimum
  });

  it("clamps top to 0 for events before START_HOUR", () => {
    const event = makeEvent({ startTime: "2026-03-26T04:00:00", endTime: "2026-03-26T05:00:00" });
    const day = new Date(2026, 2, 26);
    const { top } = getEventPosition(event, day);
    expect(top).toBe(0);
  });
});

describe("layoutOverlappingEvents", () => {
  it("returns empty array for no events", () => {
    expect(layoutOverlappingEvents([])).toEqual([]);
  });

  it("assigns a single event to column 0 with totalColumns 1", () => {
    const events = [makeEvent()];
    const result = layoutOverlappingEvents(events);
    expect(result).toHaveLength(1);
    expect(result[0]!.column).toBe(0);
    expect(result[0]!.totalColumns).toBe(1);
  });

  it("puts overlapping events in separate columns", () => {
    const events = [
      makeEvent({ id: "a", startTime: "2026-03-26T10:00:00Z", endTime: "2026-03-26T11:00:00Z" }),
      makeEvent({ id: "b", startTime: "2026-03-26T10:30:00Z", endTime: "2026-03-26T11:30:00Z" }),
    ];
    const result = layoutOverlappingEvents(events);
    expect(result).toHaveLength(2);
    expect(result[0]!.totalColumns).toBe(2);
    expect(result[1]!.totalColumns).toBe(2);
    expect(result[0]!.column).not.toBe(result[1]!.column);
  });

  it("puts non-overlapping events in the same column", () => {
    const events = [
      makeEvent({ id: "a", startTime: "2026-03-26T10:00:00Z", endTime: "2026-03-26T11:00:00Z" }),
      makeEvent({ id: "b", startTime: "2026-03-26T12:00:00Z", endTime: "2026-03-26T13:00:00Z" }),
    ];
    const result = layoutOverlappingEvents(events);
    expect(result).toHaveLength(2);
    // Non-overlapping: separate groups, each with 1 column
    expect(result[0]!.totalColumns).toBe(1);
    expect(result[1]!.totalColumns).toBe(1);
  });

  it("handles three overlapping events in three columns", () => {
    const events = [
      makeEvent({ id: "a", startTime: "2026-03-26T10:00:00Z", endTime: "2026-03-26T12:00:00Z" }),
      makeEvent({ id: "b", startTime: "2026-03-26T10:30:00Z", endTime: "2026-03-26T12:00:00Z" }),
      makeEvent({ id: "c", startTime: "2026-03-26T11:00:00Z", endTime: "2026-03-26T12:00:00Z" }),
    ];
    const result = layoutOverlappingEvents(events);
    expect(result).toHaveLength(3);
    expect(result[0]!.totalColumns).toBe(3);
    const columns = new Set(result.map((r) => r.column));
    expect(columns.size).toBe(3);
  });
});

describe("getEventsForDay", () => {
  it("separates all-day from timed events", () => {
    const day = new Date(2026, 2, 26);
    const events = [
      makeEvent({ id: "timed", startTime: "2026-03-26T10:00:00", endTime: "2026-03-26T11:00:00", allDay: false }),
      makeEvent({ id: "allday", startTime: "2026-03-26T00:00:00", endTime: "2026-03-27T00:00:00", allDay: true }),
    ];
    const { allDay, timed } = getEventsForDay(events, day);
    expect(allDay).toHaveLength(1);
    expect(allDay[0]!.id).toBe("allday");
    expect(timed).toHaveLength(1);
    expect(timed[0]!.id).toBe("timed");
  });

  it("does not show all-day event on the day before", () => {
    const dayBefore = new Date(2026, 2, 25);
    const events = [
      makeEvent({ id: "allday", startTime: "2026-03-26T00:00:00Z", endTime: "2026-03-27T00:00:00Z", allDay: true }),
    ];
    const { allDay } = getEventsForDay(events, dayBefore);
    expect(allDay).toHaveLength(0);
  });

  it("does not show all-day event on the end date (exclusive)", () => {
    const endDay = new Date(2026, 2, 27);
    const events = [
      makeEvent({ id: "allday", startTime: "2026-03-26T00:00:00Z", endTime: "2026-03-27T00:00:00Z", allDay: true }),
    ];
    const { allDay } = getEventsForDay(events, endDay);
    expect(allDay).toHaveLength(0);
  });

  it("shows multi-day all-day event on intermediate days", () => {
    const middleDay = new Date(2026, 2, 27);
    const events = [
      makeEvent({ id: "multi", startTime: "2026-03-26T00:00:00Z", endTime: "2026-03-29T00:00:00Z", allDay: true }),
    ];
    const { allDay } = getEventsForDay(events, middleDay);
    expect(allDay).toHaveLength(1);
  });

  it("excludes events on different days", () => {
    const day = new Date(2026, 2, 26);
    const events = [
      makeEvent({ id: "other", startTime: "2026-03-27T10:00:00", endTime: "2026-03-27T11:00:00" }),
    ];
    const { allDay, timed } = getEventsForDay(events, day);
    expect(allDay).toHaveLength(0);
    expect(timed).toHaveLength(0);
  });
});

describe("getTimeInZone", () => {
  it("returns local time when no timezone provided", () => {
    const date = new Date(2026, 2, 26, 14, 30);
    const { hours, minutes } = getTimeInZone(date);
    expect(hours).toBe(14);
    expect(minutes).toBe(30);
  });

  it("converts to specified timezone", () => {
    // UTC midnight should be different in US Eastern (UTC-4 in March)
    const utcMidnight = new Date("2026-03-26T00:00:00Z");
    const { hours } = getTimeInZone(utcMidnight, "America/New_York");
    // UTC midnight = 8 PM previous day in EDT (UTC-4)
    expect(hours).toBe(20);
  });

  it("falls back to local time on invalid timezone", () => {
    const date = new Date(2026, 2, 26, 14, 30);
    const { hours, minutes } = getTimeInZone(date, "Invalid/Timezone");
    expect(hours).toBe(14);
    expect(minutes).toBe(30);
  });
});

describe("getDateInZone", () => {
  it("returns local date when no timezone provided", () => {
    const date = new Date(2026, 2, 26);
    const { year, month, day } = getDateInZone(date);
    expect(year).toBe(2026);
    expect(month).toBe(3);
    expect(day).toBe(26);
  });

  it("converts date to specified timezone", () => {
    // UTC midnight March 26 is still March 25 in US Eastern
    const utcMidnight = new Date("2026-03-26T00:00:00Z");
    const { day } = getDateInZone(utcMidnight, "America/New_York");
    expect(day).toBe(25);
  });
});
