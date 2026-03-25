import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

const mockUseTasks = vi.fn();
const mockUpdateMutateAsync = vi.fn();
const mockDeleteMutateAsync = vi.fn();

vi.mock("@/lib/hooks/api/use-tasks", () => ({
  useTasks: () => mockUseTasks(),
  useUpdateTask: () => ({ mutateAsync: mockUpdateMutateAsync }),
  useDeleteTask: () => ({ mutateAsync: mockDeleteMutateAsync }),
}));

vi.mock("@/lib/api/errors", () => ({
  createErrorFromUnknown: (_err: unknown, fallback: string) => ({ message: fallback, type: "unknown" }),
}));

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

vi.mock("@/components/features/tasks/create-task-dialog", () => ({
  CreateTaskDialog: ({ open }: { open: boolean }) =>
    open ? <div role="dialog">CreateTaskDialog</div> : null,
}));

import { TasksPage } from "../TasksPage";

describe("TasksPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseTasks.mockReturnValue({ data: null, isLoading: false, isError: false });
  });

  it("renders loading skeleton when isLoading is true", () => {
    mockUseTasks.mockReturnValue({ data: null, isLoading: true, isError: false });
    render(<TasksPage />);
    // Skeleton renders multiple placeholder elements but no heading
    expect(screen.queryByText("Tasks")).not.toBeInTheDocument();
    expect(screen.getByRole("status", { name: "Loading" })).toBeInTheDocument();
  });

  it("renders error state when isError is true", () => {
    mockUseTasks.mockReturnValue({ data: null, isLoading: false, isError: true });
    render(<TasksPage />);
    expect(screen.getByText(/failed to load tasks/i)).toBeInTheDocument();
  });

  it("renders empty state when there are no tasks", () => {
    mockUseTasks.mockReturnValue({ data: { data: [] }, isLoading: false, isError: false });
    render(<TasksPage />);
    expect(screen.getByText("Tasks")).toBeInTheDocument();
    expect(screen.getByText(/no tasks yet/i)).toBeInTheDocument();
    expect(screen.getByText("Create First Task")).toBeInTheDocument();
  });

  it("renders task list when tasks exist", () => {
    mockUseTasks.mockReturnValue({
      data: {
        data: [
          { id: "t-1", type: "tasks", attributes: { title: "Buy milk", status: "pending", dueOn: "2026-04-01" } },
          { id: "t-2", type: "tasks", attributes: { title: "Walk dog", status: "completed" } },
        ],
      },
      isLoading: false,
      isError: false,
    });
    render(<TasksPage />);
    expect(screen.getByText("Buy milk")).toBeInTheDocument();
    expect(screen.getByText("Walk dog")).toBeInTheDocument();
    expect(screen.getByText("Due: 2026-04-01")).toBeInTheDocument();
  });

  it("opens create dialog when New Task button is clicked", async () => {
    const user = userEvent.setup();
    mockUseTasks.mockReturnValue({ data: { data: [] }, isLoading: false, isError: false });
    render(<TasksPage />);

    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: /new task/i }));
    expect(screen.getByRole("dialog")).toBeInTheDocument();
  });
});
