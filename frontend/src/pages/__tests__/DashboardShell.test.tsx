import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter, Routes, Route } from "react-router-dom";
import type { Dashboard } from "@/types/models/dashboard";

const mockUseDashboard = vi.fn();
vi.mock("@/lib/hooks/api/use-dashboards", () => ({
  useDashboard: () => mockUseDashboard(),
}));

vi.mock("@/context/tenant-context", () => ({
  useTenant: () => ({ household: { attributes: { timezone: "America/New_York" } } }),
}));

const authUser = { value: { attributes: { displayName: "Jordan Smith" } } as unknown };
vi.mock("@/components/providers/auth-provider", () => ({
  useAuth: () => ({ user: authUser.value }),
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

  it("shows visible explanatory text instead of an Edit button on mobile", () => {
    mobileFlag.value = true;
    renderShell();
    expect(screen.getByText("Editing needs a larger screen")).toBeVisible();
    expect(screen.getByTestId("dashboard-shell-edit-disabled")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Edit" })).not.toBeInTheDocument();
    expect(screen.queryByTestId("dashboard-shell-edit")).not.toBeInTheDocument();
  });
});

describe("DashboardShell header", () => {
  function atTime(iso: string, fn: () => void) {
    vi.useFakeTimers();
    vi.setSystemTime(new Date(iso));
    try {
      mobileFlag.value = false;
      fn();
    } finally {
      vi.useRealTimers();
    }
  }

  it("greets by household-local time of day with first name, date and dashboard name", () => {
    // 02:00Z on the 5th is 21:00 on the 4th in New York.
    atTime("2026-03-05T02:00:00Z", () => {
      renderShell();
      expect(
        screen.getByRole("heading", { level: 1, name: "Good evening, Jordan" }),
      ).toBeInTheDocument();
      expect(screen.getByText(/Wednesday, March 4/)).toBeInTheDocument();
      expect(screen.getByText("Home")).toBeInTheDocument();
    });
  });

  it("says good morning and afternoon at the right hours", () => {
    atTime("2026-03-05T14:00:00Z", () => {
      renderShell();
      expect(screen.getByRole("heading", { level: 1 })).toHaveTextContent("Good morning, Jordan");
    });
    atTime("2026-03-05T19:00:00Z", () => {
      renderShell();
      expect(screen.getAllByRole("heading", { level: 1 }).pop()).toHaveTextContent(
        "Good afternoon, Jordan",
      );
    });
  });

  it("omits the name when the user has none", () => {
    authUser.value = null;
    atTime("2026-03-05T14:00:00Z", () => {
      renderShell();
      expect(screen.getByRole("heading", { level: 1 })).toHaveTextContent(/^Good morning$/);
    });
    authUser.value = { attributes: { displayName: "Jordan Smith" } };
  });
});
