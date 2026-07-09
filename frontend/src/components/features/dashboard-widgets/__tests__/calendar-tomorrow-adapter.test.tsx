import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { CalendarTomorrowAdapter } from "@/components/features/dashboard-widgets/calendar-tomorrow-adapter";

const tenant = vi.hoisted(() => ({ timezone: "UTC" }));

vi.mock("@/lib/hooks/api/use-calendar", () => ({ useCalendarEvents: vi.fn() }));
vi.mock("@/context/tenant-context", () => ({
  useTenant: () => ({ household: { attributes: { timezone: tenant.timezone } } }),
}));

import { useCalendarEvents } from "@/lib/hooks/api/use-calendar";

const event = (id: string, title: string, startTime: string, endTime: string, allDay = false) => ({
  id, type: "calendar-events",
  attributes: { title, startTime, endTime, allDay, userColor: "#000", userDisplayName: "Me" },
});

describe("CalendarTomorrowAdapter", () => {
  afterEach(() => {
    vi.useRealTimers();
    tenant.timezone = "UTC";
  });

  it("renders tomorrow's events sorted with all-day first", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-05-01T12:00:00Z"));
    (useCalendarEvents as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: [
        event("a", "Lunch",   "2026-05-02T17:00:00Z", "2026-05-02T18:00:00Z"),
        event("b", "Holiday", "2026-05-02T00:00:00Z", "2026-05-02T23:59:59Z", true),
      ] },
      isLoading: false, isError: false,
    });
    render(<MemoryRouter><CalendarTomorrowAdapter config={{ includeAllDay: true, limit: 5 }} /></MemoryRouter>);
    const items = screen.getAllByRole("listitem");
    expect(items[0]).toHaveTextContent("Holiday");
    expect(items[1]).toHaveTextContent("Lunch");
  });

  it("shows tomorrow's own all-day event, not the day-after's, in a west-of-UTC tz", () => {
    // Regression: all-day events are stored at UTC midnight. For a household
    // west of UTC, the old tz-shifted window dropped tomorrow's all-day event
    // and surfaced the day-after's. Today is Thu 2026-07-09; tomorrow is Fri
    // 2026-07-10 in New York.
    tenant.timezone = "America/New_York";
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-07-09T12:00:00Z"));
    (useCalendarEvents as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: [
        event("fri", "Friday Run",       "2026-07-10T00:00:00Z", "2026-07-10T00:00:00Z", true),
        event("sat", "dirty burg - 10k", "2026-07-11T00:00:00Z", "2026-07-11T00:00:00Z", true),
      ] },
      isLoading: false, isError: false,
    });
    render(<MemoryRouter><CalendarTomorrowAdapter config={{ includeAllDay: true, limit: 5 }} /></MemoryRouter>);
    expect(screen.getByText("Friday Run")).toBeInTheDocument();
    expect(screen.queryByText(/dirty burg/i)).not.toBeInTheDocument();
  });

  it("shows empty state", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-05-01T12:00:00Z"));
    (useCalendarEvents as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: [] }, isLoading: false, isError: false,
    });
    render(<MemoryRouter><CalendarTomorrowAdapter config={{ includeAllDay: true, limit: 5 }} /></MemoryRouter>);
    expect(screen.getByText(/No events tomorrow/i)).toBeInTheDocument();
  });
});
