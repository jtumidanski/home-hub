import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { WorkoutShell } from "../workout-shell";

describe("WorkoutShell", () => {
  it("renders a page heading and a labelled section nav with the active tab current", () => {
    render(
      <MemoryRouter initialEntries={["/app/workouts/week"]}>
        <Routes>
          <Route path="/app/workouts" element={<WorkoutShell />}>
            <Route path="week" element={<p>week body</p>} />
          </Route>
        </Routes>
      </MemoryRouter>,
    );
    expect(screen.getByRole("heading", { level: 1, name: "Workouts" })).toBeInTheDocument();
    const nav = screen.getByRole("navigation", { name: "Workout sections" });
    expect(nav).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Week" })).toHaveAttribute("aria-current", "page");
    expect(screen.getByRole("link", { name: "Today" })).not.toHaveAttribute("aria-current");
    expect(screen.getByText("week body")).toBeInTheDocument();
  });
});
