import { describe, it, expect, vi, afterEach } from "vitest";
import { render, act } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { CalendarWidget } from "../calendar-widget";

const useCalendarEvents = vi.fn((start: string, end: string) => {
  void start;
  void end;
  return { data: { data: [] }, isLoading: false, isError: false };
});
vi.mock("@/lib/hooks/api/use-calendar", () => ({
  useCalendarEvents: (start: string, end: string) => useCalendarEvents(start, end),
}));
vi.mock("@/context/tenant-context", () => ({
  useTenant: () => ({ household: { attributes: { timezone: "America/New_York" } } }),
}));

describe("CalendarWidget day rollover", () => {
  afterEach(() => vi.useRealTimers());

  it("queries the new day after local midnight passes", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-03-06T04:58:30Z")); // 23:58 on Mar 5 in New York
    render(
      <MemoryRouter>
        <CalendarWidget />
      </MemoryRouter>,
    );
    const first = useCalendarEvents.mock.lastCall![0];
    expect(new Date(first).getDate()).toBe(5);
    act(() => {
      vi.advanceTimersByTime(180_000);
    });
    const later = useCalendarEvents.mock.lastCall![0];
    expect(new Date(later).getDate()).toBe(6);
  });
});
