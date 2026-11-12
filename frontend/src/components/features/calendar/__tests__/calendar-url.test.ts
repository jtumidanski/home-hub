import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { formatLocalDateStr } from "@/lib/date-utils";
import { calendarDateCodec, weekAnchorFor } from "../calendar-url";

const fmt = formatLocalDateStr;

describe("calendar-url", () => {
  beforeEach(() => {
    vi.useFakeTimers({ toFake: ["Date"] });
    vi.setSystemTime(new Date(2026, 10, 12, 12));
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  describe("weekAnchorFor", () => {
    it("snaps to Sunday on desktop", () => {
      expect(fmt(weekAnchorFor(new Date(2026, 10, 12), true))).toBe("2026-11-08");
    });
    it("uses the previous day on mobile", () => {
      expect(fmt(weekAnchorFor(new Date(2026, 10, 12), false))).toBe("2026-11-11");
    });
  });

  describe("calendarDateCodec", () => {
    it("month normalizes to the first of the month", () => {
      const c = calendarDateCodec("month", true);
      expect(fmt(c.parse("2026-03-17")!)).toBe("2026-03-01");
      expect(fmt(c.current())).toBe("2026-11-01");
    });
    it("week desktop normalizes to Sunday", () => {
      const c = calendarDateCodec("week", true);
      expect(fmt(c.parse("2026-03-18")!)).toBe("2026-03-15");
      expect(fmt(c.current())).toBe("2026-11-08");
    });
    it("week mobile keeps the date verbatim", () => {
      const c = calendarDateCodec("week", false);
      expect(fmt(c.parse("2026-03-18")!)).toBe("2026-03-18");
      expect(fmt(c.current())).toBe("2026-11-11");
    });
    it.each([
      ["month", true],
      ["week", true],
      ["week", false],
    ] as const)("%s (desktop=%s) rejects impossible dates and is a fixed point", (view, desktop) => {
      const c = calendarDateCodec(view, desktop);
      expect(c.parse("2026-02-31")).toBeNull();
      for (const d of [new Date(2026, 2, 18), new Date(2026, 10, 12), new Date(2025, 11, 31)]) {
        const x = c.parse(fmt(d))!; // canonical value for this view
        const s = c.serialize(x);
        expect(c.serialize(c.parse(s)!)).toBe(s);
      }
    });
  });
});
