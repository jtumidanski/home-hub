import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { TasksTodayAdapter } from "@/components/features/dashboard-widgets/tasks-today-adapter";

vi.mock("@/lib/hooks/api/use-tasks", () => ({
  useTasks: vi.fn(),
}));
vi.mock("@/context/tenant-context", () => ({
  useTenant: () => ({ household: { attributes: { timezone: "UTC" } } }),
}));
vi.mock("@/lib/hooks/use-local-date", () => ({
  useLocalDate: () => "2026-05-01",
}));

import { useTasks } from "@/lib/hooks/api/use-tasks";

const renderAdapter = (config = { includeCompleted: true }) =>
  render(
    <MemoryRouter>
      <TasksTodayAdapter config={config} />
    </MemoryRouter>,
  );

const task = (
  id: string,
  title: string,
  status: "pending" | "completed",
  dueOn?: string,
  completedAt?: string,
) => ({
  id,
  type: "tasks",
  attributes: { title, status, dueOn, completedAt, rolloverEnabled: false, createdAt: "", updatedAt: "" },
});

describe("TasksTodayAdapter", () => {
  it("renders skeleton while loading", () => {
    (useTasks as unknown as ReturnType<typeof vi.fn>).mockReturnValue({ data: undefined, isLoading: true, isError: false });
    const { container } = renderAdapter();
    expect(container.querySelectorAll('[data-slot="skeleton"]').length).toBeGreaterThan(0);
  });

  it("renders overdue and today's incomplete tasks", () => {
    (useTasks as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: [
        task("a", "Pay bill", "pending", "2026-04-30"),
        task("b", "Walk dog", "pending", "2026-05-01"),
        task("c", "Old done", "completed", "2026-04-29", "2026-04-29T09:00:00Z"),
      ] },
      isLoading: false,
      isError: false,
    });
    renderAdapter();
    expect(screen.getByText("Pay bill")).toBeInTheDocument();
    expect(screen.getByText("Walk dog")).toBeInTheDocument();
    expect(screen.getByText(/Overdue \(1\)/)).toBeInTheDocument();
  });

  it("shows the all-completed fallback when only completed-today tasks exist", () => {
    (useTasks as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: [
        task("a", "Done early", "completed", "2026-05-01", "2026-05-01T08:00:00Z"),
      ] },
      isLoading: false,
      isError: false,
    });
    renderAdapter();
    expect(screen.getByText(/All tasks completed/i)).toBeInTheDocument();
  });

  it("shows empty copy when nothing is due", () => {
    (useTasks as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: [] },
      isLoading: false,
      isError: false,
    });
    renderAdapter();
    expect(screen.getByText(/No tasks for today/i)).toBeInTheDocument();
  });

  it("renders error banner without crashing", () => {
    (useTasks as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
    });
    renderAdapter();
    expect(screen.getByText(/Failed to load/i)).toBeInTheDocument();
  });
});
