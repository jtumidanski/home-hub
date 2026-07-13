import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { WeekGrid } from "@/components/features/meals/week-grid";
import type { PlanItemAttributes } from "@/types/models/meal-plan";

// Local-midnight May 1 2026; getDaysOfWeek formats with the same local
// components, so the item's `day` string lines up regardless of TZ.
const STARTS_ON = new Date(2026, 4, 1);

function makeItem(overrides: Partial<PlanItemAttributes> = {}): PlanItemAttributes {
  return {
    id: "i1",
    day: "2026-05-01",
    slot: "dinner",
    recipe_id: "r1",
    recipe_title: "Tacos",
    recipe_servings: 4,
    recipe_deleted: false,
    serving_multiplier: null,
    planned_servings: null,
    notes: null,
    position: 0,
    ...overrides,
  };
}

function renderGrid(item: PlanItemAttributes, locked: boolean) {
  const onItemClick = vi.fn();
  const onItemNavigate = vi.fn();
  render(
    <WeekGrid
      startsOn={STARTS_ON}
      items={[item]}
      locked={locked}
      onCellClick={vi.fn()}
      onItemClick={onItemClick}
      onItemNavigate={onItemNavigate}
      onRemoveItem={vi.fn()}
    />,
  );
  return { onItemClick, onItemNavigate };
}

describe("WeekGrid meal click behavior", () => {
  it("navigates to the recipe when a locked plan's live meal is clicked", () => {
    const item = makeItem();
    const { onItemClick, onItemNavigate } = renderGrid(item, true);

    fireEvent.click(screen.getByText("Tacos"));

    expect(onItemNavigate).toHaveBeenCalledWith(item);
    expect(onItemClick).not.toHaveBeenCalled();
  });

  it("does not navigate a locked meal whose recipe was deleted", () => {
    const item = makeItem({ recipe_deleted: true });
    const { onItemClick, onItemNavigate } = renderGrid(item, true);

    fireEvent.click(screen.getByText("Tacos"));

    expect(onItemNavigate).not.toHaveBeenCalled();
    expect(onItemClick).not.toHaveBeenCalled();
  });

  it("opens the edit popover (not navigation) when the plan is unlocked", () => {
    const item = makeItem();
    const { onItemClick, onItemNavigate } = renderGrid(item, false);

    fireEvent.click(screen.getByText("Tacos"));

    expect(onItemClick).toHaveBeenCalledWith(item);
    expect(onItemNavigate).not.toHaveBeenCalled();
  });
});
