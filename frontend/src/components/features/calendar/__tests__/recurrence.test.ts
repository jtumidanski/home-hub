import { describe, it, expect } from "vitest";
import {
  composeRecurrenceRule,
  eventStartInstant,
  formatUntilUTC,
} from "@/lib/calendar/recurrence";

describe("eventStartInstant", () => {
  it("returns the local-midnight Date for an all-day event", () => {
    const d = eventStartInstant("2026-05-06", "10:30", true, "America/New_York");
    expect(d.getFullYear()).toBe(2026);
    expect(d.getMonth()).toBe(4);
    expect(d.getDate()).toBe(6);
  });

  it("returns the parsed local Date for a timed event", () => {
    const d = eventStartInstant("2026-05-06", "10:30", false, "America/New_York");
    expect(d.getFullYear()).toBe(2026);
    expect(d.getMonth()).toBe(4);
    expect(d.getDate()).toBe(6);
    expect(d.getHours()).toBe(10);
    expect(d.getMinutes()).toBe(30);
  });
});

describe("formatUntilUTC", () => {
  it("produces YYYYMMDDTHHMMSSZ for end-of-day in EST (winter)", () => {
    // 2026-01-15 23:59:59 America/New_York is UTC-5 → 2026-01-16 04:59:59Z
    expect(formatUntilUTC("2026-01-15", "America/New_York")).toBe("20260116T045959Z");
  });

  it("produces the correct value in EDT (summer DST)", () => {
    // 2026-07-15 23:59:59 America/New_York is UTC-4 → 2026-07-16 03:59:59Z
    expect(formatUntilUTC("2026-07-15", "America/New_York")).toBe("20260716T035959Z");
  });

  it("produces an unchanged value for UTC", () => {
    expect(formatUntilUTC("2026-06-10", "UTC")).toBe("20260610T235959Z");
  });
});

describe("composeRecurrenceRule", () => {
  const presets = [
    "RRULE:FREQ=DAILY",
    "RRULE:FREQ=WEEKLY",
    "RRULE:FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR",
    "RRULE:FREQ=MONTHLY",
    "RRULE:FREQ=YEARLY",
  ];

  it("returns undefined when preset is empty", () => {
    expect(
      composeRecurrenceRule("", "on", "2026-06-10", 10, "2026-05-06", "09:00", "UTC"),
    ).toBeUndefined();
  });

  it("appends UNTIL for mode=on, across all presets", () => {
    for (const preset of presets) {
      const out = composeRecurrenceRule(preset, "on", "2026-06-10", 10, "2026-05-06", "09:00", "UTC");
      expect(out).toEqual([`${preset};UNTIL=20260610T235959Z`]);
    }
  });

  it("appends COUNT for mode=after", () => {
    const out = composeRecurrenceRule(
      "RRULE:FREQ=DAILY", "after", "", 5, "2026-05-06", "09:00", "America/New_York",
    );
    expect(out).toEqual(["RRULE:FREQ=DAILY;COUNT=5"]);
  });

  it("returns the preset unchanged for mode=never", () => {
    const out = composeRecurrenceRule(
      "RRULE:FREQ=WEEKLY", "never", "", 0, "2026-05-06", "09:00", "America/New_York",
    );
    expect(out).toEqual(["RRULE:FREQ=WEEKLY"]);
  });

  it("converts UNTIL to UTC for non-UTC zones (winter)", () => {
    const out = composeRecurrenceRule(
      "RRULE:FREQ=WEEKLY", "on", "2026-01-15", 0, "2026-01-01", "09:00", "America/New_York",
    );
    expect(out).toEqual(["RRULE:FREQ=WEEKLY;UNTIL=20260116T045959Z"]);
  });
});
