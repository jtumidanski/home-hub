import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { CreateTaskDialog } from "../create-task-dialog";

const mockMutateAsync = vi.fn();

vi.mock("@/lib/hooks/api/use-tasks", () => ({
  useCreateTask: () => ({ mutateAsync: mockMutateAsync }),
}));

vi.mock("@/components/providers/auth-provider", () => ({
  useAuth: () => ({ user: { id: "user-1" } }),
}));

vi.mock("@/lib/hooks/api/use-household-members", () => ({
  useHouseholdMembers: () => ({ data: { data: [] } }),
}));

vi.mock("sonner", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

import { toast } from "sonner";

describe("CreateTaskDialog", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("does not render content when closed", () => {
    render(<CreateTaskDialog open={false} onOpenChange={vi.fn()} />);
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("renders form fields when open", () => {
    render(<CreateTaskDialog open={true} onOpenChange={vi.fn()} />);
    expect(screen.getByRole("dialog")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Create Task" })).toBeInTheDocument();
    expect(screen.getByPlaceholderText("Enter task title")).toBeInTheDocument();
    expect(screen.getByPlaceholderText("Optional notes")).toBeInTheDocument();
    expect(screen.getByText("Due Date")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /create task/i })).toBeInTheDocument();
  });

  it("shows validation error when submitting with empty title", async () => {
    const user = userEvent.setup();
    render(<CreateTaskDialog open={true} onOpenChange={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: /create task/i }));

    await waitFor(() => {
      expect(screen.getByText("Title is required")).toBeInTheDocument();
    });
    expect(mockMutateAsync).not.toHaveBeenCalled();
  });

  it("calls mutateAsync with correct data on valid submission", async () => {
    const user = userEvent.setup();
    mockMutateAsync.mockResolvedValue({});
    const onOpenChange = vi.fn();

    render(<CreateTaskDialog open={true} onOpenChange={onOpenChange} />);

    await user.type(screen.getByPlaceholderText("Enter task title"), "Buy groceries");
    await user.type(screen.getByPlaceholderText("Optional notes"), "Milk and eggs");
    await user.type(screen.getByLabelText("Due Date", { selector: "input" }), "2026-04-01");
    await user.click(screen.getByRole("button", { name: /create task/i }));

    await waitFor(() => {
      expect(mockMutateAsync).toHaveBeenCalledWith({
        title: "Buy groceries",
        notes: "Milk and eggs",
        dueOn: "2026-04-01",
        ownerUserId: "user-1",
      });
    });
  });

  it("shows toast.success and closes dialog on successful submission", async () => {
    const user = userEvent.setup();
    mockMutateAsync.mockResolvedValue({});
    const onOpenChange = vi.fn();

    render(<CreateTaskDialog open={true} onOpenChange={onOpenChange} />);

    await user.type(screen.getByPlaceholderText("Enter task title"), "Buy groceries");
    await user.click(screen.getByRole("button", { name: /create task/i }));

    await waitFor(() => {
      expect(toast.success).toHaveBeenCalledWith("Task created");
    });
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it("shows toast.error on mutation failure", async () => {
    const user = userEvent.setup();
    mockMutateAsync.mockRejectedValue(new Error("Network error"));
    const onOpenChange = vi.fn();

    render(<CreateTaskDialog open={true} onOpenChange={onOpenChange} />);

    await user.type(screen.getByPlaceholderText("Enter task title"), "Buy groceries");
    await user.click(screen.getByRole("button", { name: /create task/i }));

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalled();
    });
    expect(onOpenChange).not.toHaveBeenCalledWith(false);
  });

  it("resets form on successful submission", async () => {
    const user = userEvent.setup();
    mockMutateAsync.mockResolvedValue({});
    let isOpen = true;
    const onOpenChange = vi.fn((next: boolean) => {
      isOpen = next;
    });

    const { rerender } = render(
      <CreateTaskDialog open={isOpen} onOpenChange={onOpenChange} />
    );

    await user.type(screen.getByPlaceholderText("Enter task title"), "Some task");
    expect(screen.getByPlaceholderText("Enter task title")).toHaveValue("Some task");

    await user.click(screen.getByRole("button", { name: /create task/i }));

    await waitFor(() => {
      expect(onOpenChange).toHaveBeenCalledWith(false);
    });

    // Reopen the dialog after it was closed by submission
    rerender(<CreateTaskDialog open={true} onOpenChange={onOpenChange} />);

    await waitFor(() => {
      expect(screen.getByPlaceholderText("Enter task title")).toHaveValue("");
    });
  });
});
