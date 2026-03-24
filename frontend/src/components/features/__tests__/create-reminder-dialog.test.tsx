import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { CreateReminderDialog } from "../create-reminder-dialog";

const mockMutateAsync = vi.fn();

vi.mock("@/lib/hooks/api/use-reminders", () => ({
  useCreateReminder: () => ({ mutateAsync: mockMutateAsync }),
}));

vi.mock("sonner", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

import { toast } from "sonner";

describe("CreateReminderDialog", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("does not render content when closed", () => {
    render(<CreateReminderDialog open={false} onOpenChange={vi.fn()} />);
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("renders form fields when open", () => {
    render(<CreateReminderDialog open={true} onOpenChange={vi.fn()} />);
    expect(screen.getByRole("dialog")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Create Reminder" })).toBeInTheDocument();
    expect(screen.getByPlaceholderText("Enter reminder title")).toBeInTheDocument();
    expect(screen.getByPlaceholderText("Optional notes")).toBeInTheDocument();
    expect(screen.getByText("Scheduled For")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /create reminder/i })).toBeInTheDocument();
  });

  it("shows validation error when submitting empty title", async () => {
    const user = userEvent.setup();
    render(<CreateReminderDialog open={true} onOpenChange={vi.fn()} />);

    const form = screen.getByRole("dialog").querySelector("form")!;
    const scheduledInput = form.querySelector("input[name='scheduledFor']")!;
    await user.type(scheduledInput, "2026-04-01T10:00");

    await user.click(screen.getByRole("button", { name: /create reminder/i }));

    await waitFor(() => {
      expect(screen.getByText("Title is required")).toBeInTheDocument();
    });
    expect(mockMutateAsync).not.toHaveBeenCalled();
  });

  it("shows validation error when submitting empty scheduledFor", async () => {
    const user = userEvent.setup();
    render(<CreateReminderDialog open={true} onOpenChange={vi.fn()} />);

    await user.type(screen.getByPlaceholderText("Enter reminder title"), "Test Reminder");
    await user.click(screen.getByRole("button", { name: /create reminder/i }));

    await waitFor(() => {
      expect(screen.getByText("Scheduled time is required")).toBeInTheDocument();
    });
    expect(mockMutateAsync).not.toHaveBeenCalled();
  });

  it("calls mutateAsync with correct data on valid submission", async () => {
    const user = userEvent.setup();
    mockMutateAsync.mockResolvedValue({});
    render(<CreateReminderDialog open={true} onOpenChange={vi.fn()} />);

    await user.type(screen.getByPlaceholderText("Enter reminder title"), "Doctor appointment");
    await user.type(screen.getByPlaceholderText("Optional notes"), "Bring insurance card");
    const form = screen.getByRole("dialog").querySelector("form")!;
    const scheduledInput = form.querySelector("input[name='scheduledFor']")!;
    await user.type(scheduledInput, "2026-04-01T10:00");

    await user.click(screen.getByRole("button", { name: /create reminder/i }));

    await waitFor(() => {
      expect(mockMutateAsync).toHaveBeenCalledOnce();
    });

    const call = mockMutateAsync.mock.calls[0][0];
    expect(call.title).toBe("Doctor appointment");
    expect(call.notes).toBe("Bring insurance card");
    expect(call.scheduledFor).toBe(new Date("2026-04-01T10:00").toISOString());
  });

  it("shows toast.success and closes dialog on success", async () => {
    const user = userEvent.setup();
    mockMutateAsync.mockResolvedValue({});
    const onOpenChange = vi.fn();
    render(<CreateReminderDialog open={true} onOpenChange={onOpenChange} />);

    await user.type(screen.getByPlaceholderText("Enter reminder title"), "Buy groceries");
    const form = screen.getByRole("dialog").querySelector("form")!;
    const scheduledInput = form.querySelector("input[name='scheduledFor']")!;
    await user.type(scheduledInput, "2026-04-01T10:00");

    await user.click(screen.getByRole("button", { name: /create reminder/i }));

    await waitFor(() => {
      expect(toast.success).toHaveBeenCalledWith("Reminder created");
    });
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it("shows toast.error on mutation failure", async () => {
    const user = userEvent.setup();
    mockMutateAsync.mockRejectedValue(new Error("Network error"));
    const onOpenChange = vi.fn();
    render(<CreateReminderDialog open={true} onOpenChange={onOpenChange} />);

    await user.type(screen.getByPlaceholderText("Enter reminder title"), "Buy groceries");
    const form = screen.getByRole("dialog").querySelector("form")!;
    const scheduledInput = form.querySelector("input[name='scheduledFor']")!;
    await user.type(scheduledInput, "2026-04-01T10:00");

    await user.click(screen.getByRole("button", { name: /create reminder/i }));

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalled();
    });
    expect(onOpenChange).not.toHaveBeenCalledWith(false);
  });
});
