import { describe, it, expect } from "vitest";
import type { Dashboard } from "@/types/models/dashboard";
import { computeReorderEntries, sortDashboards } from "../ordering";

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

describe("sortDashboards", () => {
  it("sorts by sortOrder ASC then createdAt ASC", () => {
    const list = [
      dash({ id: "c", sortOrder: 1, createdAt: "2025-01-02T00:00:00Z" }),
      dash({ id: "a", sortOrder: 0, createdAt: "2025-01-02T00:00:00Z" }),
      dash({ id: "b", sortOrder: 0, createdAt: "2025-01-01T00:00:00Z" }),
    ];
    expect(sortDashboards(list).map((d) => d.id)).toEqual(["b", "a", "c"]);
  });
});

describe("computeReorderEntries", () => {
  const sorted = [
    dash({ id: "a", sortOrder: 0 }),
    dash({ id: "b", sortOrder: 1 }),
    dash({ id: "c", sortOrder: 2 }),
  ];

  it("returns null when active === over", () => {
    expect(computeReorderEntries(sorted, "a", "a")).toBeNull();
  });

  it("emits 0-indexed sortOrder after moving a down", () => {
    expect(computeReorderEntries(sorted, "a", "c")).toEqual([
      { id: "b", sortOrder: 0 },
      { id: "c", sortOrder: 1 },
      { id: "a", sortOrder: 2 },
    ]);
  });

  it("emits 0-indexed sortOrder after moving c up", () => {
    expect(computeReorderEntries(sorted, "c", "a")).toEqual([
      { id: "c", sortOrder: 0 },
      { id: "a", sortOrder: 1 },
      { id: "b", sortOrder: 2 },
    ]);
  });

  it("returns null for unknown id", () => {
    expect(computeReorderEntries(sorted, "zzz", "a")).toBeNull();
  });
});
