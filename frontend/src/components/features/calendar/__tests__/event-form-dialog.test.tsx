import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { EventFormDialog } from "../event-form-dialog";
import type { CalendarConnection, CalendarSource } from "@/types/models/calendar";

const mockCreate = vi.fn();
const mockUpdate = vi.fn();

vi.mock("@/lib/hooks/api/use-calendar", () => ({
  useCreateEvent: () => ({ mutateAsync: mockCreate }),
  useUpdateEvent: () => ({ mutateAsync: mockUpdate }),
}));

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

function makeConnection(): CalendarConnection {
  return {
    id: "conn-1",
    type: "calendar-connections",
    attributes: {
      provider: "google",
      status: "connected",
      email: "x@example.com",
      userDisplayName: "Tester",
      userColor: "#000",
      writeAccess: true,
      lastSyncAt: null,
      lastSyncAttemptAt: null,
      lastSyncEventCount: 0,
      errorCode: null,
      errorMessage: null,
      lastErrorAt: null,
      consecutiveFailures: 0,
      createdAt: "2026-01-01T00:00:00Z",
    },
  };
}

function makeSource(): CalendarSource {
  return {
    id: "src-1",
    type: "calendar-sources",
    attributes: {
      name: "Primary",
      primary: true,
      visible: true,
      color: "#000",
    },
  };
}

