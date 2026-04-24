import { describe, it, expect, vi, beforeEach } from "vitest";
import { render } from "@testing-library/react";
import { MemoryRouter, Route, Routes, Outlet } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { Dashboard } from "@/types/models/dashboard";

vi.mock("@/context/tenant-context", () => ({
  useTenant: () => ({
    tenant: { id: "t1" },
    household: { id: "h1", attributes: { timezone: "UTC" } },
  }),
}));

// Capture the onRefresh callback passed to PullToRefresh so we can invoke it
// directly without simulating touch gestures.
let capturedOnRefresh: (() => Promise<void>) | null = null;
vi.mock("@/components/common/pull-to-refresh", () => ({
  PullToRefresh: ({
    onRefresh,
    children,
  }: {
    onRefresh: () => Promise<void>;
    children: React.ReactNode;
  }) => {
    capturedOnRefresh = onRefresh;
    return <>{children}</>;
  },
}));

// Avoid pulling in adapters' data hooks.
vi.mock("@/lib/dashboard/widget-registry", () => ({
  widgetRegistry: [],
  findWidget: () => undefined,
}));

import { DashboardRenderer } from "@/pages/DashboardRenderer";

function makeDashboard(): Dashboard {
  return {
    id: "dash-1",
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

function TestShell({ dashboard }: { dashboard: Dashboard }) {
  return <Outlet context={{ dashboard }} />;
}

describe("DashboardRenderer pull-to-refresh", () => {
  beforeEach(() => {
    capturedOnRefresh = null;
  });

  it("invalidates widget query keys when onRefresh fires", async () => {
    const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const invalidateSpy = vi.spyOn(qc, "invalidateQueries");

    render(
      <QueryClientProvider client={qc}>
        <MemoryRouter initialEntries={["/"]}>
          <Routes>
            <Route path="/" element={<TestShell dashboard={makeDashboard()} />}>
              <Route index element={<DashboardRenderer />} />
            </Route>
          </Routes>
        </MemoryRouter>
      </QueryClientProvider>,
    );

    expect(capturedOnRefresh).not.toBeNull();
    await capturedOnRefresh!();

    // Collect the first element of each invalidated queryKey — that is the
    // scope string exported by each key factory.
    const scopes = invalidateSpy.mock.calls.map(
      (call) => (call[0] as { queryKey: unknown[] }).queryKey[0] as string,
    );
    expect(scopes).toEqual(
      expect.arrayContaining([
        "tasks",
        "reminders",
        "packages",
        "meals",
        "calendar",
        "trackers",
        "workouts",
      ]),
    );
  });
});
