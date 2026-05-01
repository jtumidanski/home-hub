import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import type { Dashboard } from "@/types/models/dashboard";

vi.mock("react-router-dom", async () => {
  const actual = await vi.importActual<typeof import("react-router-dom")>("react-router-dom");
  return { ...actual, useNavigate: () => vi.fn() };
});

vi.mock("@/lib/hooks/api/use-dashboards", () => ({
  useUpdateDashboard: () => ({ mutate: vi.fn(), isPending: false }),
  useDeleteDashboard: () => ({ mutate: vi.fn(), isPending: false }),
  usePromoteDashboard: () => ({ mutate: vi.fn() }),
  useCopyDashboardToMine: () => ({ mutate: vi.fn() }),
}));

vi.mock("@/lib/hooks/api/use-household-preferences", () => ({
  useHouseholdPreferences: () => ({
    data: { data: { id: "prefs-1", type: "householdPreferences", attributes: { defaultDashboardId: null, kioskDashboardSeeded: false, createdAt: "", updatedAt: "" } } },
  }),
  useUpdateHouseholdPreferences: () => ({ mutate: vi.fn() }),
}));

import { DashboardKebabMenu } from "../dashboard-kebab-menu";

function dash(scope: "household" | "user" = "household", id = "d-1"): Dashboard {
  return {
    id,
    type: "dashboards",
    attributes: {
      name: `Name ${id}`,
      scope,
      sortOrder: 0,
      layout: { version: 1, widgets: [] },
      schemaVersion: 1,
      createdAt: "2025-01-01T00:00:00Z",
      updatedAt: "2025-01-01T00:00:00Z",
    },
  };
}

async function openMenu(user: ReturnType<typeof userEvent.setup>, dashboardName: string) {
  await user.click(
    screen.getByRole("button", { name: new RegExp(`Dashboard actions for ${dashboardName}`, "i") }),
  );
}

describe("DashboardKebabMenu", () => {
  beforeEach(() => vi.clearAllMocks());

  it("shows Promote option for user-scoped dashboards (not Copy to mine)", async () => {
    const user = userEvent.setup();
    render(
      <MemoryRouter>
        <DashboardKebabMenu dashboard={dash("user")} isDefault={false} />
      </MemoryRouter>,
    );
    await openMenu(user, "Name d-1");
    expect(await screen.findByRole("menuitem", { name: /promote to household/i })).toBeInTheDocument();
    expect(screen.queryByRole("menuitem", { name: /copy to mine/i })).not.toBeInTheDocument();
  });

  it("shows Copy to mine option for household-scoped dashboards (not Promote)", async () => {
    const user = userEvent.setup();
    render(
      <MemoryRouter>
        <DashboardKebabMenu dashboard={dash("household")} isDefault={false} />
      </MemoryRouter>,
    );
    await openMenu(user, "Name d-1");
    expect(await screen.findByRole("menuitem", { name: /copy to mine/i })).toBeInTheDocument();
    expect(screen.queryByRole("menuitem", { name: /promote to household/i })).not.toBeInTheDocument();
  });

  it("disables Set as default when dashboard is already default", async () => {
    const user = userEvent.setup();
    render(
      <MemoryRouter>
        <DashboardKebabMenu dashboard={dash("household")} isDefault={true} />
      </MemoryRouter>,
    );
    await openMenu(user, "Name d-1");
    const item = await screen.findByRole("menuitem", { name: /already default/i });
    expect(item).toHaveAttribute("data-disabled");
  });

  it("enables Set as default when not default", async () => {
    const user = userEvent.setup();
    render(
      <MemoryRouter>
        <DashboardKebabMenu dashboard={dash("household")} isDefault={false} />
      </MemoryRouter>,
    );
    await openMenu(user, "Name d-1");
    const item = await screen.findByRole("menuitem", { name: /set as my default/i });
    expect(item).not.toHaveAttribute("data-disabled");
  });
});
