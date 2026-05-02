import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { MealPlanTodayDetail } from "@/components/features/meals/meal-plan-today-detail";

vi.mock("@/lib/hooks/api/use-meals", () => ({
  usePlans: vi.fn(),
  usePlan: vi.fn(),
}));
vi.mock("@/context/tenant-context", () => ({
  useTenant: () => ({ household: { attributes: { timezone: "UTC" } } }),
}));

import { usePlans, usePlan } from "@/lib/hooks/api/use-meals";

const mockPlan = (items: Array<{ id: string; day: string; slot: string; recipe_id: string; recipe_title: string }>) => {
  (usePlans as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
    data: { data: [{ id: "p1" }] },
    isLoading: false, isError: false,
  });
  (usePlan as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
    data: { data: { attributes: { items } } },
    isLoading: false, isError: false,
  });
};

describe("MealPlanTodayDetail", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-05-01T12:00:00Z"));
  });
  afterEach(() => vi.useRealTimers());

  it("renders today's full B/L/D + N follow-up days of dinners", () => {
    mockPlan([
      { id: "1", day: "2026-05-01", slot: "breakfast", recipe_id: "r1", recipe_title: "Toast" },
      { id: "2", day: "2026-05-01", slot: "lunch",     recipe_id: "r2", recipe_title: "Salad" },
      { id: "3", day: "2026-05-01", slot: "dinner",    recipe_id: "r3", recipe_title: "Tacos" },
      { id: "4", day: "2026-05-02", slot: "dinner",    recipe_id: "r4", recipe_title: "Pasta" },
      { id: "5", day: "2026-05-02", slot: "lunch",     recipe_id: "r5", recipe_title: "SkipThis" },
      { id: "6", day: "2026-05-03", slot: "dinner",    recipe_id: "r6", recipe_title: "Stir fry" },
      { id: "7", day: "2026-05-04", slot: "dinner",    recipe_id: "r7", recipe_title: "Beyond horizon" },
    ]);
    render(<MemoryRouter><MealPlanTodayDetail horizonDays={3} /></MemoryRouter>);
    expect(screen.getByText("Toast")).toBeInTheDocument();
    expect(screen.getByText("Salad")).toBeInTheDocument();
    expect(screen.getByText("Tacos")).toBeInTheDocument();
    expect(screen.getByText("Pasta")).toBeInTheDocument();
    expect(screen.getByText("Stir fry")).toBeInTheDocument();
    expect(screen.queryByText("SkipThis")).not.toBeInTheDocument();
    expect(screen.queryByText("Beyond horizon")).not.toBeInTheDocument();
  });

  it("collapses to today-only when horizonDays is 1", () => {
    mockPlan([
      { id: "1", day: "2026-05-01", slot: "dinner", recipe_id: "r1", recipe_title: "Tacos" },
      { id: "2", day: "2026-05-02", slot: "dinner", recipe_id: "r2", recipe_title: "Pasta" },
    ]);
    render(<MemoryRouter><MealPlanTodayDetail horizonDays={1} /></MemoryRouter>);
    expect(screen.getByText("Tacos")).toBeInTheDocument();
    expect(screen.queryByText("Pasta")).not.toBeInTheDocument();
  });
});
