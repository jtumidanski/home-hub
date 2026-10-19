import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import type { CalendarEvent } from "@/types/models/calendar";
import { MonthDayCell } from "../month-day-cell";
import { UserLegend } from "../user-legend";

function makeEvent(id: string, attrs: Partial<CalendarEvent["attributes"]>): CalendarEvent {
  return {
    id,
    type: "calendar-events",
    attributes: {
      title: "Event",
      description: null,
      startTime: "2026-10-19T10:00:00Z",
      endTime: "2026-10-19T11:00:00Z",
      allDay: false,
      location: null,
      visibility: "default",
      isOwner: true,
      userDisplayName: "Alex",
      userColor: "#4285F4",
      sourceId: "",
      connectionId: "",
      isRecurring: false,
      ...attrs,
    },
  };
}

function renderCell(timed: CalendarEvent[]) {
  render(
    <MonthDayCell
      day={new Date(2026, 9, 19)}
      bucket={{ allDay: [], timed }}
      inMonth
      isToday={false}
      isDesktop
      onDayClick={vi.fn()}
    />,
  );
}

describe("MonthDayCell label", () => {
  it("lists distinct owner names", () => {
    renderCell([
      makeEvent("1", { userDisplayName: "Alex" }),
      makeEvent("2", { userDisplayName: "Sam" }),
      makeEvent("3", { userDisplayName: "Alex" }),
    ]);
    expect(screen.getByRole("button")).toHaveAttribute("aria-label", "October 19, 3 events: Alex, Sam");
  });

  it("omits owner names for busy events", () => {
    renderCell([makeEvent("1", { title: "Busy", isOwner: false, userDisplayName: "Sam" })]);
    expect(screen.getByRole("button")).toHaveAttribute("aria-label", "October 19, 1 event");
  });
});

describe("UserLegend", () => {
  it("renders swatches with an edge ring", () => {
    const { container } = render(
      <UserLegend
        connections={[
          { id: "c1", attributes: { userDisplayName: "Alex", userColor: "#fff" } } as never,
        ]}
      />,
    );
    expect(container.querySelector(".ring-1")).not.toBeNull();
  });
});
