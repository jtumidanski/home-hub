import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { PerformanceStatus, SummaryDocument, WorkoutKind } from "@/types/models/workout";

const mockUseWorkoutWeekSummary = vi.fn();
const mockUseWorkoutNearestPopulatedWeek = vi.fn();

vi.mock("@/lib/hooks/api/use-workouts", () => ({
  useWorkoutWeekSummary: (weekStart: string) => mockUseWorkoutWeekSummary(weekStart),
  useWorkoutNearestPopulatedWeek: (reference: string, direction: string, enabled?: boolean) =>
    mockUseWorkoutNearestPopulatedWeek(reference, direction, enabled),
}));

import { WorkoutReviewPage } from "../WorkoutReviewPage";

function summaryDoc(overrides?: {
  weekStart?: string;
  kind?: WorkoutKind;
  status?: PerformanceStatus;
  planned?: Record<string, unknown>;
  actualSummary?: Record<string, unknown> | null;
  previousPopulatedWeek?: string | null;
  nextPopulatedWeek?: string | null;
  totalPlannedItems?: number;
  totalPerformedItems?: number;
  totalSkippedItems?: number;
}): SummaryDocument {
  const weekStart = overrides?.weekStart ?? "2026-04-13";
  return {
    data: {
      type: "week-summaries",
      id: weekStart,
      attributes: {
        weekStartDate: weekStart,
        restDayFlags: [6],
        totalPlannedItems: overrides?.totalPlannedItems ?? 3,
        totalPerformedItems: overrides?.totalPerformedItems ?? 1,
        totalSkippedItems: overrides?.totalSkippedItems ?? 1,
        previousPopulatedWeek:
          overrides && "previousPopulatedWeek" in overrides ? overrides.previousPopulatedWeek ?? null : "2026-04-06",
        nextPopulatedWeek:
          overrides && "nextPopulatedWeek" in overrides ? overrides.nextPopulatedWeek ?? null : null,
        byDay: [
          {
            dayOfWeek: 0,
            isRestDay: false,
            items: [
              {
                itemId: "item-1",
                exerciseName: "Bench Press",
                kind: overrides?.kind ?? "strength",
                status: overrides?.status ?? "done",
                planned: (overrides?.planned as any) ?? { sets: 3, reps: 10, weight: 135, weightUnit: "lb" },
                actualSummary: (overrides?.actualSummary as any) ?? { sets: 3, reps: 10, weight: 140, weightUnit: "lb" },
              },
            ],
          },
          { dayOfWeek: 1, isRestDay: false, items: [] },
          { dayOfWeek: 2, isRestDay: false, items: [] },
          { dayOfWeek: 3, isRestDay: false, items: [] },
          { dayOfWeek: 4, isRestDay: false, items: [] },
          { dayOfWeek: 5, isRestDay: false, items: [] },
          { dayOfWeek: 6, isRestDay: true, items: [] },
        ],
        byTheme: [],
        byRegion: [],
      },
    },
  };
}

