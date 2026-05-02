import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { TasksTomorrowAdapter } from "@/components/features/dashboard-widgets/tasks-tomorrow-adapter";

vi.mock("@/lib/hooks/api/use-tasks", () => ({ useTasks: vi.fn() }));
vi.mock("@/context/tenant-context", () => ({
  useTenant: () => ({ household: { attributes: { timezone: "UTC" } } }),
}));

import { useTasks } from "@/lib/hooks/api/use-tasks";

const task = (id: string, title: string, dueOn: string, status: "pending"|"completed" = "pending") => ({
  id, type: "tasks",
  attributes: { title, status, dueOn, rolloverEnabled: false, createdAt: "", updatedAt: "" },
});

describe("TasksTomorrowAdapter", () => {
  afterEach(() => vi.useRealTimers());

  it("renders incomplete tasks due tomorrow capped by limit", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-05-01T12:00:00Z"));
    (useTasks as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: [
        task("a", "Pay bill", "2026-05-02"),
        task("b", "Old", "2026-04-28"),
        task("c", "Done", "2026-05-02", "completed"),
      ] },
      isLoading: false, isError: false,
    });
    render(<MemoryRouter><TasksTomorrowAdapter config={{ limit: 5 }} /></MemoryRouter>);
    expect(screen.getByText("Pay bill")).toBeInTheDocument();
    expect(screen.queryByText("Old")).not.toBeInTheDocument();
    expect(screen.queryByText("Done")).not.toBeInTheDocument();
  });

  it("renders empty state", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-05-01T12:00:00Z"));
    (useTasks as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: [] }, isLoading: false, isError: false,
    });
    render(<MemoryRouter><TasksTomorrowAdapter config={{ limit: 5 }} /></MemoryRouter>);
    expect(screen.getByText(/No tasks for tomorrow/i)).toBeInTheDocument();
  });
});
