import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { MemoryRouter, Routes, Route, Outlet } from "react-router-dom";
import { seedLayout } from "@/lib/dashboard/seed-layout";
import type { Dashboard } from "@/types/models/dashboard";

const navigateMock = vi.fn();
vi.mock("react-router-dom", async () => {
  const actual = await vi.importActual<typeof import("react-router-dom")>("react-router-dom");
  return {
    ...actual,
    useNavigate: () => navigateMock,
  };
});

const mutateMock = vi.fn();
vi.mock("@/lib/hooks/api/use-dashboards", () => ({
  useUpdateDashboard: () => ({
    mutate: mutateMock,
    isPending: false,
  }),
}));

// Stub widgets so the designer renders without needing a query client / tenant.
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

function ShellOutlet({ dashboard }: { dashboard: Dashboard }) {
  return <Outlet context={{ dashboard }} />;
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

beforeEach(() => {
  vi.clearAllMocks();
  if (!("ResizeObserver" in globalThis)) {
    (globalThis as unknown as { ResizeObserver: unknown }).ResizeObserver = class {
      observe() {}
      unobserve() {}
      disconnect() {}
    };
  }
});

describe("DashboardDesigner save/discard", () => {
  it("Save calls useUpdateDashboard with the draft name + layout", () => {
    const dashboard = makeDashboard();
    renderDesigner(dashboard);

    fireEvent.change(screen.getByLabelText("Dashboard name"), {
      target: { value: "Renamed" },
    });
    fireEvent.click(screen.getByTestId("designer-save"));

    expect(mutateMock).toHaveBeenCalledTimes(1);
    const call = mutateMock.mock.calls[0]![0] as {
      id: string;
      attrs: { name: string; layout: unknown };
    };
    expect(call.id).toBe("d-1");
    expect(call.attrs.name).toBe("Renamed");
    expect(call.attrs.layout).toEqual(dashboard.attributes.layout);
  });

  it("Discard on a clean draft navigates back immediately", () => {
    renderDesigner();
    fireEvent.click(screen.getByTestId("designer-discard"));
    expect(navigateMock).toHaveBeenCalledWith("..");
  });

  it("Discard on a dirty draft opens the confirmation dialog", () => {
    renderDesigner();
    fireEvent.change(screen.getByLabelText("Dashboard name"), {
      target: { value: "Changed" },
    });
    fireEvent.click(screen.getByTestId("designer-discard"));
    expect(screen.getByTestId("discard-confirm")).toBeInTheDocument();
    // Confirm the dialog's Discard button navigates back.
    fireEvent.click(screen.getByTestId("discard-confirm-button"));
    expect(navigateMock).toHaveBeenCalledWith("..");
  });
});
