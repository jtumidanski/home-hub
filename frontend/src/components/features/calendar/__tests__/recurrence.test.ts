import { describe, it, expect } from "vitest";
import { eventStartInstant } from "@/lib/calendar/recurrence";

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
