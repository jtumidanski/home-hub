import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { seedLayout } from "@/lib/dashboard/seed-layout";

// Keep widgets inert — the grid test only cares that the correct
// chrome/data-testids are emitted, not about widget rendering.
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

import { DesignerGrid } from "@/pages/dashboard-designer/designer-grid";

beforeEach(() => {
  if (!("ResizeObserver" in globalThis)) {
    (globalThis as unknown as { ResizeObserver: unknown }).ResizeObserver = class {
      observe() {}
      unobserve() {}
      disconnect() {}
    };
  }
});

describe("DesignerGrid", () => {
  it("renders one grid item + chrome per widget in the layout", () => {
    const layout = seedLayout();
    const dispatch = vi.fn();
    render(<DesignerGrid widgets={layout.widgets} dispatch={dispatch} />);

    for (const w of layout.widgets) {
      expect(screen.getByTestId(`grid-item-${w.id}`)).toBeInTheDocument();
      expect(screen.getByTestId(`widget-chrome-${w.id}`)).toBeInTheDocument();
    }
    expect(layout.widgets).toHaveLength(9);
  });
});
