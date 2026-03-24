import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HouseholdSwitcher } from "../household-switcher";

const mockSetActiveHousehold = vi.fn();
const mockInvalidateAllTasks = vi.fn();
const mockInvalidateAllReminders = vi.fn();
const mockUseHouseholds = vi.fn(() => ({ data: null }));

vi.mock("@/components/providers/auth-provider", () => ({
  useAuth: () => ({
    appContext: {
      relationships: {
        activeHousehold: { data: { id: "hh-1" } },
      },
    },
  }),
}));

vi.mock("@/context/tenant-context", () => ({
  useTenant: () => ({
    setActiveHousehold: mockSetActiveHousehold,
  }),
}));

vi.mock("@/lib/hooks/api/use-households", () => ({
  useHouseholds: (...args: unknown[]) => mockUseHouseholds(...args),
}));

vi.mock("@/lib/hooks/api/use-tasks", () => ({
  useInvalidateTasks: () => ({ invalidateAll: mockInvalidateAllTasks }),
}));

vi.mock("@/lib/hooks/api/use-reminders", () => ({
  useInvalidateReminders: () => ({ invalidateAll: mockInvalidateAllReminders }),
}));

vi.mock("sonner", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

vi.mock("@/components/ui/select", () => {
  let capturedOnValueChange: ((value: string) => void) | undefined;

  return {
    Select: ({
      children,
      onValueChange,
    }: {
      value?: string;
      children: React.ReactNode;
      onValueChange?: (value: string) => void;
    }) => {
      capturedOnValueChange = onValueChange;
      return <div data-slot="select">{children}</div>;
    },
    SelectTrigger: ({ children }: { children: React.ReactNode; className?: string }) => (
      <button role="combobox">{children}</button>
    ),
    SelectValue: ({ placeholder }: { placeholder?: string }) => (
      <span>{placeholder}</span>
    ),
    SelectContent: ({ children }: { children: React.ReactNode }) => (
      <div role="listbox">{children}</div>
    ),
    SelectItem: ({
      children,
      value,
    }: {
      children: React.ReactNode;
      value: string;
    }) => (
      <button
        role="option"
        onClick={() => capturedOnValueChange?.(value)}
      >
        {children}
      </button>
    ),
  };
});

import { toast } from "sonner";

const multipleHouseholds = {
  data: [
    { id: "hh-1", attributes: { name: "Main House" } },
    { id: "hh-2", attributes: { name: "Beach House" } },
  ],
};

describe("HouseholdSwitcher", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseHouseholds.mockReturnValue({ data: null });
  });

  it("returns null when households list is empty", () => {
    const { container } = render(<HouseholdSwitcher />);
    expect(container.innerHTML).toBe("");
  });

  it("returns null when only one household exists", () => {
    mockUseHouseholds.mockReturnValue({
      data: { data: [{ id: "hh-1", attributes: { name: "Main House" } }] },
    });

    const { container } = render(<HouseholdSwitcher />);
    expect(container.innerHTML).toBe("");
  });

  it("renders select trigger when multiple households exist", () => {
    mockUseHouseholds.mockReturnValue({
      data: multipleHouseholds,
    });

    render(<HouseholdSwitcher />);

    expect(screen.getByRole("combobox")).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "Main House" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "Beach House" })).toBeInTheDocument();
  });

  it("shows toast.success after switching household", async () => {
    const user = userEvent.setup();
    mockSetActiveHousehold.mockResolvedValue(undefined);
    mockUseHouseholds.mockReturnValue({
      data: multipleHouseholds,
    });

    render(<HouseholdSwitcher />);

    await user.click(screen.getByRole("option", { name: "Beach House" }));

    await waitFor(() => {
      expect(mockSetActiveHousehold).toHaveBeenCalledWith("hh-2");
    });
    expect(toast.success).toHaveBeenCalledWith("Household switched");
  });

  it("invalidates tasks and reminders after successful switch", async () => {
    const user = userEvent.setup();
    mockSetActiveHousehold.mockResolvedValue(undefined);
    mockUseHouseholds.mockReturnValue({
      data: multipleHouseholds,
    });

    render(<HouseholdSwitcher />);

    await user.click(screen.getByRole("option", { name: "Beach House" }));

    await waitFor(() => {
      expect(mockInvalidateAllTasks).toHaveBeenCalled();
    });
    expect(mockInvalidateAllReminders).toHaveBeenCalled();
  });

  it("shows toast.error when setActiveHousehold fails", async () => {
    const user = userEvent.setup();
    mockSetActiveHousehold.mockRejectedValue(new Error("Network error"));
    mockUseHouseholds.mockReturnValue({
      data: multipleHouseholds,
    });

    render(<HouseholdSwitcher />);

    await user.click(screen.getByRole("option", { name: "Beach House" }));

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalled();
    });
    expect(toast.success).not.toHaveBeenCalled();
  });
});