describe("EventFormDialog — Ends control", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockCreate.mockResolvedValue({});
  });

  it("hides the Ends control when 'Does not repeat' is selected", () => {
    render(
      <EventFormDialog
        open
        onOpenChange={vi.fn()}
        connections={[makeConnection()]}
        sources={[makeSource()]}
      />,
    );
    expect(screen.queryByLabelText("End date", { selector: "input[aria-label='End date']" })).not.toBeInTheDocument();
  });

  it("shows the Ends control and seeds end date to start + 1y when a recurring preset is chosen", async () => {
    const user = userEvent.setup();
    render(
      <EventFormDialog
        open
        onOpenChange={vi.fn()}
        connections={[makeConnection()]}
        sources={[makeSource()]}
        prefilledStart={new Date("2026-05-06T09:00:00")}
      />,
    );
    await user.selectOptions(screen.getByLabelText("Repeats"), "RRULE:FREQ=WEEKLY");
    const endInput = await screen.findByLabelText("End date", { selector: "input[aria-label='End date']" });
    expect((endInput as HTMLInputElement).value).toBe("2027-05-06");
  });

  it("auto-updates the end date when start date changes (untouched end-date)", async () => {
    const user = userEvent.setup();
    render(
      <EventFormDialog
        open
        onOpenChange={vi.fn()}
        connections={[makeConnection()]}
        sources={[makeSource()]}
        prefilledStart={new Date("2026-05-06T09:00:00")}
      />,
    );
    await user.selectOptions(screen.getByLabelText("Repeats"), "RRULE:FREQ=WEEKLY");
    const startInput = screen.getByLabelText("Start date");
    await user.clear(startInput);
    await user.type(startInput, "2026-06-10");
    await waitFor(() => {
      expect((screen.getByLabelText("End date", { selector: "input[aria-label='End date']" }) as HTMLInputElement).value).toBe("2027-06-10");
    });
  });

  it("leaves a user-edited end date alone when start date changes later", async () => {
    const user = userEvent.setup();
    render(
      <EventFormDialog
        open
        onOpenChange={vi.fn()}
        connections={[makeConnection()]}
        sources={[makeSource()]}
        prefilledStart={new Date("2026-05-06T09:00:00")}
      />,
    );
    await user.selectOptions(screen.getByLabelText("Repeats"), "RRULE:FREQ=WEEKLY");
    const endInput = screen.getByLabelText("End date", { selector: "input[aria-label='End date']" });
    await user.clear(endInput);
    await user.type(endInput, "2026-08-01");

    const startInput = screen.getByLabelText("Start date");
    await user.clear(startInput);
    await user.type(startInput, "2026-06-10");
    expect((screen.getByLabelText("End date", { selector: "input[aria-label='End date']" }) as HTMLInputElement).value).toBe("2026-08-01");
  });

  it("blocks submit until the Never confirmation checkbox is checked", async () => {
    const user = userEvent.setup();
    render(
      <EventFormDialog
        open
        onOpenChange={vi.fn()}
        connections={[makeConnection()]}
        sources={[makeSource()]}
        prefilledStart={new Date("2026-05-06T09:00:00")}
      />,
    );
    await user.type(screen.getByPlaceholderText("Event title"), "Standup");
    await user.selectOptions(screen.getByLabelText("Repeats"), "RRULE:FREQ=WEEKLY");
    await user.click(screen.getByLabelText(/^never$/i));
    await user.click(screen.getByRole("button", { name: /create event/i }));
    expect(mockCreate).not.toHaveBeenCalled();
    await user.click(screen.getByLabelText(/I understand this event has no end date/i));
    await user.click(screen.getByRole("button", { name: /create event/i }));
    await waitFor(() => expect(mockCreate).toHaveBeenCalledTimes(1));
    expect(mockCreate.mock.calls[0]![0].data.recurrence).toEqual(["RRULE:FREQ=WEEKLY"]);
  });

  it("submits an UNTIL-terminated RRULE for mode=on", async () => {
    const user = userEvent.setup();
    render(
      <EventFormDialog
        open
        onOpenChange={vi.fn()}
        connections={[makeConnection()]}
        sources={[makeSource()]}
        prefilledStart={new Date("2026-05-06T09:00:00")}
      />,
    );
    await user.type(screen.getByPlaceholderText("Event title"), "Volleyball");
    await user.selectOptions(screen.getByLabelText("Repeats"), "RRULE:FREQ=WEEKLY");
    const endInput = screen.getByLabelText("End date", { selector: "input[aria-label='End date']" });
    await user.clear(endInput);
    await user.type(endInput, "2026-06-10");
    await user.click(screen.getByRole("button", { name: /create event/i }));
    await waitFor(() => expect(mockCreate).toHaveBeenCalledTimes(1));
    const rule = mockCreate.mock.calls[0]![0].data.recurrence[0] as string;
    expect(rule.startsWith("RRULE:FREQ=WEEKLY;UNTIL=20260611T")).toBe(true);
    expect(rule.endsWith("Z")).toBe(true);
  });

  it("submits a COUNT-terminated RRULE for mode=after", async () => {
    const user = userEvent.setup();
    render(
      <EventFormDialog
        open
        onOpenChange={vi.fn()}
        connections={[makeConnection()]}
        sources={[makeSource()]}
        prefilledStart={new Date("2026-05-06T09:00:00")}
      />,
    );
    await user.type(screen.getByPlaceholderText("Event title"), "PT");
    await user.selectOptions(screen.getByLabelText("Repeats"), "RRULE:FREQ=DAILY");
    await user.click(screen.getByLabelText(/^after$/i));
    const countInput = screen.getByLabelText("Occurrences");
    await user.clear(countInput);
    await user.type(countInput, "5");
    await user.click(screen.getByRole("button", { name: /create event/i }));
    await waitFor(() => expect(mockCreate).toHaveBeenCalledTimes(1));
    expect(mockCreate.mock.calls[0]![0].data.recurrence).toEqual(["RRULE:FREQ=DAILY;COUNT=5"]);
  });

  it("resets Ends fields when switching back to 'Does not repeat'", async () => {
    const user = userEvent.setup();
    render(
      <EventFormDialog
        open
        onOpenChange={vi.fn()}
        connections={[makeConnection()]}
        sources={[makeSource()]}
        prefilledStart={new Date("2026-05-06T09:00:00")}
      />,
    );
    const repeats = screen.getByLabelText("Repeats");
    await user.selectOptions(repeats, "RRULE:FREQ=WEEKLY");
    expect(await screen.findByLabelText("End date", { selector: "input[aria-label='End date']" })).toBeInTheDocument();
    await user.selectOptions(repeats, "");
    expect(screen.queryByLabelText("End date", { selector: "input[aria-label='End date']" })).not.toBeInTheDocument();
  });
});
