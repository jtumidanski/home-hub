import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MobileHouseholdSelector } from "../mobile-household-selector";

const mockSetActiveHousehold = vi.fn();

vi.mock("@/components/providers/auth-provider", () => ({
  useAuth: () => ({
    appContext: { relationships: { activeHousehold: { data: { id: "hh-1" } } } },
  }),
}));

vi.mock("@/context/tenant-context", () => ({
  useTenant: () => ({ setActiveHousehold: mockSetActiveHousehold }),
}));

vi.mock("@/lib/hooks/api/use-households", () => ({
  useHouseholds: () => ({
    data: {
      data: [
        { id: "hh-1", attributes: { name: "Main House" } },
        { id: "hh-2", attributes: { name: "Beach House" } },
      ],
    },
  }),
}));

vi.mock("@/lib/hooks/api/use-tasks", () => ({
  useInvalidateTasks: () => ({ invalidateAll: vi.fn() }),
}));

vi.mock("@/lib/hooks/api/use-reminders", () => ({
  useInvalidateReminders: () => ({ invalidateAll: vi.fn() }),
}));

vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn() } }));

describe("MobileHouseholdSelector", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockSetActiveHousehold.mockResolvedValue(undefined);
  });

  it("trigger advertises a dialog and reflects expanded state", async () => {
    const user = userEvent.setup();
    render(<MobileHouseholdSelector />);
    const trigger = screen.getByRole("button", { name: /main house/i });
    expect(trigger).toHaveAttribute("aria-haspopup", "dialog");
    expect(trigger).toHaveAttribute("aria-expanded", "false");
    await user.click(trigger);
    expect(await screen.findByRole("dialog", { name: "Select Household" })).toBeInTheDocument();
  });

  it("marks only the active row with aria-current", async () => {
    const user = userEvent.setup();
    render(<MobileHouseholdSelector />);
    await user.click(screen.getByRole("button", { name: /main house/i }));
    const dialog = await screen.findByRole("dialog");
    expect(dialog).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Main House", current: true })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Beach House", current: false })).toBeInTheDocument();
  });

  it("closes on Escape", async () => {
    const user = userEvent.setup();
    render(<MobileHouseholdSelector />);
    await user.click(screen.getByRole("button", { name: /main house/i }));
    await screen.findByRole("dialog");
    await user.keyboard("{Escape}");
    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
  });

  it("switches household and closes", async () => {
    const user = userEvent.setup();
    render(<MobileHouseholdSelector />);
    await user.click(screen.getByRole("button", { name: /main house/i }));
    await user.click(await screen.findByRole("button", { name: "Beach House" }));
    expect(mockSetActiveHousehold).toHaveBeenCalledWith("hh-2");
    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
  });
});
