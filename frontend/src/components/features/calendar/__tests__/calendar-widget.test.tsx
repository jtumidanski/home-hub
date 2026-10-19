import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { CalendarWidget } from "../calendar-widget";
import { AllDayEventRow } from "../all-day-event-row";
import type { CalendarEvent } from "@/types/models/calendar";

const event: CalendarEvent = {
  id: "evt-1",
  type: "calendar-events",
  attributes: {
    title: "Dentist",
    description: null,
    startTime: "2026-08-14T10:00:00Z",
    endTime: "2026-08-14T11:00:00Z",
    allDay: false,
    location: null,
    visibility: "default",
    userDisplayName: "Alex",
    userColor: "#4285F4",
    isOwner: true,
    sourceId: "",
    connectionId: "",
    isRecurring: false,
  },
};

vi.mock("@/lib/hooks/api/use-calendar", () => ({
  useCalendarEvents: () => ({ data: { data: [event] }, isLoading: false, isError: false }),
}));
vi.mock("@/context/tenant-context", () => ({
  useTenant: () => ({ household: null }),
}));

describe("CalendarWidget", () => {
  it("shows the event owner name as visible text", () => {
    render(
      <MemoryRouter>
        <CalendarWidget />
      </MemoryRouter>,
    );
    const owner = screen.getByText("Alex");
    expect(owner).toBeInTheDocument();
    expect(owner).not.toHaveClass("sr-only");
    expect(owner).toHaveClass("text-xs", "text-muted-foreground", "truncate");
  });
});

describe("AllDayEventRow", () => {
  it("gives the chip button a minimum touch-target height", () => {
    render(<AllDayEventRow events={[{ ...event, attributes: { ...event.attributes, allDay: true } }]} />);
    expect(screen.getByRole("button", { name: "Dentist" })).toHaveClass("min-h-6", "coarse:min-h-11");
  });
});