function renderPage(weekStart = "2026-04-13") {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={[`/app/workouts/review/${weekStart}`]}>
        <Routes>
          <Route path="/app/workouts/review/:weekStart" element={<WorkoutReviewPage />} />
          <Route path="/app/workouts/review" element={<WorkoutReviewPage />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

beforeEach(() => {
  mockUseWorkoutWeekSummary.mockReset();
  mockUseWorkoutNearestPopulatedWeek.mockReset();
  mockUseWorkoutNearestPopulatedWeek.mockReturnValue({ data: null, isLoading: false, error: null });
});

describe("WorkoutReviewPage totals + per-day grid", () => {
  it("renders four stats including computed Pending", () => {
    mockUseWorkoutWeekSummary.mockReturnValue({
      data: summaryDoc({ totalPlannedItems: 5, totalPerformedItems: 2, totalSkippedItems: 1 }),
      isLoading: false,
      error: null,
    });
    renderPage();
    expect(screen.getByText("Planned")).toBeInTheDocument();
    expect(screen.getByText("Performed")).toBeInTheDocument();
    expect(screen.getByText("Pending")).toBeInTheDocument();
    expect(screen.getByText("Skipped")).toBeInTheDocument();
    // Pending = 5 - 2 - 1 = 2
    const pendingValue = screen.getByText("Pending").previousSibling;
    expect(pendingValue?.textContent).toBe("2");
  });

  it("renders all seven day sections with rest pill + item counts", () => {
    mockUseWorkoutWeekSummary.mockReturnValue({ data: summaryDoc(), isLoading: false, error: null });
    renderPage();
    for (const label of ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"]) {
      expect(screen.getByRole("region", { name: label })).toBeInTheDocument();
    }
    expect(screen.getByText("Rest")).toBeInTheDocument();
    expect(screen.getByText("1 exercise")).toBeInTheDocument();
    expect(screen.getAllByText("Nothing scheduled").length).toBeGreaterThan(0);
  });
});

describe("WorkoutReviewPage per-item rendering", () => {
  it("strength summary: planned and actual lines render, target-met check shown", () => {
    mockUseWorkoutWeekSummary.mockReturnValue({
      data: summaryDoc({
        status: "done",
        kind: "strength",
        planned: { sets: 3, reps: 10, weight: 135, weightUnit: "lb" },
        actualSummary: { sets: 3, reps: 10, weight: 140, weightUnit: "lb" },
      }),
      isLoading: false,
      error: null,
    });
    renderPage();
    expect(screen.getByText(/Planned:\s*3×10/)).toBeInTheDocument();
    expect(screen.getByText(/Actual:\s*3×10/)).toBeInTheDocument();
    expect(screen.getByLabelText("Target met")).toBeInTheDocument();
  });

  it("strength per-set: enumerates each set row", () => {
    mockUseWorkoutWeekSummary.mockReturnValue({
      data: summaryDoc({
        status: "done",
        kind: "strength",
        planned: { sets: 3, reps: 10, weight: 135, weightUnit: "lb" },
        actualSummary: {
          sets: 3, reps: 10, weight: 145, weightUnit: "lb",
          setRows: [
            { setNumber: 1, reps: 10, weight: 135 },
            { setNumber: 2, reps: 10, weight: 140 },
            { setNumber: 3, reps: 8, weight: 145 },
          ],
        },
      }),
      isLoading: false,
      error: null,
    });
    renderPage();
    expect(screen.getByText(/set 1: 10 @ 135/)).toBeInTheDocument();
    expect(screen.getByText(/set 3: 8 @ 145/)).toBeInTheDocument();
  });

  it("isometric: formats sets × duration", () => {
    mockUseWorkoutWeekSummary.mockReturnValue({
      data: summaryDoc({
        kind: "isometric",
        status: "partial",
        planned: { sets: 3, durationSeconds: 60 },
        actualSummary: { sets: 3, durationSeconds: 55 },
      }),
      isLoading: false,
      error: null,
    });
    renderPage();
    expect(screen.getByText(/Planned:\s*3×1:00/)).toBeInTheDocument();
    expect(screen.getByText(/Actual:\s*3×0:55/)).toBeInTheDocument();
  });

  it("cardio: renders duration and distance", () => {
    mockUseWorkoutWeekSummary.mockReturnValue({
      data: summaryDoc({
        kind: "cardio",
        status: "done",
        planned: { durationSeconds: 1800, distance: 3.0, distanceUnit: "mi" },
        actualSummary: { durationSeconds: 1720, distance: 3.1, distanceUnit: "mi" },
      }),
      isLoading: false,
      error: null,
    });
    renderPage();
    expect(screen.getByText(/Planned:\s*30:00 · 3 mi/)).toBeInTheDocument();
    expect(screen.getByText(/Actual:\s*28:40 · 3.1 mi/)).toBeInTheDocument();
    expect(screen.getByLabelText("Target met")).toBeInTheDocument();
  });

  it("skipped: strikes through name and renders Actual: Skipped", () => {
    mockUseWorkoutWeekSummary.mockReturnValue({
      data: summaryDoc({ status: "skipped", actualSummary: null }),
      isLoading: false,
      error: null,
    });
    renderPage();
    expect(screen.getByText("Actual: Skipped")).toBeInTheDocument();
    expect(screen.getByText("Bench Press").className).toContain("line-through");
  });

  it("pending: italic muted name and Actual: —", () => {
    mockUseWorkoutWeekSummary.mockReturnValue({
      data: summaryDoc({ status: "pending", actualSummary: null }),
      isLoading: false,
      error: null,
    });
    renderPage();
    expect(screen.getByText("Actual: —")).toBeInTheDocument();
    expect(screen.getByText("Bench Press").className).toContain("italic");
  });
});

describe("WorkoutReviewPage navigation header", () => {
  it("disables populated-jump buttons when no target available", () => {
    mockUseWorkoutWeekSummary.mockReturnValue({
      data: summaryDoc({ previousPopulatedWeek: null, nextPopulatedWeek: null }),
      isLoading: false,
      error: null,
    });
    renderPage();
    const prevJump = screen.getByRole("button", { name: /No earlier populated week/i });
    const nextJump = screen.getByRole("button", { name: /No later populated week/i });
    expect(prevJump).toBeDisabled();
    expect(nextJump).toBeDisabled();
  });

  it("next-week button pushes a URL entry", () => {
    mockUseWorkoutWeekSummary.mockReturnValue({ data: summaryDoc(), isLoading: false, error: null });
    renderPage("2026-04-13");

    // Just verify the button is wired and responds to clicks.
    const next = screen.getByRole("button", { name: "Next week" });
    fireEvent.click(next);
    // The route renders again for the next week — the page continues to
    // function (new summary hook call with the new weekStart).
    expect(mockUseWorkoutWeekSummary).toHaveBeenCalledWith("2026-04-20");
  });
});

describe("WorkoutReviewPage empty week", () => {
  it("renders friendly card when summary hook returns error", () => {
    mockUseWorkoutWeekSummary.mockReturnValue({ data: undefined, isLoading: false, error: new Error("404") });
    mockUseWorkoutNearestPopulatedWeek.mockImplementation((_ref, dir) => ({
      data: dir === "prev"
        ? { data: { type: "workoutWeekPointer", id: "2026-04-06", attributes: { weekStartDate: "2026-04-06" } } }
        : null,
      isLoading: false,
      error: null,
    }));
    renderPage();
    expect(screen.getByText("No workouts logged for this week.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Jump to previous populated week 2026-04-06/ })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /No later populated week/ })).toBeDisabled();
  });
});
