import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { MealPlanAdapter } from "@/components/features/dashboard-widgets/meal-plan-adapter";

vi.mock("@/components/features/meals/meal-plan-widget", () => ({
  MealPlanWidget: () => <div>LIST_VIEW</div>,
}));
vi.mock("@/components/features/meals/meal-plan-today-detail", () => ({
  MealPlanTodayDetail: () => <div>DETAIL_VIEW</div>,
}));

describe("MealPlanAdapter", () => {
  it("renders MealPlanWidget for view='list'", () => {
    render(<MemoryRouter><MealPlanAdapter config={{ horizonDays: 1, view: "list" }} /></MemoryRouter>);
    expect(screen.getByText("LIST_VIEW")).toBeInTheDocument();
  });

  it("renders MealPlanTodayDetail for view='today-detail'", () => {
    render(<MemoryRouter><MealPlanAdapter config={{ horizonDays: 3, view: "today-detail" }} /></MemoryRouter>);
    expect(screen.getByText("DETAIL_VIEW")).toBeInTheDocument();
  });

  it("absent view defaults to list", () => {
    render(<MemoryRouter><MealPlanAdapter config={{ horizonDays: 1 } as any} /></MemoryRouter>);
    expect(screen.getByText("LIST_VIEW")).toBeInTheDocument();
  });
});
