import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";

const mockToday = vi.fn();
vi.mock("@/lib/hooks/api/use-workouts", () => ({
  useWorkoutToday: () => mockToday(),
  usePatchPerformance: () => ({ mutate: vi.fn() }),
}));
vi.mock("@/context/tenant-context", () => ({
  useTenant: () => ({ household: { attributes: { timezone: "UTC" } } }),
}));
vi.mock("@/lib/hooks/use-local-date", () => ({ useLocalDate: () => "2026-04-13" }));

import { WorkoutTodayPage } from "../WorkoutTodayPage";

function doc(items: unknown[]) {
  return {
    data: {
      id: "d",
      type: "workout-today",
      attributes: { date: "2026-04-13", weekStart: "2026-04-13", isRestDay: false, items },
    },
  };
}

describe("WorkoutTodayPage", () => {
  it("shows an empty state when nothing is planned", () => {
    mockToday.mockReturnValue({ data: doc([]), isLoading: false, error: null });
    render(<WorkoutTodayPage />);
    expect(screen.getByRole("heading", { name: "Nothing planned for today" })).toBeInTheDocument();
  });

  it("associates log form labels with their inputs", () => {
    mockToday.mockReturnValue({
      data: doc([
        {
          id: "i1",
          exerciseName: "Squat",
          kind: "strength",
          weightType: "free",
          planned: { sets: 3, reps: 5, weight: 100, weightUnit: "lb" },
          performance: null,
        },
      ]),
      isLoading: false,
      error: null,
    });
    render(<WorkoutTodayPage />);
    fireEvent.click(screen.getByText("Squat"));
    expect(screen.getByLabelText("Sets")).toBeInstanceOf(HTMLInputElement);
    expect(screen.getByLabelText("Reps")).toBeInstanceOf(HTMLInputElement);
    expect(screen.getByLabelText("Weight")).toBeInstanceOf(HTMLInputElement);
  });
});
