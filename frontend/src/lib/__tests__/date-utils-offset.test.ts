import { describe, it, expect, vi, afterEach } from "vitest";
import { getLocalDateStrOffset } from "@/lib/date-utils";

describe("getLocalDateStrOffset", () => {
  afterEach(() => {
    vi.useRealTimers();
  });

  it("returns the next-day date in the given timezone", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-05-01T12:00:00Z"));
    expect(getLocalDateStrOffset("America/New_York", 1)).toBe("2026-05-02");
  });

  it("offset 0 matches today's date in the same timezone", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-05-01T12:00:00Z"));
    expect(getLocalDateStrOffset("UTC", 0)).toBe("2026-05-01");
  });

  it("crosses month boundaries correctly", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-01-31T12:00:00Z"));
    expect(getLocalDateStrOffset("UTC", 1)).toBe("2026-02-01");
  });

  it.each([
    ["Pacific/Auckland", "2026-05-02", "2026-05-03"],
    ["Pacific/Honolulu", "2026-05-01", "2026-05-02"],
  ])("%s: offset 0 is the local date and offset 1 the next", (tz, today, tomorrow) => {
    vi.useFakeTimers();
    // 2026-05-01T23:00Z is May 2 11:00 in Auckland (UTC+12) and May 1 13:00 in Honolulu (UTC-10).
    vi.setSystemTime(new Date("2026-05-01T23:00:00Z"));
    expect(getLocalDateStrOffset(tz, 0)).toBe(today);
    expect(getLocalDateStrOffset(tz, 1)).toBe(tomorrow);
  });
});
