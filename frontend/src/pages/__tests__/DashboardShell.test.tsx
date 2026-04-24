import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter, Routes, Route } from "react-router-dom";
import type { Dashboard } from "@/types/models/dashboard";

const mockUseDashboard = vi.fn();
vi.mock("@/lib/hooks/api/use-dashboards", () => ({
  useDashboard: () => mockUseDashboard(),
}));

const mobileFlag = { value: false };
vi.mock("@/lib/hooks/use-mobile", () => ({
  useMobile: () => mobileFlag.value,
}));

import { DashboardShell } from "@/pages/DashboardShell";

function makeDashboard(): Dashboard {
  return {
    id: "d-1",
    type: "dashboards",
    attributes: {
      name: "Home",
      scope: "household",
      sortOrder: 0,
      schemaVersion: 1,
      createdAt: "2026-01-01T00:00:00Z",
      updatedAt: "2026-01-01T00:00:00Z",
      layout: { version: 1, widgets: [] },
    },
  };
}

function renderShell() {
  return render(
    <MemoryRouter initialEntries={["/app/dashboards/d-1"]}>
      <Routes>
        <Route path="/app/dashboards/:dashboardId" element={<DashboardShell />}>
          <Route index element={<div data-testid="child" />} />
        </Route>
      </Routes>
    </MemoryRouter>,
  );
}

beforeEach(() => {
  mockUseDashboard.mockReturnValue({
    data: { data: makeDashboard() },
    isLoading: false,
    isError: false,
  });
});

describe("DashboardShell Edit button", () => {
  it("renders enabled Edit link on desktop", () => {
    mobileFlag.value = false;
    renderShell();
    const link = screen.getByTestId("dashboard-shell-edit");
    expect(link).toBeInTheDocument();
    expect(link).not.toBeDisabled();
  });

  it("renders disabled Edit button with tooltip on mobile", () => {
    mobileFlag.value = true;
    renderShell();
    const btn = screen.getByTestId("dashboard-shell-edit-disabled");
    expect(btn).toBeInTheDocument();
    expect(btn).toBeDisabled();
    expect(btn).toHaveAttribute("title", "Editing requires a larger screen");
  });
});
