import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { PaletteDrawer } from "@/pages/dashboard-designer/palette-drawer";

// Keep widget registry intact but stub the rendered components to avoid
// router/tenant dependencies being loaded indirectly.
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

describe("PaletteDrawer", () => {
  it("lists every registered widget when open", () => {
    render(<PaletteDrawer open onOpenChange={vi.fn()} dispatch={vi.fn()} />);
    // At minimum, weather + tasks-summary entries exist.
    expect(screen.getByTestId("palette-add-weather")).toBeInTheDocument();
    expect(screen.getByTestId("palette-add-tasks-summary")).toBeInTheDocument();
  });

  it("adds a widget and closes the drawer on click", () => {
    const dispatch = vi.fn();
    const onOpenChange = vi.fn();
    render(<PaletteDrawer open onOpenChange={onOpenChange} dispatch={dispatch} />);
    fireEvent.click(screen.getByTestId("palette-add-tasks-summary"));
    expect(dispatch).toHaveBeenCalledTimes(1);
    const call = dispatch.mock.calls[0][0];
    expect(call.type).toBe("add");
    expect(call.widget.type).toBe("tasks-summary");
    expect(call.widget.id).toBeTruthy();
    expect(call.widget.w).toBeGreaterThan(0);
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });
});
