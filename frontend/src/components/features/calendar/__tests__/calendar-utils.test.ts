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
  getStartOfMonth,
  addMonths,
  getMonthGridDays,
  getMonthGridRange,
  isSameMonth,
  formatMonthYear,
  formatChipTime,
  toDayKey,
  bucketEventsByDay,
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
      sourceId: "",
      connectionId: "",
      isRecurring: false,
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
      makeEvent({ id: "allday", startTime: "2026-03-26T00:00:00", endTime: "2026-03-26T00:00:00", allDay: true }),
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
      makeEvent({ id: "allday", startTime: "2026-03-26T00:00:00Z", endTime: "2026-03-26T00:00:00Z", allDay: true }),
    ];
    const { allDay } = getEventsForDay(events, dayBefore);
    expect(allDay).toHaveLength(0);
  });

  it("does not show all-day event on the day after end date (inclusive)", () => {
    const dayAfter = new Date(2026, 2, 27);
    const events = [
      makeEvent({ id: "allday", startTime: "2026-03-26T00:00:00Z", endTime: "2026-03-26T00:00:00Z", allDay: true }),
    ];
    const { allDay } = getEventsForDay(events, dayAfter);
    expect(allDay).toHaveLength(0);
  });

  it("shows multi-day all-day event on all days including end date", () => {
    const endDay = new Date(2026, 2, 29);
    const events = [
      makeEvent({ id: "multi", startTime: "2026-03-26T00:00:00Z", endTime: "2026-03-29T00:00:00Z", allDay: true }),
    ];
    const { allDay } = getEventsForDay(events, endDay);
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

describe("getStartOfMonth", () => {
  it("returns the 1st of the month at local midnight", () => {
    const d = new Date(2026, 7, 14, 15, 30); // Aug 14, 2026 3:30pm
    const start = getStartOfMonth(d);
    expect(start.getFullYear()).toBe(2026);
    expect(start.getMonth()).toBe(7); // August
    expect(start.getDate()).toBe(1);
    expect(start.getHours()).toBe(0);
    expect(start.getMinutes()).toBe(0);
    expect(start.getSeconds()).toBe(0);
  });

  it("returns the same month when already on the 1st", () => {
    const start = getStartOfMonth(new Date(2026, 0, 1));
    expect(start.getMonth()).toBe(0);
    expect(start.getDate()).toBe(1);
  });
});

describe("addMonths", () => {
  it("advances by one calendar month", () => {
    const next = addMonths(new Date(2026, 7, 1), 1);
    expect(next.getFullYear()).toBe(2026);
    expect(next.getMonth()).toBe(8); // September
    expect(next.getDate()).toBe(1);
  });

  it("rolls over the year going forward (Dec -> Jan)", () => {
    const next = addMonths(new Date(2026, 11, 1), 1);
    expect(next.getFullYear()).toBe(2027);
    expect(next.getMonth()).toBe(0); // January
  });

  it("rolls back the year going backward (Jan -> Dec)", () => {
    const prev = addMonths(new Date(2026, 0, 1), -1);
    expect(prev.getFullYear()).toBe(2025);
    expect(prev.getMonth()).toBe(11); // December
  });
});

describe("getMonthGridDays", () => {
  it("returns a 35-cell grid for a month that starts on Sunday (March 2026)", () => {
    // March 1, 2026 is a Sunday; March has 31 days.
    const days = getMonthGridDays(new Date(2026, 2, 1));
    expect(days).toHaveLength(35);
    expect(days[0]!.getDay()).toBe(0); // first cell is Sunday
    expect(days[days.length - 1]!.getDay()).toBe(6); // last cell is Saturday
    // First cell is March 1 (in-month), last cell is April 4 (trailing).
    expect(days[0]!.getMonth()).toBe(2);
    expect(days[0]!.getDate()).toBe(1);
    expect(days[34]!.getMonth()).toBe(3); // April
    expect(days[34]!.getDate()).toBe(4);
  });

  it("returns a 42-cell grid with leading and trailing days (August 2026)", () => {
    // August 1, 2026 is a Saturday; grid starts on the prior Sunday (Jul 26).
    const days = getMonthGridDays(new Date(2026, 7, 1));
    expect(days).toHaveLength(42);
    expect(days[0]!.getDay()).toBe(0);
    expect(days[days.length - 1]!.getDay()).toBe(6);
    // Leading day belongs to July, trailing day belongs to September.
    expect(days[0]!.getMonth()).toBe(6); // July
    expect(days[0]!.getDate()).toBe(26);
    expect(days[41]!.getMonth()).toBe(8); // September
    expect(days[41]!.getDate()).toBe(5);
  });

  it("produces cells at local midnight", () => {
    const days = getMonthGridDays(new Date(2026, 7, 1));
    for (const d of days) {
      expect(d.getHours()).toBe(0);
      expect(d.getMinutes()).toBe(0);
    }
  });
});

describe("getMonthGridRange", () => {
  it("spans the first grid day to one day past the last (exclusive end)", () => {
    const { start, end } = getMonthGridRange(new Date(2026, 7, 1));
    // start = Jul 26 2026, end = Sep 6 2026 (last grid day Sep 5 + 1).
    expect(start.getMonth()).toBe(6);
    expect(start.getDate()).toBe(26);
    expect(end.getMonth()).toBe(8);
    expect(end.getDate()).toBe(6);
  });
});

describe("isSameMonth", () => {
  const anchor = new Date(2026, 7, 1); // August 2026

  it("is true for a day within the anchor's month", () => {
    expect(isSameMonth(new Date(2026, 7, 15), anchor)).toBe(true);
  });

  it("is false for a leading adjacent-month day", () => {
    expect(isSameMonth(new Date(2026, 6, 26), anchor)).toBe(false); // July 26
  });

  it("is false for a trailing adjacent-month day", () => {
    expect(isSameMonth(new Date(2026, 8, 5), anchor)).toBe(false); // Sept 5
  });

  it("respects the year (Dec vs next Jan)", () => {
    expect(isSameMonth(new Date(2026, 11, 31), new Date(2027, 0, 1))).toBe(false);
  });
});

describe("formatMonthYear", () => {
  it("formats as full month name and year", () => {
    expect(formatMonthYear(new Date(2026, 5, 1))).toBe("June 2026");
  });
});

describe("formatChipTime", () => {
  it("drops :00 minutes on the hour (AM)", () => {
    expect(formatChipTime("2026-03-26T09:00:00Z", "UTC")).toBe("9a");
  });

  it("keeps non-zero minutes", () => {
    expect(formatChipTime("2026-03-26T09:30:00Z", "UTC")).toBe("9:30a");
  });

  it("formats noon as 12p", () => {
    expect(formatChipTime("2026-03-26T12:00:00Z", "UTC")).toBe("12p");
  });

  it("formats afternoon times as PM", () => {
    expect(formatChipTime("2026-03-26T14:30:00Z", "UTC")).toBe("2:30p");
  });

  it("formats midnight as 12a", () => {
    expect(formatChipTime("2026-03-26T00:00:00Z", "UTC")).toBe("12a");
  });

  it("is timezone-aware", () => {
    // 14:00 UTC is 10:00 in America/New_York (EDT, UTC-4) on this date.
    expect(formatChipTime("2026-03-26T14:00:00Z", "America/New_York")).toBe("10a");
  });
});

describe("toDayKey", () => {
  it("formats a local date as YYYY-MM-DD with zero padding", () => {
    expect(toDayKey(new Date(2026, 7, 5))).toBe("2026-08-05");
  });
});

describe("bucketEventsByDay", () => {
  it("places a timed event in the correct day's bucket", () => {
    const gridDays = [new Date(2026, 2, 25), new Date(2026, 2, 26), new Date(2026, 2, 27)];
    const evt = makeEvent({ startTime: "2026-03-26T10:00:00Z", endTime: "2026-03-26T11:00:00Z" });
    const map = bucketEventsByDay(gridDays, [evt], "UTC");
    expect(map.get("2026-03-26")!.timed).toHaveLength(1);
    expect(map.get("2026-03-25")!.timed).toHaveLength(0);
    expect(map.get("2026-03-27")!.timed).toHaveLength(0);
  });

  it("sorts timed events within a day by start time ascending", () => {
    const gridDays = [new Date(2026, 2, 26)];
    const later = makeEvent({ id: "later", startTime: "2026-03-26T14:00:00Z", endTime: "2026-03-26T15:00:00Z" });
    const earlier = makeEvent({ id: "earlier", startTime: "2026-03-26T09:00:00Z", endTime: "2026-03-26T10:00:00Z" });
    const map = bucketEventsByDay(gridDays, [later, earlier], "UTC");
    expect(map.get("2026-03-26")!.timed.map((e) => e.id)).toEqual(["earlier", "later"]);
  });

  it("spans a multi-day all-day event across every covered day", () => {
    const gridDays = [new Date(2026, 3, 1), new Date(2026, 3, 2), new Date(2026, 3, 3), new Date(2026, 3, 4)];
    const evt = makeEvent({ allDay: true, startTime: "2026-04-01", endTime: "2026-04-03" });
    const map = bucketEventsByDay(gridDays, [evt], "UTC");
    expect(map.get("2026-04-01")!.allDay).toHaveLength(1);
    expect(map.get("2026-04-02")!.allDay).toHaveLength(1);
    expect(map.get("2026-04-03")!.allDay).toHaveLength(1);
    expect(map.get("2026-04-04")!.allDay).toHaveLength(0);
  });

  it("returns an entry for every grid day", () => {
    const gridDays = getMonthGridDays(new Date(2026, 7, 1));
    const map = bucketEventsByDay(gridDays, [], "UTC");
    expect(map.size).toBe(gridDays.length);
  });
});

describe("getEventsForDay timezone correctness", () => {
  it("buckets a UTC-timestamp timed event to the correct household-tz day", () => {
    // 03:00 UTC on Aug 14 is 23:00 EDT on Aug 13 in America/New_York.
    const evt = makeEvent({ startTime: "2026-08-14T03:00:00Z", endTime: "2026-08-14T04:00:00Z" });
    const aug13 = new Date(2026, 7, 13);
    const aug14 = new Date(2026, 7, 14);
    expect(getEventsForDay([evt], aug13, "America/New_York").timed).toHaveLength(1);
    expect(getEventsForDay([evt], aug14, "America/New_York").timed).toHaveLength(0);
  });

  it("buckets correctly on the spring-forward DST day (America/New_York 2026-03-08)", () => {
    // 06:00 UTC = 01:00 EST on Mar 8 (before the 02:00 spring-forward).
    const evt = makeEvent({ startTime: "2026-03-08T06:00:00Z", endTime: "2026-03-08T06:30:00Z" });
    const mar8 = new Date(2026, 2, 8);
    expect(getEventsForDay([evt], mar8, "America/New_York").timed).toHaveLength(1);
  });
});
