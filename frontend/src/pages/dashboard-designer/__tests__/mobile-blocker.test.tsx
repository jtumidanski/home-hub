import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter, Routes, Route, Outlet } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { seedLayout } from "@/lib/dashboard/seed-layout";
import type { Dashboard } from "@/types/models/dashboard";

vi.mock("@/context/tenant-context", () => ({
  useTenant: () => ({ tenant: { id: "t1" }, household: { id: "h1" } }),
}));

// Stub widgets so the designer renders under desktop without extra providers.
vi.mock("@/components/features/weather/weather-widget", () => ({
  WeatherWidget: () => null,
}));
vi.mock("@/components/features/dashboard-widgets/tasks-summary", () => ({
  TasksSummaryWidget: () => null,
}));
vi.mock("@/components/features/dashboard-widgets/reminders-summary", () => ({
  RemindersSummaryWidget: () => null,
}));
vi.mock("@/components/features/dashboard-widgets/overdue-summary", () => ({
  OverdueSummaryWidget: () => null,
}));
vi.mock("@/components/features/meals/meal-plan-widget", () => ({
  MealPlanWidget: () => null,
}));
vi.mock("@/components/features/calendar/calendar-widget", () => ({
  CalendarWidget: () => null,
}));
vi.mock("@/components/features/packages/package-summary-widget", () => ({
  PackageSummaryWidget: () => null,
}));
vi.mock("@/components/features/trackers/habits-widget", () => ({
  HabitsWidget: () => null,
}));
vi.mock("@/components/features/workouts/workout-widget", () => ({
  WorkoutWidget: () => null,
}));

const mobileFlag = { value: false };
vi.mock("@/lib/hooks/use-mobile", () => ({
  useMobile: () => mobileFlag.value,
}));

import DashboardDesigner from "@/pages/DashboardDesigner";

function makeDashboard(): Dashboard {
  return {
    id: "d-1",
    type: "dashboards",
    attributes: {
      name: "Home",
      scope: "household",
      sortOrder: 0,
      schemaVersion: 1,
      createdAt: "2026-01-01T00:00:00Z",
      updatedAt: "2026-01-01T00:00:00Z",
      layout: seedLayout(),
    },
  };
}

function renderDesigner() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={["/edit"]}>
        <Routes>
          <Route element={<ShellOutlet dashboard={makeDashboard()} />}>
            <Route path="/edit" element={<DashboardDesigner />} />
          </Route>
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

function ShellOutlet({ dashboard }: { dashboard: Dashboard }) {
  return <Outlet context={{ dashboard }} />;
}

beforeEach(() => {
  if (!("ResizeObserver" in globalThis)) {
    (globalThis as unknown as { ResizeObserver: unknown }).ResizeObserver = class {
      observe() {}
      unobserve() {}
      disconnect() {}
    };
  }
});

describe("DashboardDesigner mobile blocker", () => {
  it("renders the blocker pane when useMobile is true", () => {
    mobileFlag.value = true;
    renderDesigner();
    expect(screen.getByTestId("designer-mobile-blocker")).toBeInTheDocument();
    expect(screen.getByText(/Editing needs a larger screen/)).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /view only/i })).toBeInTheDocument();
  });

  it("renders the full designer when useMobile is false", () => {
    mobileFlag.value = false;
    renderDesigner();
    expect(screen.queryByTestId("designer-mobile-blocker")).not.toBeInTheDocument();
    expect(screen.getByTestId("dashboard-designer")).toBeInTheDocument();
  });
});
