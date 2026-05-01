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
const mockUseMarkKioskSeeded = vi.fn();
vi.mock("@/lib/hooks/api/use-household-preferences", () => ({
  useHouseholdPreferences: () => mockUseHouseholdPreferences(),
  useMarkKioskSeeded: () => mockUseMarkKioskSeeded(),
}));

const mockUseTenant = vi.fn();
vi.mock("@/context/tenant-context", () => ({
  useTenant: () => mockUseTenant(),
}));

import { DashboardRedirect } from "@/pages/DashboardRedirect";

function prefs(
  defaultDashboardId: string | null,
  kioskDashboardSeeded: boolean = false,
  id: string = "hp-1",
) {
  const p: HouseholdPreferences = {
    id,
    type: "householdPreferences",
    attributes: {
      defaultDashboardId,
      kioskDashboardSeeded,
      createdAt: "2026-01-01T00:00:00Z",
      updatedAt: "2026-01-01T00:00:00Z",
    },
  };
  return { data: [p] };
}

function makeDashboard(
  id: string,
  scope: "household" | "user",
  sortOrder = 0,
  name?: string,
): Dashboard {
  return {
    id,
    type: "dashboards",
    attributes: {
      name: name ?? `Dash ${id}`,
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

/**
 * Configure mockUseSeedDashboard to return distinct mocks on each call —
 * DashboardRedirect calls useSeedDashboard() twice (once for Home, once
 * for Kiosk). Each call site needs its own mutateAsync spy so tests can
 * assert which one fired.
 */
function setupSeedHooks(
  homeMutateAsync: ReturnType<typeof vi.fn>,
  kioskMutateAsync: ReturnType<typeof vi.fn>,
) {
  // useSeedDashboard() is called twice per render — once for Home, once for
  // Kiosk. React may rerender, so we cycle returns by parity of call count
  // rather than mockReturnValueOnce (which exhausts after two calls).
  const homeReturn = {
    mutateAsync: homeMutateAsync,
    isError: false,
    error: null,
  };
  const kioskReturn = {
    mutateAsync: kioskMutateAsync,
    isError: false,
    error: null,
  };
  mockUseSeedDashboard.mockImplementation(() => {
    const calls = mockUseSeedDashboard.mock.calls.length;
    // 1st, 3rd, 5th, ... calls = Home; 2nd, 4th, 6th, ... = Kiosk.
    return calls % 2 === 1 ? homeReturn : kioskReturn;
  });
}

describe("DashboardRedirect", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Default: both seed hooks resolve quickly with stub responses.
    setupSeedHooks(
      vi.fn().mockResolvedValue({ data: { id: "home-seeded" } }),
      vi.fn().mockResolvedValue({ data: { id: "kiosk-seeded" } }),
    );
    mockUseMarkKioskSeeded.mockReturnValue({
      mutate: vi.fn(),
      isError: false,
      error: null,
    });
    mockUseTenant.mockReturnValue({
      tenant: { id: "t-1" },
      household: { id: "hh-1" },
    });
  });

  it("navigates to the preferred default when present in the list", async () => {
    mockUseHouseholdPreferences.mockReturnValue({ data: prefs("d-user", true) });
    mockUseDashboards.mockReturnValue({
      data: {
        data: [makeDashboard("d-household", "household"), makeDashboard("d-user", "user")],
      },
      refetch: vi.fn(),
    });

    renderIt();
    await waitFor(() =>
      expect(navigateMock).toHaveBeenCalledWith("/app/dashboards/d-user", { replace: true }),
    );
  });

  it("falls back to the first household-scoped dashboard when the preferred id is missing", async () => {
    mockUseHouseholdPreferences.mockReturnValue({ data: prefs("missing-id", true) });
    mockUseDashboards.mockReturnValue({
      data: {
        data: [makeDashboard("d-user", "user"), makeDashboard("d-household", "household")],
      },
      refetch: vi.fn(),
    });

    renderIt();
    await waitFor(() =>
      expect(navigateMock).toHaveBeenCalledWith("/app/dashboards/d-household", {
        replace: true,
      }),
    );
  });

  it("renders an error message when the dashboards query fails", async () => {
    mockUseHouseholdPreferences.mockReturnValue({ data: undefined });
    mockUseDashboards.mockReturnValue({
      data: undefined,
      refetch: vi.fn(),
      isError: true,
      error: new Error("dashboard-service unreachable"),
    });

    const { findByText } = renderIt();
    expect(await findByText(/Couldn't load dashboards/i)).toBeInTheDocument();
    expect(navigateMock).not.toHaveBeenCalled();
  });

  it("renders an empty-state when no household is selected", async () => {
    mockUseTenant.mockReturnValue({ tenant: { id: "t-1" }, household: null });
    mockUseHouseholdPreferences.mockReturnValue({ data: undefined });
    mockUseDashboards.mockReturnValue({ data: undefined, refetch: vi.fn() });

    const { findByText } = renderIt();
    expect(await findByText(/No household selected/i)).toBeInTheDocument();
    expect(navigateMock).not.toHaveBeenCalled();
  });

  describe("two-seed orchestration", () => {
    it("brand-new household: seeds Home + Kiosk, marks flag, navigates to Home", async () => {
      const homeSeedMutate = vi
        .fn()
        .mockResolvedValue({ data: { id: "home-1" } });
      const kioskSeedMutate = vi
        .fn()
        .mockResolvedValue({ data: { id: "kiosk-1" } });
      const markSeededMutate = vi.fn();
      setupSeedHooks(homeSeedMutate, kioskSeedMutate);
      mockUseMarkKioskSeeded.mockReturnValue({
        mutate: markSeededMutate,
        isError: false,
        error: null,
      });

      mockUseHouseholdPreferences.mockReturnValue({
        data: prefs(null, false, "hp-1"),
      });
      const refetchMock = vi.fn().mockResolvedValue({
        data: {
          data: [
            makeDashboard("home-1", "household", 0, "Home"),
            makeDashboard("kiosk-1", "household", 1, "Kiosk"),
          ],
        },
      });
      mockUseDashboards.mockReturnValue({
        data: { data: [] },
        refetch: refetchMock,
      });

      renderIt();

      await waitFor(() => expect(homeSeedMutate).toHaveBeenCalledTimes(1));
      expect(kioskSeedMutate).toHaveBeenCalledTimes(1);
      const homeArg = homeSeedMutate.mock.calls[0]?.[0] as {
        name: string;
        key?: string;
        layout: { widgets: unknown[] };
      };
      expect(homeArg).toMatchObject({ name: "Home", key: "home" });
      expect(homeArg.layout.widgets.length).toBeGreaterThan(0);
      const kioskArg = kioskSeedMutate.mock.calls[0]?.[0] as {
        name: string;
        key?: string;
        layout: { widgets: unknown[] };
      };
      expect(kioskArg).toMatchObject({ name: "Kiosk", key: "kiosk" });
      expect(kioskArg.layout.widgets.length).toBeGreaterThan(0);

      await waitFor(() => expect(markSeededMutate).toHaveBeenCalledWith("hp-1"));
      await waitFor(() =>
        expect(navigateMock).toHaveBeenCalledWith("/app/dashboards/home-1", {
          replace: true,
        }),
      );
    });

    it("brownfield household: Home exists, Kiosk seeded, flag marked", async () => {
      const homeSeedMutate = vi.fn();
      const kioskSeedMutate = vi
        .fn()
        .mockResolvedValue({ data: { id: "kiosk-1" } });
      const markSeededMutate = vi.fn();
      setupSeedHooks(homeSeedMutate, kioskSeedMutate);
      mockUseMarkKioskSeeded.mockReturnValue({
        mutate: markSeededMutate,
        isError: false,
        error: null,
      });

      mockUseHouseholdPreferences.mockReturnValue({
        data: prefs(null, false, "hp-1"),
      });
      const refetchMock = vi.fn().mockResolvedValue({
        data: {
          data: [
            makeDashboard("home-existing", "household", 0, "Home"),
            makeDashboard("kiosk-1", "household", 1, "Kiosk"),
          ],
        },
      });
      mockUseDashboards.mockReturnValue({
        data: {
          data: [makeDashboard("home-existing", "household", 0, "Home")],
        },
        refetch: refetchMock,
      });

      renderIt();

      await waitFor(() => expect(kioskSeedMutate).toHaveBeenCalledTimes(1));
      expect(homeSeedMutate).not.toHaveBeenCalled();
      await waitFor(() => expect(markSeededMutate).toHaveBeenCalledWith("hp-1"));
      await waitFor(() =>
        expect(navigateMock).toHaveBeenCalledWith(
          "/app/dashboards/home-existing",
          { replace: true },
        ),
      );
    });

    it("returning household: both seeded, flag set — no mutations fire", async () => {
      const homeSeedMutate = vi.fn();
      const kioskSeedMutate = vi.fn();
      const markSeededMutate = vi.fn();
      setupSeedHooks(homeSeedMutate, kioskSeedMutate);
      mockUseMarkKioskSeeded.mockReturnValue({
        mutate: markSeededMutate,
        isError: false,
        error: null,
      });

      mockUseHouseholdPreferences.mockReturnValue({
        data: prefs(null, true, "hp-1"),
      });
      const refetchMock = vi.fn();
      mockUseDashboards.mockReturnValue({
        data: {
          data: [
            makeDashboard("home-1", "household", 0, "Home"),
            makeDashboard("kiosk-1", "household", 1, "Kiosk"),
          ],
        },
        refetch: refetchMock,
      });

      renderIt();

      await waitFor(() =>
        expect(navigateMock).toHaveBeenCalledWith("/app/dashboards/home-1", {
          replace: true,
        }),
      );
      expect(homeSeedMutate).not.toHaveBeenCalled();
      expect(kioskSeedMutate).not.toHaveBeenCalled();
      expect(markSeededMutate).not.toHaveBeenCalled();
      // No refetch needed when nothing is seeded.
      expect(refetchMock).not.toHaveBeenCalled();
    });

    it("user deleted Kiosk: flag is gate, kiosk NOT re-seeded", async () => {
      const homeSeedMutate = vi.fn();
      const kioskSeedMutate = vi.fn();
      const markSeededMutate = vi.fn();
      setupSeedHooks(homeSeedMutate, kioskSeedMutate);
      mockUseMarkKioskSeeded.mockReturnValue({
        mutate: markSeededMutate,
        isError: false,
        error: null,
      });

      mockUseHouseholdPreferences.mockReturnValue({
        data: prefs(null, true, "hp-1"),
      });
      const refetchMock = vi.fn();
      mockUseDashboards.mockReturnValue({
        data: {
          data: [makeDashboard("home-1", "household", 0, "Home")],
        },
        refetch: refetchMock,
      });

      renderIt();

      await waitFor(() =>
        expect(navigateMock).toHaveBeenCalledWith("/app/dashboards/home-1", {
          replace: true,
        }),
      );
      expect(homeSeedMutate).not.toHaveBeenCalled();
      expect(kioskSeedMutate).not.toHaveBeenCalled();
      expect(markSeededMutate).not.toHaveBeenCalled();
    });
  });
});
