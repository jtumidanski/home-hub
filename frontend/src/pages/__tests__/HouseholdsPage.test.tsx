import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

const mockUseAuth = vi.fn();
const mockUseHouseholds = vi.fn();

vi.mock("@/components/providers/auth-provider", () => ({
  useAuth: () => mockUseAuth(),
}));

vi.mock("@/lib/hooks/api/use-households", () => ({
  useHouseholds: () => mockUseHouseholds(),
}));

vi.mock("@/components/features/households/create-household-dialog", () => ({
  CreateHouseholdDialog: ({ open }: { open: boolean }) =>
    open ? <div role="dialog">CreateHouseholdDialog</div> : null,
}));

import { HouseholdsPage } from "../HouseholdsPage";

describe("HouseholdsPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseAuth.mockReturnValue({
      appContext: {
        attributes: { canCreateHousehold: true },
        relationships: { activeHousehold: { data: { id: "hh-1" } } },
      },
    });
    mockUseHouseholds.mockReturnValue({ data: null, isLoading: false, isError: false });
  });

  it("renders loading skeleton when isLoading is true", () => {
    mockUseHouseholds.mockReturnValue({ data: null, isLoading: true, isError: false });
    render(<HouseholdsPage />);
    expect(screen.queryByText("Households")).not.toBeInTheDocument();
    expect(screen.getByRole("status", { name: "Loading" })).toBeInTheDocument();
  });

  it("renders error state when isError is true", () => {
    mockUseHouseholds.mockReturnValue({ data: null, isLoading: false, isError: true });
    render(<HouseholdsPage />);
    expect(screen.getByText(/failed to load households/i)).toBeInTheDocument();
  });

  it("renders empty state when there are no households", () => {
    mockUseHouseholds.mockReturnValue({ data: { data: [] }, isLoading: false, isError: false });
    render(<HouseholdsPage />);
    expect(screen.getByText("Households")).toBeInTheDocument();
    expect(screen.getByText(/no households yet/i)).toBeInTheDocument();
  });

  it("renders household list with active badge", () => {
    mockUseHouseholds.mockReturnValue({
      data: {
        data: [
          { id: "hh-1", type: "households", attributes: { name: "Main House", timezone: "UTC", units: "imperial" } },
          { id: "hh-2", type: "households", attributes: { name: "Beach House", timezone: "US/Eastern", units: "metric" } },
        ],
      },
      isLoading: false,
      isError: false,
    });
    render(<HouseholdsPage />);
    expect(screen.getByText("Main House")).toBeInTheDocument();
    expect(screen.getByText("Beach House")).toBeInTheDocument();
    expect(screen.getByText("Active")).toBeInTheDocument();
  });

  it("opens create dialog when New Household button is clicked", async () => {
    const user = userEvent.setup();
    mockUseHouseholds.mockReturnValue({ data: { data: [] }, isLoading: false, isError: false });
    render(<HouseholdsPage />);

    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: /new household/i }));
    expect(screen.getByRole("dialog")).toBeInTheDocument();
  });
});
