import { render, screen, fireEvent } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { MemoryRouter, Route, Routes } from "react-router-dom";

import { LocationProbe } from "@/test/location-probe";
import { MealsPage } from "../MealsPage";

const usePlansMock = vi.fn();

vi.mock("@/lib/hooks/api/use-meals", () => {
  const mutation = () => ({ mutate: vi.fn(), mutateAsync: vi.fn(), isPending: false });
  return {
    usePlans: (...args: unknown[]) => {
      usePlansMock(...args);
      return { data: undefined, isLoading: false };
    },
    usePlan: () => ({ data: undefined, isLoading: false }),
    useCreatePlan: mutation,
    useLockPlan: mutation,
    useUnlockPlan: mutation,
    useDuplicatePlan: mutation,
    useAddPlanItem: mutation,
    useUpdatePlanItem: mutation,
    useRemovePlanItem: mutation,
  };
});
vi.mock("@/lib/hooks/api/use-shopping", () => ({
  useShoppingLists: () => ({ data: undefined, isLoading: false }),
  useImportMealPlan: () => ({ mutate: vi.fn(), mutateAsync: vi.fn(), isPending: false }),
  useCreateShoppingList: () => ({ mutate: vi.fn(), mutateAsync: vi.fn(), isPending: false }),
}));
vi.mock("@/context/tenant-context", () => ({ useTenant: () => ({ tenant: null, household: null }) }));
vi.mock("@/components/features/meals/week-grid", () => ({ WeekGrid: () => <div>week-grid</div> }));
vi.mock("@/components/features/meals/recipe-selector", () => ({ RecipeSelector: () => <div /> }));
vi.mock("@/components/features/meals/plan-item-popover", () => ({ PlanItemPopover: () => <div /> }));
vi.mock("@/components/features/meals/ingredient-preview", () => ({ IngredientPreview: () => <div /> }));
vi.mock("@/components/features/meals/export-modal", () => ({ ExportModal: () => <div /> }));
vi.mock("@/components/features/meals/week-selector", () => ({
  WeekSelector: ({ startsOn, onWeekChange }: { startsOn: Date; onWeekChange: (d: Date) => void }) => (
    <div>
      <div data-testid="selector-starts-on">{startsOn.toDateString()}</div>
      <button onClick={() => onWeekChange(new Date(2026, 10, 16))}>next-week</button>
      <button onClick={() => onWeekChange(new Date(2026, 10, 9))}>this-week</button>
    </div>
  ),
}));

function renderAt(url: string) {
  return render(
    <MemoryRouter initialEntries={[url]}>
      <Routes>
        <Route path="/app/meals" element={<MealsPage />} />
      </Routes>
      <LocationProbe />
    </MemoryRouter>,
  );
}

const loc = () => screen.getByTestId("location").textContent;
const navType = () => screen.getByTestId("nav-type").textContent;
const lastStartsOn = () => usePlansMock.mock.calls.at(-1)?.[0];

describe("MealsPage week param", () => {
  beforeEach(() => {
    usePlansMock.mockClear();
    vi.useFakeTimers({ toFake: ["Date"] });
    vi.setSystemTime(new Date(2026, 10, 12, 12));
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  it("defaults to the current week without changing the URL", () => {
    renderAt("/app/meals");
    expect(lastStartsOn()).toEqual({ starts_on: "2026-11-09" });
    expect(loc()).toBe("/app/meals");
  });

  it("reads a canonical Monday from the URL", () => {
    renderAt("/app/meals?week=2026-11-02");
    expect(lastStartsOn()).toEqual({ starts_on: "2026-11-02" });
    expect(loc()).toBe("/app/meals?week=2026-11-02");
  });

  it("normalizes a non-Monday to its Monday with a replace", () => {
    renderAt("/app/meals?week=2026-11-04");
    expect(lastStartsOn()).toEqual({ starts_on: "2026-11-02" });
    expect(loc()).toBe("/app/meals?week=2026-11-02");
    expect(navType()).toBe("REPLACE");
  });

  it.each(["2026-02-31", "2026-11-09"])("removes invalid or current-week param %s", (week) => {
    renderAt(`/app/meals?week=${week}`);
    expect(loc()).toBe("/app/meals");
    expect(lastStartsOn()).toEqual({ starts_on: "2026-11-09" });
  });

  it("selector changes replace the param and omit the current week", () => {
    renderAt("/app/meals");
    fireEvent.click(screen.getByText("next-week"));
    expect(loc()).toBe("/app/meals?week=2026-11-16");
    expect(navType()).toBe("REPLACE");
    expect(lastStartsOn()).toEqual({ starts_on: "2026-11-16" });
    fireEvent.click(screen.getByText("this-week"));
    expect(loc()).toBe("/app/meals");
  });
});
