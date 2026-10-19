import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";

const list = (names: string[]) => ({
  data: { data: names.map((name, i) => ({ id: String(i), attributes: { name } })) },
});

vi.mock("@/lib/hooks/api/use-workouts", () => ({
  useWorkoutThemes: () => list(["Strength"]),
  useCreateWorkoutTheme: () => ({ mutate: vi.fn() }),
  useDeleteWorkoutTheme: () => ({ mutate: vi.fn() }),
  useWorkoutRegions: () => list(["Upper body"]),
  useCreateWorkoutRegion: () => ({ mutate: vi.fn() }),
  useDeleteWorkoutRegion: () => ({ mutate: vi.fn() }),
}));

import { WorkoutTaxonomyPage } from "../WorkoutTaxonomyPage";

describe("WorkoutTaxonomyPage", () => {
  it("names every icon-only button", () => {
    render(<WorkoutTaxonomyPage />);
    expect(screen.getByRole("button", { name: "Add theme" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Add region" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Delete Strength" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Delete Upper body" })).toBeInTheDocument();
  });
});
