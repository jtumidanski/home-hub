import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { RemindersTodayAdapter } from "@/components/features/dashboard-widgets/reminders-today-adapter";

vi.mock("@/lib/hooks/api/use-reminders", () => ({ useReminders: vi.fn() }));

import { useReminders } from "@/lib/hooks/api/use-reminders";

const reminder = (id: string, title: string, scheduledFor: string, active = true) => ({
  id, type: "reminders",
  attributes: { title, scheduledFor, active, createdAt: "", updatedAt: "" },
});

describe("RemindersTodayAdapter", () => {
  it("renders the active list capped by limit", () => {
    (useReminders as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: [
        reminder("1", "Take meds", "2026-05-01T08:00:00Z"),
        reminder("2", "Standup", "2026-05-01T09:00:00Z"),
        reminder("3", "Inactive", "2026-05-01T10:00:00Z", false),
      ] },
      isLoading: false, isError: false,
    });
    render(<MemoryRouter><RemindersTodayAdapter config={{ limit: 5 }} /></MemoryRouter>);
    expect(screen.getByText("Take meds")).toBeInTheDocument();
    expect(screen.getByText("Standup")).toBeInTheDocument();
    expect(screen.queryByText("Inactive")).not.toBeInTheDocument();
  });

  it("shows empty copy when no active reminders", () => {
    (useReminders as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: [] }, isLoading: false, isError: false,
    });
    render(<MemoryRouter><RemindersTodayAdapter config={{ limit: 5 }} /></MemoryRouter>);
    expect(screen.getByText(/No active reminders/i)).toBeInTheDocument();
  });

  it("respects the limit", () => {
    (useReminders as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: [
        reminder("1", "A", "2026-05-01T08:00:00Z"),
        reminder("2", "B", "2026-05-01T09:00:00Z"),
        reminder("3", "C", "2026-05-01T10:00:00Z"),
      ] },
      isLoading: false, isError: false,
    });
    render(<MemoryRouter><RemindersTodayAdapter config={{ limit: 2 }} /></MemoryRouter>);
    expect(screen.getAllByRole("listitem")).toHaveLength(2);
  });
});
