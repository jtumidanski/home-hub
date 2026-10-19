import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { EventPopover } from "../event-popover";
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

describe("EventPopover", () => {
  it("renders as a dialog named after the event title", () => {
    render(<EventPopover event={event} onClose={vi.fn()} />);
    expect(screen.getByRole("dialog", { name: "Dentist" })).toBeInTheDocument();
    expect(screen.getByText("Alex")).toBeInTheDocument();
  });

  it("closes on Escape", async () => {
    const onClose = vi.fn();
    render(<EventPopover event={event} onClose={onClose} />);
    await userEvent.keyboard("{Escape}");
    expect(onClose).toHaveBeenCalled();
  });

  it("closes via the named Close button", async () => {
    const onClose = vi.fn();
    render(<EventPopover event={event} onClose={onClose} />);
    await userEvent.click(screen.getByRole("button", { name: "Close" }));
    expect(onClose).toHaveBeenCalled();
  });
});
