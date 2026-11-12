import { render, screen, fireEvent } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { MemoryRouter, Route, Routes } from "react-router-dom";

import { LocationProbe } from "@/test/location-probe";

// CalendarPage reads matchMedia at module scope, so this must run before import.
vi.hoisted(() => {
  window.matchMedia = ((query: string) => ({
    matches: true,
    media: query,
    onchange: null,
    addEventListener: () => {},
    removeEventListener: () => {},
    addListener: () => {},
    removeListener: () => {},
    dispatchEvent: () => false,
  })) as unknown as typeof window.matchMedia;
});

const { toast } = vi.hoisted(() => ({ toast: { success: vi.fn(), error: vi.fn() } }));
vi.mock("sonner", () => ({ toast }));

vi.mock("@/lib/hooks/api/use-calendar", () => ({
  useCalendarConnections: () => ({
    isLoading: false,
    isError: false,
    data: {
      data: [{ id: "c1", attributes: { writeAccess: true, status: "connected", userDisplayName: "A" } }],
    },
  }),
  useCalendarEvents: () => ({ isLoading: false, data: { data: [] } }),
  useCalendarSources: () => ({ data: { data: [] } }),
  useDeleteEvent: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));
vi.mock("@/lib/hooks/api/use-packages", () => ({ usePackages: () => ({ data: { data: [] } }) }));

vi.mock("@/components/features/calendar/calendar-grid", () => ({
  CalendarGrid: ({ weekStart }: { weekStart: Date }) => (
    <div data-testid="week-start">{weekStart.toDateString()}</div>
  ),
}));
vi.mock("@/components/features/calendar/month-grid", () => ({
  MonthGrid: ({ monthAnchor, onDayClick }: { monthAnchor: Date; onDayClick: (d: Date) => void }) => (
    <div>
      <div data-testid="month-anchor">{monthAnchor.toDateString()}</div>
      <button onClick={() => onDayClick(new Date(2026, 2, 18))}>click-day</button>
    </div>
  ),
}));
vi.mock("@/components/features/calendar/event-form-dialog", () => ({ EventFormDialog: () => null }));
vi.mock("@/components/features/calendar/recurring-scope-dialog", () => ({ RecurringScopeDialog: () => null }));
vi.mock("@/components/features/calendar/calendar-selection-panel", () => ({ CalendarSelectionPanel: () => null }));
vi.mock("@/components/features/calendar/connection-status", () => ({ ConnectionStatus: () => null }));
vi.mock("@/components/features/calendar/reauthorize-banner", () => ({ ReauthorizeBanner: () => null }));
vi.mock("@/components/features/calendar/connect-calendar-button", () => ({ ConnectCalendarButton: () => null }));

import { CalendarPage } from "../CalendarPage";

function renderAt(url: string) {
  return render(
    <MemoryRouter initialEntries={[url]}>
      <Routes>
        <Route path="/app/calendar/week" element={<CalendarPage view="week" />} />
        <Route path="/app/calendar/month" element={<CalendarPage view="month" />} />
      </Routes>
      <LocationProbe />
    </MemoryRouter>,
  );
}

const loc = () => screen.getByTestId("location").textContent;
const navType = () => screen.getByTestId("nav-type").textContent;
const click = (name: string | RegExp) => fireEvent.click(screen.getByRole("button", { name }));

describe("CalendarPage", () => {
  beforeEach(() => {
    toast.success.mockClear();
    toast.error.mockClear();
    vi.useFakeTimers({ toFake: ["Date"] });
    vi.setSystemTime(new Date(2026, 10, 12, 12));
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  it("shows the current desktop week and replaces on next/prev", () => {
    renderAt("/app/calendar/week");
    expect(screen.getByTestId("week-start").textContent).toBe(new Date(2026, 10, 8).toDateString());
    click("Next period");
    expect(loc()).toBe("/app/calendar/week?date=2026-11-15");
    expect(navType()).toBe("REPLACE");
    click("Previous period");
    expect(loc()).toBe("/app/calendar/week");
  });

  it("normalizes a month date with one replace", () => {
    renderAt("/app/calendar/month?date=2026-03-17");
    expect(screen.getByTestId("month-anchor").textContent).toBe(new Date(2026, 2, 1).toDateString());
    expect(loc()).toBe("/app/calendar/month?date=2026-03-01");
    expect(navType()).toBe("REPLACE");
  });

  it("falls back to the current period for an invalid date", () => {
    renderAt("/app/calendar/week?date=bogus");
    expect(loc()).toBe("/app/calendar/week");
    expect(screen.getByTestId("week-start").textContent).toBe(new Date(2026, 10, 8).toDateString());
  });

  it("pushes month when toggling from week", () => {
    renderAt("/app/calendar/week?date=2026-03-15");
    click("Month");
    expect(loc()).toBe("/app/calendar/month?date=2026-03-01");
    expect(navType()).toBe("PUSH");
  });

  it("toggles month to week on week 1 of a non-current month", () => {
    renderAt("/app/calendar/month?date=2026-03-01");
    click("Week");
    expect(loc()).toBe("/app/calendar/week?date=2026-03-01");
  });

  it("toggles month to the current week, omitting date", () => {
    renderAt("/app/calendar/month");
    click("Week");
    expect(loc()).toBe("/app/calendar/week");
  });

  it("day click pushes the week and Back returns to the month", () => {
    renderAt("/app/calendar/month?date=2026-03-01");
    click("click-day");
    expect(loc()).toBe("/app/calendar/week?date=2026-03-15");
    expect(navType()).toBe("PUSH");
    click("probe-back");
    expect(loc()).toBe("/app/calendar/month?date=2026-03-01");
  });

  it("consumes connected=true without dropping date", () => {
    renderAt("/app/calendar/week?connected=true");
    expect(toast.success).toHaveBeenCalledTimes(1);
    expect(loc()).toBe("/app/calendar/week");
  });

  it("keeps date when clearing connected", () => {
    renderAt("/app/calendar/week?connected=true&date=2026-03-15");
    expect(toast.success).toHaveBeenCalledTimes(1);
    expect(loc()).toBe("/app/calendar/week?date=2026-03-15");
  });

  it("shows an error toast for auth_failed", () => {
    renderAt("/app/calendar/week?error=auth_failed");
    expect(toast.error).toHaveBeenCalledWith("Google Calendar authorization failed. Please try again.");
    expect(loc()).toBe("/app/calendar/week");
  });
});
