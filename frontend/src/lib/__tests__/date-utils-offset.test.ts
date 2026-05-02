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
});
