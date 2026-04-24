import { describe, it, expect } from "vitest";
import { draftReducer, fromServer, type DraftState } from "@/pages/dashboard-designer/state";
import type { Dashboard } from "@/types/models/dashboard";
import type { WidgetInstance } from "@/lib/dashboard/schema";

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
      layout: {
        version: 1,
        widgets: [
          { id: "w-1", type: "weather", x: 0, y: 0, w: 4, h: 2, config: { units: "imperial" } },
          { id: "w-2", type: "tasks-summary", x: 4, y: 0, w: 4, h: 2, config: { status: "pending" } },
        ],
      },
    },
  };
}

function makeState(): DraftState {
  return fromServer(makeDashboard());
}

function newWidget(id = "w-new"): WidgetInstance {
  return { id, type: "habits-today", x: 0, y: 4, w: 4, h: 3, config: {} };
}

describe("draftReducer", () => {
  it("fromServer initializes non-dirty state with no selection", () => {
    const state = makeState();
    expect(state.dirty).toBe(false);
    expect(state.selectedWidgetId).toBeNull();
    expect(state.paletteOpen).toBe(false);
    expect(state.name).toBe("Home");
    expect(state.layout.widgets).toHaveLength(2);
  });

  it("add appends widget, selects it and marks dirty", () => {
    const s0 = makeState();
    const w = newWidget();
    const s1 = draftReducer(s0, { type: "add", widget: w });
    expect(s1.layout.widgets).toHaveLength(3);
    expect(s1.layout.widgets.at(-1)).toBe(w);
    expect(s1.dirty).toBe(true);
    expect(s1.selectedWidgetId).toBe("w-new");
  });

  it("remove filters out widget and marks dirty", () => {
    const s0 = makeState();
    const s1 = draftReducer(s0, { type: "remove", id: "w-1" });
    expect(s1.layout.widgets.map((w) => w.id)).toEqual(["w-2"]);
    expect(s1.dirty).toBe(true);
  });

  it("remove clears selectedWidgetId when removed id was selected", () => {
    const s0 = { ...makeState(), selectedWidgetId: "w-1" };
    const s1 = draftReducer(s0, { type: "remove", id: "w-1" });
    expect(s1.selectedWidgetId).toBeNull();
  });

  it("remove preserves selectedWidgetId when different", () => {
    const s0 = { ...makeState(), selectedWidgetId: "w-2" };
    const s1 = draftReducer(s0, { type: "remove", id: "w-1" });
    expect(s1.selectedWidgetId).toBe("w-2");
  });

  it("move-or-resize replaces widgets array and marks dirty", () => {
    const s0 = makeState();
    const next: WidgetInstance[] = [
      { id: "w-1", type: "weather", x: 6, y: 1, w: 4, h: 2, config: { units: "imperial" } },
      { id: "w-2", type: "tasks-summary", x: 0, y: 0, w: 4, h: 2, config: { status: "pending" } },
    ];
    const s1 = draftReducer(s0, { type: "move-or-resize", widgets: next });
    expect(s1.layout.widgets).toBe(next);
    expect(s1.dirty).toBe(true);
  });

  it("update-config only mutates matching widget and marks dirty", () => {
    const s0 = makeState();
    const s1 = draftReducer(s0, {
      type: "update-config",
      id: "w-2",
      config: { status: "overdue", title: "Overdue Tasks" },
    });
    expect(s1.layout.widgets[0]?.config).toEqual({ units: "imperial" });
    expect(s1.layout.widgets[1]?.config).toEqual({ status: "overdue", title: "Overdue Tasks" });
    expect(s1.dirty).toBe(true);
  });

  it("rename sets name and marks dirty", () => {
    const s0 = makeState();
    const s1 = draftReducer(s0, { type: "rename", name: "My Dash" });
    expect(s1.name).toBe("My Dash");
    expect(s1.dirty).toBe(true);
  });

  it("select does NOT mark dirty", () => {
    const s0 = makeState();
    const s1 = draftReducer(s0, { type: "select", id: "w-1" });
    expect(s1.selectedWidgetId).toBe("w-1");
    expect(s1.dirty).toBe(false);
  });

  it("select null clears selection", () => {
    const s0 = { ...makeState(), selectedWidgetId: "w-1" };
    const s1 = draftReducer(s0, { type: "select", id: null });
    expect(s1.selectedWidgetId).toBeNull();
    expect(s1.dirty).toBe(false);
  });

  it("toggle-palette does NOT mark dirty", () => {
    const s0 = makeState();
    const s1 = draftReducer(s0, { type: "toggle-palette", open: true });
    expect(s1.paletteOpen).toBe(true);
    expect(s1.dirty).toBe(false);
  });

  it("reset returns fromServer state with dirty cleared", () => {
    const s0 = { ...makeState(), dirty: true, name: "Changed", selectedWidgetId: "w-1" };
    const s1 = draftReducer(s0, { type: "reset", server: makeDashboard() });
    expect(s1.dirty).toBe(false);
    expect(s1.name).toBe("Home");
    expect(s1.selectedWidgetId).toBeNull();
  });

  it("saved returns fromServer state with dirty cleared", () => {
    const s0 = { ...makeState(), dirty: true, name: "Pending save" };
    const saved: Dashboard = {
      ...makeDashboard(),
      attributes: { ...makeDashboard().attributes, name: "Pending save" },
    };
    const s1 = draftReducer(s0, { type: "saved", server: saved });
    expect(s1.dirty).toBe(false);
    expect(s1.name).toBe("Pending save");
  });
});
