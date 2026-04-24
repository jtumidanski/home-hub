import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes, Outlet } from "react-router-dom";
import { z } from "zod";
import type { Dashboard } from "@/types/models/dashboard";

// Stub the registry so the renderer mounts simple marker components
// without pulling in each adapter's data hooks.
vi.mock("@/lib/dashboard/widget-registry", () => {
  const weatherDef = {
    type: "weather",
    displayName: "Weather",
    description: "",
    component: ({ config }: { config: { units: string } }) => (
      <div data-testid="stub-weather">units={config.units}</div>
    ),
    configSchema: z.object({ units: z.enum(["imperial", "metric"]) }),
    defaultConfig: { units: "imperial" as const },
    defaultSize: { w: 12, h: 3 },
    minSize: { w: 4, h: 2 },
    maxSize: { w: 12, h: 4 },
    dataScope: "household" as const,
  };
  const tasksDef = {
    type: "tasks-summary",
    displayName: "Tasks",
    description: "",
    component: ({ config }: { config: { status: string } }) => (
      <div data-testid="stub-tasks">status={config.status}</div>
    ),
    configSchema: z.object({ status: z.enum(["pending", "overdue", "completed"]) }),
    defaultConfig: { status: "pending" as const },
    defaultSize: { w: 4, h: 2 },
    minSize: { w: 2, h: 2 },
    maxSize: { w: 6, h: 3 },
    dataScope: "household" as const,
  };
  const registry = [weatherDef, tasksDef];
  return {
    widgetRegistry: registry,
    findWidget: (t: string) => registry.find((r) => r.type === t),
  };
});

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
      layout: {
        version: 1,
        widgets: [
          {
            id: "w-weather",
            type: "weather",
            x: 0,
            y: 0,
            w: 12,
            h: 3,
            config: { units: "metric" },
          },
          {
            id: "w-tasks",
            type: "tasks-summary",
            x: 0,
            y: 3,
            w: 4,
            h: 2,
            config: { status: "overdue" },
          },
          {
            id: "w-unknown",
            type: "future-widget",
            x: 4,
            y: 3,
            w: 4,
            h: 2,
            config: {},
          },
        ],
      },
    },
  };
}

function TestShell({ dashboard }: { dashboard: Dashboard }) {
  return <Outlet context={{ dashboard }} />;
}

function renderRenderer(dashboard: Dashboard) {
  return render(
    <MemoryRouter initialEntries={["/"]}>
      <Routes>
        <Route path="/" element={<TestShell dashboard={dashboard} />}>
          <Route index element={<DashboardRenderer />} />
        </Route>
      </Routes>
    </MemoryRouter>,
  );
}

describe("DashboardRenderer", () => {
  it("mounts known widgets with parsed config", () => {
    renderRenderer(makeDashboard());
    // Each widget is rendered twice (desktop grid + mobile stack)
    const weatherNodes = screen.getAllByTestId("stub-weather");
    expect(weatherNodes.length).toBe(2);
    expect(weatherNodes[0]).toHaveTextContent("units=metric");

    const tasksNodes = screen.getAllByTestId("stub-tasks");
    expect(tasksNodes.length).toBe(2);
    expect(tasksNodes[0]).toHaveTextContent("status=overdue");
  });

  it("renders UnknownWidgetPlaceholder for unregistered widget types", () => {
    renderRenderer(makeDashboard());
    // Two instances: desktop + mobile.
    expect(screen.getAllByText("Unknown widget").length).toBe(2);
    expect(screen.getAllByText("future-widget").length).toBe(2);
  });

  it("applies gridColumn/gridRow span styles to desktop slots", () => {
    const { container } = renderRenderer(makeDashboard());
    const grid = container.querySelector("[data-testid=dashboard-renderer-grid]") as HTMLElement;
    expect(grid).not.toBeNull();
    expect(grid.style.gridTemplateColumns).toContain("repeat(12");

    const slots = grid.querySelectorAll("[data-testid^=widget-slot-]");
    expect(slots.length).toBe(3);
    const first = slots[0] as HTMLElement;
    expect(first.style.gridColumn).toBe("span 12");
    expect(first.style.gridRow).toBe("span 3");
  });

  it("sorts widgets by y then x", () => {
    const d = makeDashboard();
    // shuffle order in persisted layout
    d.attributes.layout.widgets.reverse();
    const { container } = renderRenderer(d);
    const grid = container.querySelector("[data-testid=dashboard-renderer-grid]") as HTMLElement;
    const slots = Array.from(grid.querySelectorAll("[data-testid^=widget-slot-]"));
    // Expect weather first (y=0), then tasks (y=3,x=0), then unknown (y=3,x=4)
    expect(slots[0].getAttribute("data-testid")).toBe("widget-slot-w-weather");
    expect(slots[1].getAttribute("data-testid")).toBe("widget-slot-w-tasks");
    expect(slots[2].getAttribute("data-testid")).toBe("widget-slot-w-unknown");
  });
});
