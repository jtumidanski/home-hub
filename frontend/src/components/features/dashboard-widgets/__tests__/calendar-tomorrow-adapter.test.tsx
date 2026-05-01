import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { CalendarTomorrowAdapter } from "@/components/features/dashboard-widgets/calendar-tomorrow-adapter";

vi.mock("@/lib/hooks/api/use-calendar", () => ({ useCalendarEvents: vi.fn() }));
vi.mock("@/context/tenant-context", () => ({
  useTenant: () => ({ household: { attributes: { timezone: "UTC" } } }),
}));

import { useCalendarEvents } from "@/lib/hooks/api/use-calendar";

const event = (id: string, title: string, startTime: string, endTime: string, allDay = false) => ({
  id, type: "calendar-events",
  attributes: { title, startTime, endTime, allDay, userColor: "#000", userDisplayName: "Me" },
});

describe("CalendarTomorrowAdapter", () => {
  afterEach(() => vi.useRealTimers());

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
