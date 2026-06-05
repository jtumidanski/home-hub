import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import type { Dashboard } from "@/types/models/dashboard";

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

import { DashboardRow } from "../dashboard-row";

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

function renderRow(d: Dashboard, defaultDashboardId: string | null = null) {
  return render(
    <MemoryRouter>
      <DashboardRow dashboard={d} defaultDashboardId={defaultDashboardId} />
    </MemoryRouter>,
  );
}

describe("DashboardRow", () => {
  beforeEach(() => vi.clearAllMocks());

  it("links the name to the dashboard and exposes an edit link to the designer", () => {
    renderRow(dash({ id: "d-1", name: "Home" }));
    expect(screen.getByRole("link", { name: /^home$/i })).toHaveAttribute("href", "/app/dashboards/d-1");
    expect(screen.getByRole("link", { name: /edit home/i })).toHaveAttribute("href", "/app/dashboards/d-1/edit");
  });

  it("renders the kebab actions trigger", () => {
    renderRow(dash({ id: "d-1", name: "Home" }));
    expect(screen.getByRole("button", { name: /dashboard actions for home/i })).toBeInTheDocument();
  });

  it("renders the grip with an accessible label", () => {
    renderRow(dash({ id: "d-1", name: "Home" }));
    expect(screen.getByRole("button", { name: /drag home to reorder/i })).toBeInTheDocument();
  });

  it("shows a Default badge when the row is the current default", () => {
    renderRow(dash({ id: "d-1", name: "Home" }), "d-1");
    expect(screen.getByText(/default/i)).toBeInTheDocument();
  });

  it("hides the Default badge when the row is not the default", () => {
    renderRow(dash({ id: "d-1", name: "Home" }), "other");
    expect(screen.queryByText(/^default$/i)).not.toBeInTheDocument();
  });
});
