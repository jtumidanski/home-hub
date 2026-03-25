import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { TaskCard } from "../task-card";
import { type Task } from "@/types/models/task";

function makeTask(overrides: Partial<Task["attributes"]> = {}): Task {
  return {
    id: "task-1",
    type: "tasks",
    attributes: {
      title: "Buy groceries",
      status: "pending",
      rolloverEnabled: false,
      createdAt: "2026-03-01T00:00:00Z",
      updatedAt: "2026-03-01T00:00:00Z",
      ...overrides,
    },
  };
}

describe("TaskCard", () => {
  it("renders task title, status badge, and due date", () => {
    const task = makeTask({ dueOn: "2026-04-01" });
    render(<TaskCard task={task} onToggleComplete={vi.fn()} onDelete={vi.fn()} />);

    expect(screen.getByText("Buy groceries")).toBeInTheDocument();
    expect(screen.getByText("pending")).toBeInTheDocument();
    expect(screen.getByText("2026-04-01")).toBeInTheDocument();
  });

  it('renders "Complete" action for pending task', async () => {
    const user = userEvent.setup();
    const task = makeTask({ status: "pending" });
    render(<TaskCard task={task} onToggleComplete={vi.fn()} onDelete={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: "Actions" }));

    expect(screen.getByText("Complete")).toBeInTheDocument();
  });

  it('renders "Reopen" action for completed task', async () => {
    const user = userEvent.setup();
    const task = makeTask({ status: "completed" });
    render(<TaskCard task={task} onToggleComplete={vi.fn()} onDelete={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: "Actions" }));

    expect(screen.getByText("Reopen")).toBeInTheDocument();
  });

  it("calls onToggleComplete with task id and status when Complete is clicked", async () => {
    const user = userEvent.setup();
    const onToggleComplete = vi.fn();
    const task = makeTask({ status: "pending" });
    render(<TaskCard task={task} onToggleComplete={onToggleComplete} onDelete={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: "Actions" }));
    await user.click(screen.getByText("Complete"));

    expect(onToggleComplete).toHaveBeenCalledWith("task-1", "pending");
  });

  it("calls onDelete with task id when Delete is clicked", async () => {
    const user = userEvent.setup();
    const onDelete = vi.fn();
    const task = makeTask();
    render(<TaskCard task={task} onToggleComplete={vi.fn()} onDelete={onDelete} />);

    await user.click(screen.getByRole("button", { name: "Actions" }));
    await user.click(screen.getByText("Delete"));

    expect(onDelete).toHaveBeenCalledWith("task-1");
  });

  it("does not render due date when not provided", () => {
    const task = makeTask();
    render(<TaskCard task={task} onToggleComplete={vi.fn()} onDelete={vi.fn()} />);

    expect(screen.queryByText(/\d{4}-\d{2}-\d{2}/)).not.toBeInTheDocument();
  });
});
