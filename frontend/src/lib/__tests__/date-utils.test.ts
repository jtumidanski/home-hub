import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import {
  getLocalToday,
  getLocalTodayStr,
  getLocalTodayRange,
  getLocalWeekStart,
  getLocalMonth,
} from "../date-utils";

describe("date-utils", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  describe("getLocalTodayStr", () => {
    it("returns the correct date for a UTC-5 timezone (west of UTC)", () => {
      // 2026-04-09 03:00 UTC = 2026-04-08 22:00 in America/New_York (EDT, UTC-4)
      vi.setSystemTime(new Date("2026-04-09T03:00:00Z"));
      expect(getLocalTodayStr("America/New_York")).toBe("2026-04-08");
    });

    it("returns the correct date for a UTC+9 timezone (east of UTC)", () => {
      // 2026-04-08 20:00 UTC = 2026-04-09 05:00 in Asia/Tokyo (UTC+9)
      vi.setSystemTime(new Date("2026-04-08T20:00:00Z"));
      expect(getLocalTodayStr("Asia/Tokyo")).toBe("2026-04-09");
    });

    it("returns the correct date for UTC+13 timezone", () => {
      // 2026-04-08 14:00 UTC = 2026-04-09 03:00 in Pacific/Auckland (NZDT, UTC+12 or +13)
      vi.setSystemTime(new Date("2026-04-08T14:00:00Z"));
      expect(getLocalTodayStr("Pacific/Auckland")).toBe("2026-04-09");
    });

    it("handles midnight edge case", () => {
      // Exactly midnight in UTC = still previous day in US timezones
      vi.setSystemTime(new Date("2026-04-09T00:00:00Z"));
      expect(getLocalTodayStr("America/Chicago")).toBe("2026-04-08");
    });

    it("handles DST spring-forward boundary (US)", () => {
      // 2026 US spring forward: March 8, 2026 at 2:00 AM EST -> 3:00 AM EDT
      // At 11:30 PM EST on March 7, UTC is March 8 04:30
      vi.setSystemTime(new Date("2026-03-08T04:30:00Z"));
      expect(getLocalTodayStr("America/New_York")).toBe("2026-03-07");
    });

    it("handles DST fall-back boundary (US)", () => {
      // 2026 US fall back: November 1, 2026 at 2:00 AM EDT -> 1:00 AM EST
      // At 11:30 PM EDT on Oct 31, UTC is Nov 1 03:30
      vi.setSystemTime(new Date("2026-11-01T03:30:00Z"));
      expect(getLocalTodayStr("America/New_York")).toBe("2026-10-31");
    });
  });

  describe("getLocalToday", () => {
    it("returns a Date at midnight for the correct local day", () => {
      vi.setSystemTime(new Date("2026-04-09T03:00:00Z"));
      const today = getLocalToday("America/New_York");
      expect(today.getFullYear()).toBe(2026);
      expect(today.getMonth()).toBe(3); // April = 3
      expect(today.getDate()).toBe(8);
      expect(today.getHours()).toBe(0);
      expect(today.getMinutes()).toBe(0);
    });
  });

  describe("getLocalTodayRange", () => {
    it("returns ISO strings spanning the correct local day", () => {
      vi.setSystemTime(new Date("2026-04-09T03:00:00Z"));
      const { start, end } = getLocalTodayRange("America/New_York");
      // In New York it's April 8. The ISO strings are UTC-converted, so
      // start of April 8 local = April 8 04:00 UTC (EDT offset),
      // end of April 8 local = April 9 03:59 UTC.
      const startDate = new Date(start);
      const endDate = new Date(end);
      expect(startDate.getFullYear()).toBe(2026);
      expect(startDate.getMonth()).toBe(3); // April
      expect(startDate.getDate()).toBe(8);
      expect(startDate.getHours()).toBe(0);
      // end should be 23:59:59 on the same local day
      expect(endDate.getFullYear()).toBe(2026);
      expect(endDate.getMonth()).toBe(3);
      expect(endDate.getDate()).toBe(8);
      expect(endDate.getHours()).toBe(23);
    });
  });

  describe("getLocalWeekStart", () => {
    it("returns Monday for a Wednesday", () => {
      // 2026-04-08 is a Wednesday
      vi.setSystemTime(new Date("2026-04-08T12:00:00Z"));
      const monday = getLocalWeekStart("UTC");
      expect(monday.getFullYear()).toBe(2026);
      expect(monday.getMonth()).toBe(3);
      expect(monday.getDate()).toBe(6); // Monday April 6
    });

    it("returns Monday for a Sunday", () => {
      // 2026-04-12 is a Sunday
      vi.setSystemTime(new Date("2026-04-12T12:00:00Z"));
      const monday = getLocalWeekStart("UTC");
      expect(monday.getDate()).toBe(6); // Previous Monday
    });

    it("returns Monday for a Monday", () => {
      // 2026-04-06 is a Monday
      vi.setSystemTime(new Date("2026-04-06T12:00:00Z"));
      const monday = getLocalWeekStart("UTC");
      expect(monday.getDate()).toBe(6);
    });

    it("respects timezone when day differs", () => {
      // 2026-04-13 01:00 UTC = 2026-04-12 (Sunday) in New York
      vi.setSystemTime(new Date("2026-04-13T01:00:00Z"));
      const monday = getLocalWeekStart("America/New_York");
      // Sunday April 12 -> Monday April 6
      expect(monday.getDate()).toBe(6);
    });
  });

  describe("getLocalMonth", () => {
    it("returns correct month string", () => {
      vi.setSystemTime(new Date("2026-04-08T12:00:00Z"));
      expect(getLocalMonth("UTC")).toBe("2026-04");
    });

    it("handles month boundary with timezone", () => {
      // 2026-05-01 03:00 UTC = 2026-04-30 in New York
      vi.setSystemTime(new Date("2026-05-01T03:00:00Z"));
      expect(getLocalMonth("America/New_York")).toBe("2026-04");
    });

    it("pads single-digit months", () => {
      vi.setSystemTime(new Date("2026-01-15T12:00:00Z"));
      expect(getLocalMonth("UTC")).toBe("2026-01");
    });
  });
});
