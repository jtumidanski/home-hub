import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { CalendarEvent } from "@/types/models/calendar";
import { MonthGrid } from "../month-grid";

vi.mock("@/context/tenant-context", () => ({
  useTenant: () => ({ household: { attributes: { timezone: "UTC" } } }),
}));

function makeEvent(overrides: Partial<CalendarEvent["attributes"]> & { id?: string } = {}): CalendarEvent {
  const { id = "evt-1", ...attrs } = overrides;
  return {
    id,
    type: "calendar-events",
    attributes: {
      title: "Test Event",
      description: null,
      startTime: "2026-08-14T10:00:00Z",
      endTime: "2026-08-14T11:00:00Z",
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

afterEach(() => {
  vi.useRealTimers();
});

describe("MonthGrid", () => {
  it("renders seven weekday headers", () => {
    render(<MonthGrid monthAnchor={new Date(2026, 7, 1)} events={[]} isDesktop onDayClick={vi.fn()} />);
    for (const label of ["SUN", "MON", "TUE", "WED", "THU", "FRI", "SAT"]) {
      expect(screen.getByText(label)).toBeInTheDocument();
    }
  });

  it("renders 42 day cells for August 2026 (6-row month)", () => {
    render(<MonthGrid monthAnchor={new Date(2026, 7, 1)} events={[]} isDesktop onDayClick={vi.fn()} />);
    // Each cell is a button with an aria-label; weekday headers are not buttons.
    expect(screen.getAllByRole("button")).toHaveLength(42);
  });

  it("renders 35 day cells for March 2026 (5-row month)", () => {
    render(<MonthGrid monthAnchor={new Date(2026, 2, 1)} events={[]} isDesktop onDayClick={vi.fn()} />);
    expect(screen.getAllByRole("button")).toHaveLength(35);
  });

  it("highlights today's cell", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-08-14T12:00:00Z"));
    render(<MonthGrid monthAnchor={new Date(2026, 7, 1)} events={[]} isDesktop onDayClick={vi.fn()} />);
    const todayCell = screen.getByRole("button", { name: /August 14, / });
    expect(todayCell.className).toContain("bg-primary/10");
  });

  it("renders a timed event as a chip with a time prefix on desktop", () => {
    const evt = makeEvent({ startTime: "2026-08-14T09:00:00Z", endTime: "2026-08-14T10:00:00Z", title: "Standup" });
    render(<MonthGrid monthAnchor={new Date(2026, 7, 1)} events={[evt]} isDesktop onDayClick={vi.fn()} />);
    expect(screen.getByText("Standup")).toBeInTheDocument();
    expect(screen.getByText("9a")).toBeInTheDocument();
  });

  it("renders all events of a dense day in the DOM (no truncation)", () => {
    const events = Array.from({ length: 8 }, (_, i) =>
      makeEvent({
        id: `e${i}`,
        title: `Event ${i}`,
        startTime: `2026-08-14T0${i}:00:00Z`,
        endTime: `2026-08-14T0${i}:30:00Z`,
      }),
    );
    render(<MonthGrid monthAnchor={new Date(2026, 7, 1)} events={events} isDesktop onDayClick={vi.fn()} />);
    for (let i = 0; i < 8; i++) {
      expect(screen.getByText(`Event ${i}`)).toBeInTheDocument();
    }
  });

  it("renders dots (not chips) on mobile and caps at 4 with an overflow indicator", () => {
    const events = Array.from({ length: 6 }, (_, i) =>
      makeEvent({ id: `e${i}`, title: `Event ${i}`, startTime: `2026-08-14T0${i}:00:00Z`, endTime: `2026-08-14T0${i}:30:00Z` }),
    );
    render(<MonthGrid monthAnchor={new Date(2026, 7, 1)} events={events} isDesktop={false} onDayClick={vi.fn()} />);
    // Chips are not rendered in mobile mode.
    expect(screen.queryByText("Event 0")).not.toBeInTheDocument();
    // 6 events > 4 -> 3 colored dots + 1 overflow dot.
    expect(screen.getByTestId("dot-overflow")).toBeInTheDocument();
  });

  it("fires onDayClick with the cell's date for an in-month day", async () => {
    const onDayClick = vi.fn();
    render(<MonthGrid monthAnchor={new Date(2026, 7, 1)} events={[]} isDesktop onDayClick={onDayClick} />);
    await userEvent.click(screen.getByRole("button", { name: /August 14, / }));
    expect(onDayClick).toHaveBeenCalledTimes(1);
    const arg = onDayClick.mock.calls[0]![0] as Date;
    expect(arg.getMonth()).toBe(7);
    expect(arg.getDate()).toBe(14);
  });

  it("fires onDayClick for a trailing adjacent-month day", async () => {
    const onDayClick = vi.fn();
    render(<MonthGrid monthAnchor={new Date(2026, 7, 1)} events={[]} isDesktop onDayClick={onDayClick} />);
    // Sept 5, 2026 is the last (trailing) grid cell.
    await userEvent.click(screen.getByRole("button", { name: /September 5, / }));
    const arg = onDayClick.mock.calls[0]![0] as Date;
    expect(arg.getMonth()).toBe(8);
    expect(arg.getDate()).toBe(5);
  });
});
