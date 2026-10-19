import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";

const mockToday = vi.fn();
vi.mock("@/lib/hooks/api/use-workouts", () => ({
  useWorkoutToday: () => mockToday(),
}));
vi.mock("@/context/tenant-context", () => ({
  useTenant: () => ({ household: { attributes: { timezone: "America/New_York" } } }),
}));
vi.mock("@/lib/hooks/use-local-date", () => ({
  useLocalDate: () => "2026-05-15",
}));

import { WorkoutWidget } from "../workout-widget";

function renderWidget() {
  return render(
    <MemoryRouter>
      <WorkoutWidget />
    </MemoryRouter>,
  );
}

beforeEach(() => mockToday.mockReset());

describe("WorkoutWidget", () => {
  it("shows the shared widget error with an alert role on failure", () => {
    mockToday.mockReturnValue({ data: undefined, isLoading: false, isError: true });
    renderWidget();
    expect(screen.getByRole("alert")).toHaveTextContent("Failed to load workout");
  });

  it("renders the title and today's exercises in the data state", () => {
    mockToday.mockReturnValue({
      data: {
        data: {
          attributes: {
            isRestDay: false,
            items: [{ id: "i1", exerciseName: "Squat", performance: { status: "done" } }],
          },
        },
      },
      isLoading: false,
      isError: false,
    });
    renderWidget();
    expect(screen.getByText("Workout")).toBeInTheDocument();
    expect(screen.getByText("Squat")).toBeInTheDocument();
  });

  it("is a whole-card link with a focus ring and background hover", () => {
    mockToday.mockReturnValue({
      data: { data: { attributes: { isRestDay: true, items: [] } } },
      isLoading: false,
      isError: false,
    });
    renderWidget();
    const link = screen.getByRole("link");
    expect(link).toHaveAttribute("href", "/app/workouts");
    expect(link.className).toContain("focus-visible:ring-2");
    expect(link.className).toContain("ring-ring");
    expect(link.innerHTML).toContain("group-hover/workout:bg-accent");
    expect(screen.getByText("Rest Day")).toBeInTheDocument();
  });
});
