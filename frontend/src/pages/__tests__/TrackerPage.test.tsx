import { render, screen, fireEvent } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { MemoryRouter, Route, Routes } from "react-router-dom";

import { LocationProbe } from "@/test/location-probe";
import { TrackerPage } from "../TrackerPage";

vi.mock("@/components/features/tracker/today-view", () => ({ TodayView: () => <div>today-view</div> }));
vi.mock("@/components/features/tracker/tracker-setup", () => ({ TrackerSetup: () => <div>setup-view</div> }));
vi.mock("@/components/features/tracker/month-report", () => ({
  MonthReport: ({ month }: { month: string }) => <div>report-view:{month}</div>,
}));
vi.mock("@/components/features/tracker/calendar-grid", () => ({
  CalendarGrid: ({
    month,
    onMonthChange,
    onViewReport,
    onOpenSetup,
  }: {
    month: string;
    onMonthChange: (m: string) => void;
    onViewReport: () => void;
    onOpenSetup: () => void;
  }) => (
    <div>
      <div data-testid="grid-month">{month}</div>
      <button onClick={() => onMonthChange("2026-01")}>go-jan</button>
      <button onClick={() => onMonthChange("2026-11")}>go-nov</button>
      <button onClick={onViewReport}>view-report</button>
      <button onClick={onOpenSetup}>open-setup-from-calendar</button>
    </div>
  ),
}));

function renderAt(url: string) {
  return render(
    <MemoryRouter initialEntries={[url]}>
      <Routes>
        <Route path="/app/habits/today" element={<TrackerPage view="today" />} />
        <Route path="/app/habits/calendar" element={<TrackerPage view="calendar" />} />
        <Route path="/app/habits/setup" element={<TrackerPage view="setup" />} />
        <Route path="/app/habits/report" element={<TrackerPage view="report" />} />
      </Routes>
      <LocationProbe />
    </MemoryRouter>,
  );
}

const loc = () => screen.getByTestId("location").textContent;
const navType = () => screen.getByTestId("nav-type").textContent;
const pressed = (name: RegExp) => screen.getByRole("button", { name }).getAttribute("aria-pressed");

describe("TrackerPage", () => {
  beforeEach(() => {
    vi.useFakeTimers({ toFake: ["Date"] });
    vi.setSystemTime(new Date(2026, 10, 12, 12));
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  it("renders the Habits page heading", () => {
    renderAt("/app/habits/today");
    expect(screen.getByRole("heading", { level: 1, name: "Habits" })).toBeInTheDocument();
  });

  it.each([
    ["today", ["true", "false", "false"]],
    ["calendar", ["false", "true", "false"]],
    ["setup", ["false", "false", "true"]],
    ["report", ["false", "false", "false"]],
  ])("aria-pressed matches the %s view", (view, expected) => {
    renderAt(`/app/habits/${view}`);
    expect([pressed(/Today/), pressed(/Calendar/), pressed(/Setup/)]).toEqual(expected);
  });

  it("pushes the view, keeps month, and goes back", () => {
    renderAt("/app/habits/today?month=2026-01");
    fireEvent.click(screen.getByRole("button", { name: /Calendar/ }));
    expect(loc()).toBe("/app/habits/calendar?month=2026-01");
    expect(navType()).toBe("PUSH");
    fireEvent.click(screen.getByText("probe-back"));
    expect(loc()).toBe("/app/habits/today?month=2026-01");
  });

  it("shows the current month, and month changes replace and omit the current month", () => {
    renderAt("/app/habits/calendar");
    expect(screen.getByTestId("grid-month").textContent).toBe("2026-11");
    fireEvent.click(screen.getByText("go-jan"));
    expect(loc()).toBe("/app/habits/calendar?month=2026-01");
    expect(navType()).toBe("REPLACE");
    fireEvent.click(screen.getByText("go-nov"));
    expect(loc()).toBe("/app/habits/calendar");
  });

  it("falls back to the current month for an invalid month and removes it", () => {
    renderAt("/app/habits/calendar?month=2026-13");
    expect(screen.getByTestId("grid-month").textContent).toBe("2026-11");
    expect(loc()).toBe("/app/habits/calendar");
  });

  it("opens the report with the month", () => {
    renderAt("/app/habits/calendar?month=2026-01");
    fireEvent.click(screen.getByText("view-report"));
    expect(loc()).toBe("/app/habits/report?month=2026-01");
    expect(navType()).toBe("PUSH");
    expect(screen.getByText("report-view:2026-01")).toBeInTheDocument();
  });

  it("lets the calendar jump to Setup", () => {
    renderAt("/app/habits/calendar");
    fireEvent.click(screen.getByText("open-setup-from-calendar"));
    expect(loc()).toBe("/app/habits/setup");
    expect(navType()).toBe("PUSH");
    expect(screen.getByText("setup-view")).toBeInTheDocument();
  });

  it.each([
    ["calendar", "max-w-none"],
    ["report", "max-w-6xl"],
    ["today", "max-w-3xl"],
    ["setup", "max-w-3xl"],
  ])("uses the width class for %s", (view, cls) => {
    const { container } = renderAt(`/app/habits/${view}`);
    expect(container.querySelector(`.mx-auto.${cls}`)).not.toBeNull();
  });
});
