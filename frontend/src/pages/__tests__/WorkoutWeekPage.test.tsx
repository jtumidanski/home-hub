import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";

const reorderMutate = vi.fn();
const patchMutate = vi.fn();
const weekState = vi.fn();

vi.mock("@/lib/hooks/api/use-workouts", () => ({
  useWorkoutWeek: () => weekState(),
  useCopyWorkoutWeek: () => ({ mutate: vi.fn() }),
  useAddPlannedItem: () => ({ mutate: vi.fn() }),
  useDeletePlannedItem: () => ({ mutate: vi.fn() }),
  useReorderPlannedItems: () => ({ mutate: reorderMutate }),
  useWorkoutExercises: () => ({ data: { data: [] } }),
  useWorkoutThemes: () => ({ data: { data: [] } }),
  useWorkoutRegions: () => ({ data: { data: [] } }),
  usePatchWorkoutWeek: () => ({ mutate: patchMutate }),
  useUpdatePlannedItem: () => ({ mutate: vi.fn() }),
}));

import { WorkoutWeekPage } from "../WorkoutWeekPage";

function item(id: string, name: string, dayOfWeek: number, position: number) {
  return {
    id,
    dayOfWeek,
    position,
    exerciseId: `e-${id}`,
    exerciseName: name,
    exerciseDeleted: false,
    kind: "strength",
    weightType: "free",
    planned: { sets: 3, reps: 5, weight: 100, weightUnit: "lb" },
  };
}

function setWeek(items: unknown[], restDayFlags: number[] = []) {
  weekState.mockReturnValue({
    data: { data: { id: "w", type: "weeks", attributes: { items, restDayFlags } } },
    error: null,
  });
}

function renderPage() {
  return render(
    <MemoryRouter>
      <WorkoutWeekPage />
    </MemoryRouter>,
  );
}

beforeEach(() => {
  reorderMutate.mockReset();
  patchMutate.mockReset();
});

describe("WorkoutWeekPage", () => {
  it("exposes the Rest toggle with aria-pressed", () => {
    setWeek([], [2]);
    renderPage();
    expect(screen.getByRole("button", { name: "Rest day, Wed" })).toHaveAttribute("aria-pressed", "true");
    const mon = screen.getByRole("button", { name: "Rest day, Mon" });
    expect(mon).toHaveAttribute("aria-pressed", "false");
    fireEvent.click(mon);
    expect(patchMutate).toHaveBeenCalledWith({ weekStart: expect.any(String), restDayFlags: [2, 0] }, expect.anything());
  });

  it("uses the visible summary text as the edit button's accessible name and contains no <p>", () => {
    setWeek([item("a", "Squat", 0, 0)]);
    renderPage();
    const edit = screen.getByRole("button", { name: /^Squat/ });
    expect(edit).toHaveAttribute("aria-expanded", "false");
    expect(edit.querySelector("p")).toBeNull();
    expect(edit).not.toHaveAttribute("aria-label");
    expect(edit).toHaveTextContent("3×5 @ 100 lb");
  });

  it("labels the inline edit inputs", () => {
    setWeek([item("a", "Squat", 0, 0)]);
    renderPage();
    fireEvent.click(screen.getByRole("button", { name: /^Squat/ }));
    expect(screen.getByLabelText("Sets")).toBeInstanceOf(HTMLInputElement);
    expect(screen.getByLabelText("Reps")).toBeInstanceOf(HTMLInputElement);
    expect(screen.getByLabelText("Weight")).toBeInstanceOf(HTMLInputElement);
  });

  it("moves an item to another day via the Move to day control, same reorder call as a drop", async () => {
    const user = userEvent.setup();
    setWeek([item("a", "Squat", 0, 0), item("b", "Row", 0, 1), item("c", "Press", 2, 0)]);
    renderPage();
    await user.click(screen.getByRole("combobox", { name: "Move Squat to day" }));
    const listbox = await screen.findByRole("listbox");
    // Current day is not offered.
    expect(within(listbox).queryByRole("option", { name: "Mon" })).toBeNull();
    await user.click(within(listbox).getByRole("option", { name: "Wed" }));
    expect(reorderMutate).toHaveBeenCalledTimes(1);
    expect(reorderMutate.mock.calls[0]![0]).toEqual({
      weekStart: expect.any(String),
      items: [
        { itemId: "b", dayOfWeek: 0, position: 0 },
        { itemId: "c", dayOfWeek: 2, position: 0 },
        { itemId: "a", dayOfWeek: 2, position: 1 },
      ],
    });
  });
});
