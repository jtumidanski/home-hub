import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import type { Dashboard, HouseholdPreferences } from "@/types/models/dashboard";

const navigateMock = vi.fn();
vi.mock("react-router-dom", async () => {
  const actual = await vi.importActual<typeof import("react-router-dom")>(
    "react-router-dom",
  );
  return {
    ...actual,
    useNavigate: () => navigateMock,
  };
});

const mockUseDashboards = vi.fn();
const mockUseSeedDashboard = vi.fn();
vi.mock("@/lib/hooks/api/use-dashboards", () => ({
  useDashboards: () => mockUseDashboards(),
  useSeedDashboard: () => mockUseSeedDashboard(),
}));

const mockUseHouseholdPreferences = vi.fn();
vi.mock("@/lib/hooks/api/use-household-preferences", () => ({
  useHouseholdPreferences: () => mockUseHouseholdPreferences(),
}));

import { DashboardRedirect } from "@/pages/DashboardRedirect";

function prefs(defaultDashboardId: string | null) {
  const p: HouseholdPreferences = {
    id: "hp-1",
    type: "householdPreferences",
    attributes: {
      defaultDashboardId,
      createdAt: "2026-01-01T00:00:00Z",
      updatedAt: "2026-01-01T00:00:00Z",
    },
  };
  return { data: [p] };
}

function makeDashboard(id: string, scope: "household" | "user", sortOrder = 0): Dashboard {
  return {
    id,
    type: "dashboards",
    attributes: {
      name: `Dash ${id}`,
      scope,
      sortOrder,
      schemaVersion: 1,
      createdAt: "2026-01-01T00:00:00Z",
      updatedAt: "2026-01-01T00:00:00Z",
      layout: { version: 1, widgets: [] },
    },
  };
}

function renderIt() {
  return render(
    <MemoryRouter>
      <DashboardRedirect />
    </MemoryRouter>,
  );
}

describe("DashboardRedirect", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseSeedDashboard.mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
      isSuccess: false,
    });
  });

  it("navigates to the preferred default when present in the list", async () => {
    mockUseHouseholdPreferences.mockReturnValue({ data: prefs("d-user") });
    mockUseDashboards.mockReturnValue({
      data: { data: [makeDashboard("d-household", "household"), makeDashboard("d-user", "user")] },
      refetch: vi.fn(),
    });

    renderIt();
    await waitFor(() =>
      expect(navigateMock).toHaveBeenCalledWith("/app/dashboards/d-user", { replace: true }),
    );
  });

  it("falls back to the first household-scoped dashboard when the preferred id is missing", async () => {
    mockUseHouseholdPreferences.mockReturnValue({ data: prefs("missing-id") });
    mockUseDashboards.mockReturnValue({
      data: { data: [makeDashboard("d-user", "user"), makeDashboard("d-household", "household")] },
      refetch: vi.fn(),
    });

    renderIt();
    await waitFor(() =>
      expect(navigateMock).toHaveBeenCalledWith("/app/dashboards/d-household", { replace: true }),
    );
  });

  it("calls the seed mutation when the dashboard list is empty", async () => {
    const mutateMock = vi.fn();
    mockUseSeedDashboard.mockReturnValue({
      mutate: mutateMock,
      isPending: false,
      isSuccess: false,
    });
    mockUseHouseholdPreferences.mockReturnValue({ data: prefs(null) });
    mockUseDashboards.mockReturnValue({
      data: { data: [] },
      refetch: vi.fn(),
    });

    renderIt();
    await waitFor(() => expect(mutateMock).toHaveBeenCalledTimes(1));
    const firstArg = mutateMock.mock.calls[0]?.[0] as {
      name: string;
      layout: { widgets: unknown[] };
    };
    expect(firstArg).toMatchObject({ name: "Home" });
    expect(firstArg.layout.widgets.length).toBeGreaterThan(0);
  });
});
