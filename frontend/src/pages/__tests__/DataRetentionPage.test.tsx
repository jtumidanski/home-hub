import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

const mockPurgeAsync = vi.fn();

vi.mock("@/lib/hooks/api/use-retention", () => ({
  useRetentionPolicies: () => ({
    isLoading: false,
    error: null,
    data: {
      data: {
        attributes: {
          household: {
            id: "h-1",
            categories: {
              "productivity.reminders": { days: 30, source: "default" },
            },
          },
          user: null,
        },
      },
    },
  }),
  useRetentionRuns: () => ({ isLoading: false, data: { data: [] } }),
  usePatchHouseholdRetention: () => ({ mutateAsync: vi.fn(), isPending: false }),
  usePatchUserRetention: () => ({ mutateAsync: vi.fn(), isPending: false }),
  usePurgeRetention: () => ({ mutateAsync: mockPurgeAsync, isPending: false }),
}));

vi.mock("@/context/tenant-context", () => ({
  useTenant: () => ({ tenant: { id: "t-1" } }),
}));

import { DataRetentionPage } from "../DataRetentionPage";

describe("DataRetentionPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockPurgeAsync.mockResolvedValue({ data: { attributes: { deleted: 3 } } });
  });

  it("labels the days input and ties validation errors to it", async () => {
    const user = userEvent.setup();
    render(<DataRetentionPage />);
    const input = screen.getByRole("spinbutton", { name: "Reminders retention days" });
    expect(input).not.toHaveAttribute("aria-invalid");
    await user.clear(input);
    await user.type(input, "0");
    await user.click(screen.getByRole("button", { name: "Save" }));
    await waitFor(() => expect(input).toHaveAttribute("aria-invalid", "true"));
    const errorId = input.getAttribute("aria-describedby");
    expect(errorId).toBeTruthy();
    expect(document.getElementById(errorId!)).toHaveTextContent("Minimum 1 day");
  });

  it("shows a friendly source badge, not the raw key", () => {
    render(<DataRetentionPage />);
    expect(screen.getByText("Default")).toBeInTheDocument();
    expect(screen.queryByText("productivity.reminders")).not.toBeInTheDocument();
  });

  it("does not purge until the confirm dialog is accepted", async () => {
    const user = userEvent.setup();
    render(<DataRetentionPage />);
    await user.click(screen.getByRole("button", { name: "Purge Reminders now" }));
    expect(mockPurgeAsync).not.toHaveBeenCalled();
    const dialog = await screen.findByRole("dialog");
    await user.click(screen.getByRole("button", { name: "Purge now" }));
    expect(mockPurgeAsync).toHaveBeenCalledWith({ category: "productivity.reminders", scope: "household" });
    await waitFor(() => expect(dialog).not.toBeInTheDocument());
  });

  it("does not purge when the confirm dialog is cancelled", async () => {
    const user = userEvent.setup();
    render(<DataRetentionPage />);
    await user.click(screen.getByRole("button", { name: "Purge Reminders now" }));
    await user.click(await screen.findByRole("button", { name: "Cancel" }));
    expect(mockPurgeAsync).not.toHaveBeenCalled();
  });

  it("renders an empty state when there are no reaper runs", () => {
    render(<DataRetentionPage />);
    expect(screen.getByText("No reaper runs yet")).toBeInTheDocument();
  });
});
