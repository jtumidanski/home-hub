import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

const mockPurgeAsync = vi.fn();
const mockPreview = vi.fn();
const mockPatchHousehold = vi.fn();
let patchPending = false;

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
  usePatchHouseholdRetention: () => ({ mutateAsync: mockPatchHousehold, isPending: patchPending }),
  usePatchUserRetention: () => ({ mutateAsync: vi.fn(), isPending: false }),
  usePurgeRetention: () => ({ mutateAsync: mockPurgeAsync, isPending: false }),
}));

vi.mock("@/context/tenant-context", () => ({
  useTenant: () => ({ tenant: { id: "t-1" } }),
}));

vi.mock("@/services/api/retention", () => ({
  retentionService: { purge: (...args: unknown[]) => mockPreview(...args) },
}));

import { DataRetentionPage } from "../DataRetentionPage";

async function openPurge(user: ReturnType<typeof userEvent.setup>) {
  await user.click(screen.getByRole("button", { name: "More actions for Reminders" }));
  await user.click(await screen.findByRole("menuitem", { name: "Purge now" }));
}

describe("DataRetentionPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    patchPending = false;
    mockPatchHousehold.mockResolvedValue({});
    mockPreview.mockResolvedValue({ data: { attributes: { deleted: 7 } } });
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
    await openPurge(user);
    expect(mockPurgeAsync).not.toHaveBeenCalled();
    const dialog = await screen.findByRole("dialog");
    await user.click(within(dialog).getByRole("button", { name: "Purge now" }));
    expect(mockPurgeAsync).toHaveBeenCalledWith({ category: "productivity.reminders", scope: "household" });
    await waitFor(() => expect(dialog).not.toBeInTheDocument());
  });

  it("does not purge when the confirm dialog is cancelled", async () => {
    const user = userEvent.setup();
    render(<DataRetentionPage />);
    await openPurge(user);
    await user.click(await screen.findByRole("button", { name: "Cancel" }));
    expect(mockPurgeAsync).not.toHaveBeenCalled();
  });

  it("renders an empty state when there are no reaper runs", () => {
    render(<DataRetentionPage />);
    expect(screen.getByText("No reaper runs yet")).toBeInTheDocument();
  });

  it("shrink-warning confirm is destructive-styled and disabled while the patch is pending (DA-138)", async () => {
    const user = userEvent.setup();
    const { rerender } = render(<DataRetentionPage />);
    const input = screen.getByRole("spinbutton", { name: "Reminders retention days" });
    await user.clear(input);
    await user.type(input, "10");
    await user.click(screen.getByRole("button", { name: "Save" }));
    const dialog = await screen.findByRole("dialog");
    const confirm = within(dialog).getByRole("button", { name: "Confirm" });
    expect(confirm).toBeEnabled();
    expect(confirm.className).toContain("destructive");

    patchPending = true;
    rerender(<DataRetentionPage />);
    const pendingBtn = within(screen.getByRole("dialog")).getByRole("button", { name: "Working…" });
    expect(pendingBtn).toBeDisabled();
    expect(pendingBtn.className).toContain("destructive");
  });

  it("keeps the shrink dialog open without an unhandled rejection when the save fails", async () => {
    mockPatchHousehold.mockRejectedValue(new Error("boom"));
    const user = userEvent.setup();
    render(<DataRetentionPage />);
    const input = screen.getByRole("spinbutton", { name: "Reminders retention days" });
    await user.clear(input);
    await user.type(input, "10");
    await user.click(screen.getByRole("button", { name: "Save" }));
    const dialog = await screen.findByRole("dialog");
    await user.click(within(dialog).getByRole("button", { name: "Confirm" }));
    await waitFor(() => expect(mockPatchHousehold).toHaveBeenCalled());
    expect(screen.getByRole("dialog")).toBeInTheDocument();
  });
});
