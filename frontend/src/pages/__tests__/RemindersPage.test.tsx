import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";

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

vi.mock("@/lib/hooks/api/use-household-members", () => ({
  useMemberMap: () => new Map(),
  useHouseholdMembers: () => ({ data: { data: [] } }),
}));

vi.mock("@/components/providers/auth-provider", () => ({
  useAuth: () => ({ user: { id: "user-1" } }),
}));

vi.mock("@/lib/api/errors", () => ({
  createErrorFromUnknown: (_err: unknown, fallback: string) => ({ message: fallback, type: "unknown" }),
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
    open ? <div role="dialog">CreateReminderDialog</div> : null,
}));

import { RemindersPage } from "../RemindersPage";

function renderWithRouter(ui: React.ReactElement) {
  return render(<MemoryRouter>{ui}</MemoryRouter>);
}

describe("RemindersPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseReminders.mockReturnValue({ data: null, isLoading: false, isError: false });
  });

  it("renders loading skeleton when isLoading is true", () => {
    mockUseReminders.mockReturnValue({ data: null, isLoading: true, isError: false });
    renderWithRouter(<RemindersPage />);
    expect(screen.queryByText("Reminders")).not.toBeInTheDocument();
    expect(screen.getByRole("status", { name: "Loading" })).toBeInTheDocument();
  });

  it("renders error state when isError is true", () => {
    mockUseReminders.mockReturnValue({ data: null, isLoading: false, isError: true });
    renderWithRouter(<RemindersPage />);
    expect(screen.getByText(/failed to load reminders/i)).toBeInTheDocument();
  });

  it("renders empty state when there are no reminders", () => {
    mockUseReminders.mockReturnValue({ data: { data: [] }, isLoading: false, isError: false });
    renderWithRouter(<RemindersPage />);
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
    renderWithRouter(<RemindersPage />);
    expect(screen.getByText("Doctor appointment")).toBeInTheDocument();
    expect(screen.getByText("active")).toBeInTheDocument();
  });

  it("opens create dialog when New Reminder button is clicked", async () => {
    const user = userEvent.setup();
    mockUseReminders.mockReturnValue({ data: { data: [] }, isLoading: false, isError: false });
    renderWithRouter(<RemindersPage />);

    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: /new reminder/i }));
    expect(screen.getByRole("dialog")).toBeInTheDocument();
  });
});
