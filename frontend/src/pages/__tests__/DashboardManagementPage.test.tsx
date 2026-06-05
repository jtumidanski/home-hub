import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import type { Dashboard } from "@/types/models/dashboard";

function dash(overrides: Partial<Dashboard["attributes"]> & { id: string }): Dashboard {
  const { id, ...attrs } = overrides;
  return {
    id,
    type: "dashboards",
    attributes: {
      name: id,
      scope: "household",
      sortOrder: 0,
      layout: { version: 1, widgets: [] },
      schemaVersion: 1,
      createdAt: "2025-01-01T00:00:00Z",
      updatedAt: "2025-01-01T00:00:00Z",
      ...attrs,
    },
  };
}

const mockUseDashboards = vi.fn();
const mockUseHouseholdPreferences = vi.fn();

vi.mock("@/lib/hooks/api/use-dashboards", () => ({
  useDashboards: () => mockUseDashboards(),
  useReorderDashboards: () => ({ mutate: vi.fn() }),
  useCreateDashboard: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useCopyDashboardToMine: () => ({ mutate: vi.fn(), mutateAsync: vi.fn(), isPending: false }),
  useUpdateDashboard: () => ({ mutate: vi.fn(), isPending: false }),
  useDeleteDashboard: () => ({ mutate: vi.fn(), isPending: false }),
  usePromoteDashboard: () => ({ mutate: vi.fn() }),
}));

vi.mock("@/lib/hooks/api/use-household-preferences", () => ({
  useHouseholdPreferences: () => mockUseHouseholdPreferences(),
  useUpdateHouseholdPreferences: () => ({ mutate: vi.fn() }),
}));

import { DashboardManagementPage } from "../DashboardManagementPage";

function renderPage() {
  return render(
    <MemoryRouter>
      <DashboardManagementPage />
    </MemoryRouter>,
  );
}

describe("DashboardManagementPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseHouseholdPreferences.mockReturnValue({
      data: { data: { id: "prefs-1", type: "householdPreferences", attributes: { defaultDashboardId: null, kioskDashboardSeeded: false, createdAt: "", updatedAt: "" } } },
    });
    mockUseDashboards.mockReturnValue({ data: null, isLoading: false, isError: false });
  });

  it("renders a loading skeleton while dashboards load", () => {
    mockUseDashboards.mockReturnValue({ data: null, isLoading: true, isError: false });
    renderPage();
    expect(screen.getByRole("status", { name: "Loading" })).toBeInTheDocument();
  });

  it("renders an error state on failure", () => {
    mockUseDashboards.mockReturnValue({ data: null, isLoading: false, isError: true });
    renderPage();
    expect(screen.getByText(/failed to load dashboards/i)).toBeInTheDocument();
  });

  it("renders household and user sections, each sorted, and an open link per row", () => {
    mockUseDashboards.mockReturnValue({
      data: {
        data: [
          dash({ id: "hh-2", name: "Home B", scope: "household", sortOrder: 1 }),
          dash({ id: "u-1", name: "Mine A", scope: "user", sortOrder: 0 }),
          dash({ id: "hh-1", name: "Home A", scope: "household", sortOrder: 0 }),
        ],
      },
      isLoading: false,
      isError: false,
    });
    renderPage();

    expect(screen.getByText("Household Dashboards")).toBeInTheDocument();
    expect(screen.getByText("My Dashboards")).toBeInTheDocument();

    const openLinks = screen
      .getAllByRole("link")
      .filter((a) => /^\/app\/dashboards\/[^/]+$/.test(a.getAttribute("href") ?? ""));
    expect(openLinks.map((a) => a.getAttribute("href"))).toEqual([
      "/app/dashboards/hh-1",
      "/app/dashboards/hh-2",
      "/app/dashboards/u-1",
    ]);
  });

  it("omits a scope section that has no dashboards", () => {
    mockUseDashboards.mockReturnValue({
      data: { data: [dash({ id: "hh-1", scope: "household", sortOrder: 0 })] },
      isLoading: false,
      isError: false,
    });
    renderPage();
    expect(screen.getByText("Household Dashboards")).toBeInTheDocument();
    expect(screen.queryByText("My Dashboards")).not.toBeInTheDocument();
  });

  it("renders a page-level empty state when there are no dashboards", () => {
    mockUseDashboards.mockReturnValue({ data: { data: [] }, isLoading: false, isError: false });
    renderPage();
    expect(screen.getByText(/no dashboards yet/i)).toBeInTheDocument();
    expect(screen.queryByText("Household Dashboards")).not.toBeInTheDocument();
    expect(screen.queryByText("My Dashboards")).not.toBeInTheDocument();
  });

  it("opens the New Dashboard modal when the header button is clicked", async () => {
    const user = userEvent.setup();
    mockUseDashboards.mockReturnValue({ data: { data: [] }, isLoading: false, isError: false });
    renderPage();

    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: /new dashboard/i }));
    expect(screen.getByRole("dialog")).toBeInTheDocument();
  });
});
