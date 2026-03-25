import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

const mockUseReminders = vi.fn();
const mockSnoozeMutateAsync = vi.fn();
const mockDismissMutateAsync = vi.fn();
const mockDeleteMutateAsync = vi.fn();

vi.mock("@/lib/hooks/api/use-reminders", () => ({
  useReminders: () => mockUseReminders(),
  useSnoozeReminder: () => ({ mutateAsync: mockSnoozeMutateAsync }),
  useDismissReminder: () => ({ mutateAsync: mockDismissMutateAsync }),
  useDeleteReminder: () => ({ mutateAsync: mockDeleteMutateAsync }),
}));

vi.mock("@/lib/api/errors", () => ({
  getErrorMessage: (_err: unknown, fallback: string) => fallback,
}));

vi.mock("@/types/models/reminder", () => ({
  isReminderDismissed: () => false,
  isReminderSnoozed: () => false,
}));

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

vi.mock("@/components/features/reminders/create-reminder-dialog", () => ({
  CreateReminderDialog: ({ open }: { open: boolean }) =>
    open ? <div data-testid="create-reminder-dialog">CreateReminderDialog</div> : null,
}));

import { RemindersPage } from "../RemindersPage";

describe("RemindersPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseReminders.mockReturnValue({ data: null, isLoading: false, isError: false });
  });

  it("renders loading skeleton when isLoading is true", () => {
    mockUseReminders.mockReturnValue({ data: null, isLoading: true, isError: false });
    const { container } = render(<RemindersPage />);
    expect(screen.queryByText("Reminders")).not.toBeInTheDocument();
    expect(container.querySelector(".animate-pulse")).toBeTruthy();
  });

  it("renders error state when isError is true", () => {
    mockUseReminders.mockReturnValue({ data: null, isLoading: false, isError: true });
    render(<RemindersPage />);
    expect(screen.getByText(/failed to load reminders/i)).toBeInTheDocument();
  });

  it("renders empty state when there are no reminders", () => {
    mockUseReminders.mockReturnValue({ data: { data: [] }, isLoading: false, isError: false });
    render(<RemindersPage />);
    expect(screen.getByText("Reminders")).toBeInTheDocument();
    expect(screen.getByText(/no reminders yet/i)).toBeInTheDocument();
    expect(screen.getByText("Create First Reminder")).toBeInTheDocument();
  });

  it("renders reminder list when reminders exist", () => {
    mockUseReminders.mockReturnValue({
      data: {
        data: [
          {
            id: "r-1",
            type: "reminders",
            attributes: { title: "Doctor appointment", scheduledFor: "2026-04-01T10:00:00Z", active: true },
          },
        ],
      },
      isLoading: false,
      isError: false,
    });
    render(<RemindersPage />);
    expect(screen.getByText("Doctor appointment")).toBeInTheDocument();
    expect(screen.getByText("active")).toBeInTheDocument();
  });

  it("opens create dialog when New Reminder button is clicked", async () => {
    const user = userEvent.setup();
    mockUseReminders.mockReturnValue({ data: { data: [] }, isLoading: false, isError: false });
    render(<RemindersPage />);

    expect(screen.queryByTestId("create-reminder-dialog")).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: /new reminder/i }));
    expect(screen.getByTestId("create-reminder-dialog")).toBeInTheDocument();
  });
});
