import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, within } from "@testing-library/react";
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

vi.mock("@/lib/hooks/api/use-dashboards", () => ({
  useDashboards: () => mockUseDashboards(),
}));

import { DashboardsNavGroup } from "../dashboards-nav-group";

describe("DashboardsNavGroup", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  function renderIt() {
    return render(
      <MemoryRouter>
        <DashboardsNavGroup isOpen={true} onToggle={() => {}} />
      </MemoryRouter>,
    );
  }

  it("renders household dashboards sorted, then user dashboards, as plain links", () => {
    mockUseDashboards.mockReturnValue({
      data: {
        data: [
          dash({ id: "hh-2", name: "Home B", scope: "household", sortOrder: 1 }),
          dash({ id: "u-1", name: "Mine A", scope: "user", sortOrder: 0 }),
          dash({ id: "hh-1", name: "Home A", scope: "household", sortOrder: 0 }),
        ],
      },
    });
    renderIt();

    const links = screen.getAllByRole("link");
    expect(links).toHaveLength(3);
    expect(links[0]).toHaveAttribute("href", "/app/dashboards/hh-1");
    expect(links[1]).toHaveAttribute("href", "/app/dashboards/hh-2");
    expect(links[2]).toHaveAttribute("href", "/app/dashboards/u-1");
    expect(screen.getByText("My Dashboards")).toBeInTheDocument();
  });

  it("renders no management affordances (no new-dashboard button, kebab, or grip)", () => {
    mockUseDashboards.mockReturnValue({
      data: { data: [dash({ id: "hh-1", name: "Home A", scope: "household", sortOrder: 0 })] },
    });
    renderIt();

    expect(screen.queryByRole("button", { name: /new dashboard/i })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /dashboard actions for/i })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /drag .* to reorder/i })).not.toBeInTheDocument();
  });

  it("renders no links and no new-dashboard button when the list is empty", () => {
    mockUseDashboards.mockReturnValue({ data: { data: [] } });
    renderIt();

    expect(screen.queryAllByRole("link")).toHaveLength(0);
    expect(screen.queryByRole("button", { name: /new dashboard/i })).not.toBeInTheDocument();
    expect(screen.queryByText("My Dashboards")).not.toBeInTheDocument();
  });

  it("omits the user section when there are no user dashboards", () => {
    mockUseDashboards.mockReturnValue({
      data: { data: [dash({ id: "hh-1", scope: "household", sortOrder: 0 })] },
    });
    const { container } = renderIt();
    expect(screen.queryByText("My Dashboards")).not.toBeInTheDocument();
    expect(within(container).getAllByRole("link")).toHaveLength(1);
  });
});
