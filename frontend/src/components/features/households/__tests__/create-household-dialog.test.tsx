import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { CreateHouseholdDialog } from "../create-household-dialog";

const mockMutateAsync = vi.fn();

vi.mock("@/lib/hooks/api/use-households", () => ({
  useCreateHousehold: () => ({ mutateAsync: mockMutateAsync }),
}));

vi.mock("sonner", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

import { toast } from "sonner";

describe("CreateHouseholdDialog", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("does not render content when closed", () => {
    render(<CreateHouseholdDialog open={false} onOpenChange={vi.fn()} />);
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("renders form fields when open", () => {
    render(<CreateHouseholdDialog open={true} onOpenChange={vi.fn()} />);
    expect(screen.getByRole("dialog")).toBeInTheDocument();
    expect(screen.getByText("Name")).toBeInTheDocument();
    expect(screen.getByText("Timezone")).toBeInTheDocument();
    expect(screen.getByRole("radio", { name: /imperial/i })).toBeInTheDocument();
    expect(screen.getByRole("radio", { name: /metric/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /create household/i })).toBeInTheDocument();
  });

  it("shows validation error when submitting with empty name", async () => {
    const user = userEvent.setup();
    render(<CreateHouseholdDialog open={true} onOpenChange={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: /create household/i }));

    await waitFor(() => {
      expect(screen.getByText("Name is required")).toBeInTheDocument();
    });
    expect(mockMutateAsync).not.toHaveBeenCalled();
  });

  it("calls mutateAsync with correct data on valid submission", async () => {
    const user = userEvent.setup();
    mockMutateAsync.mockResolvedValue({});
    const onOpenChange = vi.fn();

    render(<CreateHouseholdDialog open={true} onOpenChange={onOpenChange} />);

    await user.type(screen.getByPlaceholderText("Enter household name"), "My Home");
    await user.click(screen.getByRole("button", { name: /create household/i }));

    await waitFor(() => {
      expect(mockMutateAsync).toHaveBeenCalledWith({
        name: "My Home",
        timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
        units: "imperial",
      });
    });
  });

  it("shows toast.success and closes dialog on successful submission", async () => {
    const user = userEvent.setup();
    mockMutateAsync.mockResolvedValue({});
    const onOpenChange = vi.fn();

    render(<CreateHouseholdDialog open={true} onOpenChange={onOpenChange} />);

    await user.type(screen.getByPlaceholderText("Enter household name"), "My Home");
    await user.click(screen.getByRole("button", { name: /create household/i }));

    await waitFor(() => {
      expect(toast.success).toHaveBeenCalledWith("Household created");
    });
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it("shows toast.error on mutation failure", async () => {
    const user = userEvent.setup();
    mockMutateAsync.mockRejectedValue(new Error("Network error"));
    const onOpenChange = vi.fn();

    render(<CreateHouseholdDialog open={true} onOpenChange={onOpenChange} />);

    await user.type(screen.getByPlaceholderText("Enter household name"), "My Home");
    await user.click(screen.getByRole("button", { name: /create household/i }));

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalled();
    });
    expect(onOpenChange).not.toHaveBeenCalledWith(false);
  });

  it("has Imperial radio checked by default", () => {
    render(<CreateHouseholdDialog open={true} onOpenChange={vi.fn()} />);

    const imperialRadio = screen.getByRole("radio", { name: /imperial/i });
    const metricRadio = screen.getByRole("radio", { name: /metric/i });

    expect(imperialRadio).toBeChecked();
    expect(metricRadio).not.toBeChecked();
  });
});
