import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { MealPlanWidget } from "@/components/features/meals/meal-plan-widget";

vi.mock("@/lib/hooks/api/use-meals", () => ({
  usePlans: vi.fn(),
  usePlan: vi.fn(),
}));
vi.mock("@/context/tenant-context", () => ({
  useTenant: () => ({ household: { attributes: { timezone: "UTC" } } }),
}));

import { usePlans, usePlan } from "@/lib/hooks/api/use-meals";

describe("MealPlanWidget", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-05-01T12:00:00Z"));
    (usePlans as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: [{ id: "p1" }] },
      isLoading: false,
      isError: false,
    });
    (usePlan as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      data: {
        data: {
          attributes: {
            items: [
              { id: "1", day: "2026-05-01", slot: "dinner", recipe_id: "r3", recipe_title: "Tacos" },
            ],
          },
        },
      },
      isLoading: false,
      isError: false,
    });
  });
  afterEach(() => vi.useRealTimers());

  it("renders sibling links with no link nested inside another link", () => {
    const { container } = render(
      <MemoryRouter>
        <MealPlanWidget />
      </MemoryRouter>,
    );
    expect(screen.getByRole("link", { name: "Open meals" })).toHaveAttribute("href", "/app/meals");
    expect(screen.getByRole("link", { name: "Tacos" })).toHaveAttribute("href", "/app/recipes/r3");
    expect(container.querySelectorAll("a a")).toHaveLength(0);
  });
});
