import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { MemoryRouter, Routes, Route, Outlet } from "react-router-dom";
import { seedLayout } from "@/lib/dashboard/seed-layout";
import type { Dashboard } from "@/types/models/dashboard";

// Stub widget components — none of them should actually fetch during this
// high-level designer-render test.
vi.mock("@/components/features/weather/weather-widget", () => ({
  WeatherWidget: () => <div data-testid="stub-weather" />,
}));
vi.mock("@/components/features/dashboard-widgets/tasks-summary", () => ({
  TasksSummaryWidget: () => <div data-testid="stub-tasks" />,
}));
vi.mock("@/components/features/dashboard-widgets/reminders-summary", () => ({
  RemindersSummaryWidget: () => <div data-testid="stub-reminders" />,
}));
vi.mock("@/components/features/dashboard-widgets/overdue-summary", () => ({
  OverdueSummaryWidget: () => <div data-testid="stub-overdue" />,
}));
vi.mock("@/components/features/meals/meal-plan-widget", () => ({
  MealPlanWidget: () => <div data-testid="stub-meals" />,
}));
vi.mock("@/components/features/calendar/calendar-widget", () => ({
  CalendarWidget: () => <div data-testid="stub-calendar" />,
}));
vi.mock("@/components/features/packages/package-summary-widget", () => ({
  PackageSummaryWidget: () => <div data-testid="stub-packages" />,
}));
vi.mock("@/components/features/trackers/habits-widget", () => ({
  HabitsWidget: () => <div data-testid="stub-habits" />,
}));
vi.mock("@/components/features/workouts/workout-widget", () => ({
  WorkoutWidget: () => <div data-testid="stub-workouts" />,
}));

import DashboardDesigner from "@/pages/DashboardDesigner";

beforeEach(() => {
  if (!("ResizeObserver" in globalThis)) {
    (globalThis as unknown as { ResizeObserver: unknown }).ResizeObserver = class {
      observe() {}
      unobserve() {}
      disconnect() {}
    };
  }
});

function makeDashboard(): Dashboard {
  return {
    id: "d-1",
    type: "dashboards",
    attributes: {
      name: "My Home",
      scope: "household",
      sortOrder: 0,
      schemaVersion: 1,
      createdAt: "2026-01-01T00:00:00Z",
      updatedAt: "2026-01-01T00:00:00Z",
      layout: seedLayout(),
    },
  };
}

function renderDesigner(dashboard: Dashboard = makeDashboard()) {
  return render(
    <MemoryRouter initialEntries={["/edit"]}>
      <Routes>
        <Route element={<ShellOutlet dashboard={dashboard} />}>
          <Route path="/edit" element={<DashboardDesigner />} />
        </Route>
      </Routes>
    </MemoryRouter>,
  );
}

function ShellOutlet({ dashboard }: { dashboard: Dashboard }) {
  // Provides the `{ dashboard }` outlet context the designer expects.
  return <Outlet context={{ dashboard }} />;
}

describe("DashboardDesigner", () => {
  it("mounts with seeded dashboard and renders header + grid items", () => {
    renderDesigner();
    expect(screen.getByTestId("dashboard-designer")).toBeInTheDocument();
    expect(screen.getByLabelText("Dashboard name")).toHaveValue("My Home");
    expect(screen.getByTestId("designer-save")).toBeInTheDocument();
    expect(screen.getByTestId("designer-discard")).toBeInTheDocument();
    // All 9 seed widgets render grid items.
    for (let i = 0; i < 9; i++) {
      expect(
        screen.getAllByTestId(/^grid-item-/),
      ).toHaveLength(9);
      break;
    }
  });

  it("renaming marks the designer dirty", () => {
    renderDesigner();
    const input = screen.getByLabelText("Dashboard name") as HTMLInputElement;
    fireEvent.change(input, { target: { value: "Renamed" } });
    expect(input.value).toBe("Renamed");
    expect(screen.getByText("Unsaved changes")).toBeInTheDocument();
  });
});
