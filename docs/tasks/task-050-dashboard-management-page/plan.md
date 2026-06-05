# Dashboard Management Page Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Relocate every dashboard-management affordance out of the sidebar's `DashboardsNavGroup` into a dedicated, mobile-friendly Dashboard Management page rendered at `/app/dashboards`, leaving the sidebar group navigation-only.

**Architecture:** Frontend-only **move + extract** refactor. Pure ordering helpers (`sortDashboards`, `computeReorderEntries`) move to a neutral module. A new `DashboardRow` component (extracted from the sidebar's `SortableDashboardRow`) hosts the reused `DashboardKebabMenu` (trigger restyled to always-visible) on a full-width card. A new `DashboardManagementPage` composes two per-scope `DndContext`s plus the reused `NewDashboardModal`. The sidebar group is stripped to plain `NavLink`s, the route is repointed, and the obsolete `DashboardsIndexRedirect` is deleted.

**Tech Stack:** React + TypeScript, Vite, react-router-dom, @dnd-kit (core + sortable), TanStack React Query, react-hook-form + Zod, Tailwind, shadcn/ui (base-ui primitives), Vitest + Testing Library.

---

## File Structure

**New files:**
- `frontend/src/lib/dashboard/ordering.ts` — pure `sortDashboards` + `computeReorderEntries`.
- `frontend/src/lib/dashboard/__tests__/ordering.test.ts` — relocated unit tests for the helpers.
- `frontend/src/components/features/dashboards/dashboard-row.tsx` — one draggable, full-width row (grip + name link + edit link + Default badge + kebab).
- `frontend/src/components/features/dashboards/__tests__/dashboard-row.test.tsx` — row render tests.
- `frontend/src/pages/DashboardManagementPage.tsx` — the page (header, per-scope sections, modal host, loading/error/empty states).
- `frontend/src/pages/__tests__/DashboardManagementPage.test.tsx` — page tests.

**Modified files:**
- `frontend/src/components/features/dashboards/dashboard-kebab-menu.tsx` — trigger restyled to always-visible (drop hover-gated opacity classes).
- `frontend/src/components/features/navigation/dashboards-nav-group.tsx` — strip drag/kebab/new; plain `NavLink` rows; import `sortDashboards` from the new module.
- `frontend/src/components/features/navigation/__tests__/dashboards-nav-group.test.tsx` — drop the moved helper `describe` blocks; rewrite `DashboardsNavGroup` cases for nav-only behavior.
- `frontend/src/components/features/navigation/nav-config.ts` — add "Manage Dashboards" to the `management` group.
- `frontend/src/App.tsx` — repoint `/app/dashboards` to `DashboardManagementPage`; remove `DashboardsIndexRedirect` import.

**Deleted files:**
- `frontend/src/pages/DashboardsIndexRedirect.tsx` — only `App.tsx` imports it (verified); `DashboardRedirect` is unaffected.

---

## Task 1: Move ordering helpers to a neutral module

Move the two pure functions out of the navigation component (which stops reordering after this task) into `lib/dashboard/ordering.ts`, and relocate their unit tests. The nav group keeps working by importing from the new module.

**Files:**
- Create: `frontend/src/lib/dashboard/ordering.ts`
- Create: `frontend/src/lib/dashboard/__tests__/ordering.test.ts`
- Modify: `frontend/src/components/features/navigation/dashboards-nav-group.tsx` (remove local helper definitions, import from new module)
- Modify: `frontend/src/components/features/navigation/__tests__/dashboards-nav-group.test.tsx` (remove the two moved `describe` blocks + their helper import)

- [ ] **Step 1: Write the relocated helper tests (failing — module does not exist yet)**

Create `frontend/src/lib/dashboard/__tests__/ordering.test.ts`:

```typescript
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
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd frontend && npx vitest run src/lib/dashboard/__tests__/ordering.test.ts`
Expected: FAIL — cannot resolve import `../ordering`.

- [ ] **Step 3: Create the ordering module**

Create `frontend/src/lib/dashboard/ordering.ts`:

```typescript
import type { Dashboard } from "@/types/models/dashboard";

/**
 * Sort dashboards by sortOrder ASC, then createdAt ASC as a stable tiebreaker.
 */
export function sortDashboards(list: Dashboard[]): Dashboard[] {
  return [...list].sort((a, b) => {
    if (a.attributes.sortOrder !== b.attributes.sortOrder) {
      return a.attributes.sortOrder - b.attributes.sortOrder;
    }
    return a.attributes.createdAt.localeCompare(b.attributes.createdAt);
  });
}

/**
 * Given a sorted list and active/over ids from a dnd-kit drag-end, returns
 * the reorder payload with 0-indexed sortOrder. Returns null when the drag
 * is a no-op.
 */
export function computeReorderEntries(
  sorted: Dashboard[],
  activeId: string,
  overId: string,
): Array<{ id: string; sortOrder: number }> | null {
  if (activeId === overId) return null;
  const fromIdx = sorted.findIndex((d) => d.id === activeId);
  const toIdx = sorted.findIndex((d) => d.id === overId);
  if (fromIdx < 0 || toIdx < 0) return null;
  const next = [...sorted];
  const [moved] = next.splice(fromIdx, 1);
  if (!moved) return null;
  next.splice(toIdx, 0, moved);
  return next.map((d, i) => ({ id: d.id, sortOrder: i }));
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `cd frontend && npx vitest run src/lib/dashboard/__tests__/ordering.test.ts`
Expected: PASS (5 tests).

- [ ] **Step 5: Update `dashboards-nav-group.tsx` to import from the new module**

In `frontend/src/components/features/navigation/dashboards-nav-group.tsx`:

Delete the two local function definitions (`sortDashboards` and `computeReorderEntries`, lines 28–59) entirely.

Add an import near the other `@/lib` imports (after the `cn` import on line 21):

```typescript
import { sortDashboards, computeReorderEntries } from "@/lib/dashboard/ordering";
```

(The component body still calls both functions unchanged; only their source moved.)

- [ ] **Step 6: Remove the moved `describe` blocks from the nav-group test**

In `frontend/src/components/features/navigation/__tests__/dashboards-nav-group.test.tsx`:

Change the helper import on line 5 from:

```typescript
import { computeReorderEntries, sortDashboards } from "../dashboards-nav-group";
```

to remove it entirely (the `DashboardsNavGroup` import on line 44 stays).

Delete the entire `describe("sortDashboards", ...)` block (lines 46–55) and the entire `describe("computeReorderEntries", ...)` block (lines 57–87). Leave the `dash` helper, the mocks, and the `describe("DashboardsNavGroup", ...)` block intact (that block is rewritten later in Task 6).

- [ ] **Step 7: Run both affected suites and type-check**

Run: `cd frontend && npx vitest run src/lib/dashboard/__tests__/ordering.test.ts src/components/features/navigation/__tests__/dashboards-nav-group.test.tsx`
Expected: PASS (ordering: 5; nav-group `DashboardsNavGroup`: 3 — still the original sidebar assertions, unchanged at this point).

Run: `cd frontend && npx tsc -b`
Expected: no errors.

- [ ] **Step 8: Commit**

```bash
git add frontend/src/lib/dashboard/ordering.ts \
        frontend/src/lib/dashboard/__tests__/ordering.test.ts \
        frontend/src/components/features/navigation/dashboards-nav-group.tsx \
        frontend/src/components/features/navigation/__tests__/dashboards-nav-group.test.tsx
git commit -m "refactor(frontend): move dashboard ordering helpers to lib/dashboard/ordering"
```

---

## Task 2: Restyle the kebab menu trigger to always-visible

The page must operate without hover (FR-5). The reused `DashboardKebabMenu` trigger is currently hover-gated. Drop the opacity classes so the trigger is always visible. The sidebar drops the kebab in Task 6, so this restyle is effectively page-only.

**Files:**
- Modify: `frontend/src/components/features/dashboards/dashboard-kebab-menu.tsx:93-95`

The existing test (`__tests__/dashboard-kebab-menu.test.tsx`) opens the menu via its accessible label and asserts nothing about opacity, so it remains green — no test change needed.

- [ ] **Step 1: Edit the trigger className**

In `frontend/src/components/features/dashboards/dashboard-kebab-menu.tsx`, replace the trigger `className` (lines 93–95):

```typescript
          className={cn(
            "flex h-7 w-7 shrink-0 items-center justify-center rounded-md text-muted-foreground opacity-0 transition-opacity hover:bg-sidebar-accent/50 hover:text-sidebar-foreground focus-visible:opacity-100 group-hover/row:opacity-100 data-[popup-open]:opacity-100 outline-none",
          )}
```

with (drop `opacity-0`, `transition-opacity`, `focus-visible:opacity-100`, `group-hover/row:opacity-100`, `data-[popup-open]:opacity-100`):

```typescript
          className={cn(
            "flex h-7 w-7 shrink-0 items-center justify-center rounded-md text-muted-foreground hover:bg-accent hover:text-accent-foreground outline-none",
          )}
```

- [ ] **Step 2: Run the kebab test to confirm no regression**

Run: `cd frontend && npx vitest run src/components/features/dashboards/__tests__/dashboard-kebab-menu.test.tsx`
Expected: PASS (4 tests).

- [ ] **Step 3: Type-check**

Run: `cd frontend && npx tsc -b`
Expected: no errors.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/features/dashboards/dashboard-kebab-menu.tsx
git commit -m "refactor(frontend): make dashboard kebab trigger always-visible for touch"
```

---

## Task 3: Create the `DashboardRow` component

A full-width, draggable card row extracted from the sidebar's `SortableDashboardRow`: grip + name link (open) + edit link (designer) + Default badge + the reused kebab. Props mirror the sidebar's, minus sidebar-only sizing.

**Files:**
- Create: `frontend/src/components/features/dashboards/dashboard-row.tsx`
- Create: `frontend/src/components/features/dashboards/__tests__/dashboard-row.test.tsx`

- [ ] **Step 1: Write the failing row test**

Create `frontend/src/components/features/dashboards/__tests__/dashboard-row.test.tsx`:

```typescript
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
    expect(screen.getByRole("link", { name: /home/i })).toHaveAttribute("href", "/app/dashboards/d-1");
    expect(screen.getByRole("link", { name: /edit d-1/i })).toHaveAttribute("href", "/app/dashboards/d-1/edit");
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
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd frontend && npx vitest run src/components/features/dashboards/__tests__/dashboard-row.test.tsx`
Expected: FAIL — cannot resolve import `../dashboard-row`.

- [ ] **Step 3: Create the `DashboardRow` component**

Create `frontend/src/components/features/dashboards/dashboard-row.tsx`:

```typescript
import { Link } from "react-router-dom";
import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { GripVertical, LayoutDashboard, Pencil } from "lucide-react";
import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";
import { DashboardKebabMenu } from "@/components/features/dashboards/dashboard-kebab-menu";
import type { Dashboard } from "@/types/models/dashboard";

interface DashboardRowProps {
  dashboard: Dashboard;
  defaultDashboardId: string | null;
}

export function DashboardRow({ dashboard, defaultDashboardId }: DashboardRowProps) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } =
    useSortable({ id: dashboard.id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };

  const isDefault = defaultDashboardId === dashboard.id;

  return (
    <div
      ref={setNodeRef}
      style={style}
      className="flex items-center gap-2 rounded-md border bg-card p-3"
    >
      <button
        type="button"
        aria-label={`Drag ${dashboard.attributes.name} to reorder`}
        className="flex h-9 w-9 shrink-0 cursor-grab touch-none items-center justify-center rounded-md text-muted-foreground hover:bg-accent hover:text-accent-foreground outline-none"
        {...attributes}
        {...listeners}
      >
        <GripVertical className="h-4 w-4" />
      </button>

      <LayoutDashboard className="h-4 w-4 shrink-0 text-muted-foreground" />

      <Link
        to={`/app/dashboards/${dashboard.id}`}
        className="min-w-0 flex-1 truncate text-sm font-medium hover:underline"
      >
        {dashboard.attributes.name}
      </Link>

      {isDefault && (
        <Badge variant="secondary" className="shrink-0">
          Default
        </Badge>
      )}

      <Link
        to={`/app/dashboards/${dashboard.id}/edit`}
        aria-label={`Edit ${dashboard.attributes.name}`}
        className={cn(
          "flex h-9 w-9 shrink-0 items-center justify-center rounded-md text-muted-foreground",
          "hover:bg-accent hover:text-accent-foreground outline-none",
        )}
      >
        <Pencil className="h-4 w-4" />
      </Link>

      <DashboardKebabMenu dashboard={dashboard} isDefault={isDefault} />
    </div>
  );
}
```

- [ ] **Step 4: Run the row test to verify it passes**

Run: `cd frontend && npx vitest run src/components/features/dashboards/__tests__/dashboard-row.test.tsx`
Expected: PASS (5 tests).

> Note: `useSortable` works outside a `SortableContext` (it falls back to a no-op draggable), so the row renders fine in isolation for the unit test.

- [ ] **Step 5: Type-check**

Run: `cd frontend && npx tsc -b`
Expected: no errors.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/components/features/dashboards/dashboard-row.tsx \
        frontend/src/components/features/dashboards/__tests__/dashboard-row.test.tsx
git commit -m "feat(frontend): add full-width DashboardRow for the management page"
```

---

## Task 4: Create the `DashboardManagementPage`

The page composes the header + "New Dashboard" button, loading/error/empty states, and two per-scope `DndContext`/`SortableContext` sections rendering `DashboardRow`s. Reorder issues one `useReorderDashboards.mutate` call per scope via `computeReorderEntries`.

**Files:**
- Create: `frontend/src/pages/DashboardManagementPage.tsx`
- Create: `frontend/src/pages/__tests__/DashboardManagementPage.test.tsx`

- [ ] **Step 1: Write the failing page test**

Create `frontend/src/pages/__tests__/DashboardManagementPage.test.tsx`:

```typescript
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
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
const mockUseHouseholdPreferences = vi.fn();

vi.mock("@/lib/hooks/api/use-dashboards", () => ({
  useDashboards: () => mockUseDashboards(),
  useReorderDashboards: () => ({ mutate: vi.fn() }),
  useCreateDashboard: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useCopyDashboardToMine: () => ({ mutate: vi.fn(), mutateAsync: vi.fn(), isPending: false }),
  useUpdateDashboard: () => ({ mutate: vi.fn(), isPending: false }),
  useDeleteDashboard: () => ({ mutate: vi.fn(), isPending: false }),
  usePromoteDashboard: () => ({ mutate: vi.fn() }),
}));

vi.mock("@/lib/hooks/api/use-household-preferences", () => ({
  useHouseholdPreferences: () => mockUseHouseholdPreferences(),
  useUpdateHouseholdPreferences: () => ({ mutate: vi.fn() }),
}));

import { DashboardManagementPage } from "../DashboardManagementPage";

function renderPage() {
  return render(
    <MemoryRouter>
      <DashboardManagementPage />
    </MemoryRouter>,
  );
}

describe("DashboardManagementPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseHouseholdPreferences.mockReturnValue({
      data: { data: { id: "prefs-1", type: "householdPreferences", attributes: { defaultDashboardId: null, kioskDashboardSeeded: false, createdAt: "", updatedAt: "" } } },
    });
    mockUseDashboards.mockReturnValue({ data: null, isLoading: false, isError: false });
  });

  it("renders a loading skeleton while dashboards load", () => {
    mockUseDashboards.mockReturnValue({ data: null, isLoading: true, isError: false });
    renderPage();
    expect(screen.getByRole("status", { name: "Loading" })).toBeInTheDocument();
  });

  it("renders an error state on failure", () => {
    mockUseDashboards.mockReturnValue({ data: null, isLoading: false, isError: true });
    renderPage();
    expect(screen.getByText(/failed to load dashboards/i)).toBeInTheDocument();
  });

  it("renders household and user sections, each sorted, and an open link per row", () => {
    mockUseDashboards.mockReturnValue({
      data: {
        data: [
          dash({ id: "hh-2", name: "Home B", scope: "household", sortOrder: 1 }),
          dash({ id: "u-1", name: "Mine A", scope: "user", sortOrder: 0 }),
          dash({ id: "hh-1", name: "Home A", scope: "household", sortOrder: 0 }),
        ],
      },
      isLoading: false,
      isError: false,
    });
    renderPage();

    expect(screen.getByText("Household Dashboards")).toBeInTheDocument();
    expect(screen.getByText("My Dashboards")).toBeInTheDocument();

    const openLinks = screen
      .getAllByRole("link")
      .filter((a) => /^\/app\/dashboards\/[^/]+$/.test(a.getAttribute("href") ?? ""));
    expect(openLinks.map((a) => a.getAttribute("href"))).toEqual([
      "/app/dashboards/hh-1",
      "/app/dashboards/hh-2",
      "/app/dashboards/u-1",
    ]);
  });

  it("omits a scope section that has no dashboards", () => {
    mockUseDashboards.mockReturnValue({
      data: { data: [dash({ id: "hh-1", scope: "household", sortOrder: 0 })] },
      isLoading: false,
      isError: false,
    });
    renderPage();
    expect(screen.getByText("Household Dashboards")).toBeInTheDocument();
    expect(screen.queryByText("My Dashboards")).not.toBeInTheDocument();
  });

  it("renders a page-level empty state when there are no dashboards", () => {
    mockUseDashboards.mockReturnValue({ data: { data: [] }, isLoading: false, isError: false });
    renderPage();
    expect(screen.getByText(/no dashboards yet/i)).toBeInTheDocument();
    expect(screen.queryByText("Household Dashboards")).not.toBeInTheDocument();
    expect(screen.queryByText("My Dashboards")).not.toBeInTheDocument();
  });

  it("opens the New Dashboard modal when the header button is clicked", async () => {
    const user = userEvent.setup();
    mockUseDashboards.mockReturnValue({ data: { data: [] }, isLoading: false, isError: false });
    renderPage();

    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: /new dashboard/i }));
    expect(screen.getByRole("dialog")).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd frontend && npx vitest run src/pages/__tests__/DashboardManagementPage.test.tsx`
Expected: FAIL — cannot resolve import `../DashboardManagementPage`.

- [ ] **Step 3: Create the page component**

Create `frontend/src/pages/DashboardManagementPage.tsx`:

```typescript
import { useMemo, useState } from "react";
import {
  DndContext,
  KeyboardSensor,
  PointerSensor,
  closestCenter,
  useSensor,
  useSensors,
  type DragEndEvent,
} from "@dnd-kit/core";
import {
  SortableContext,
  sortableKeyboardCoordinates,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { Plus } from "lucide-react";
import { useDashboards, useReorderDashboards } from "@/lib/hooks/api/use-dashboards";
import { useHouseholdPreferences } from "@/lib/hooks/api/use-household-preferences";
import { sortDashboards, computeReorderEntries } from "@/lib/dashboard/ordering";
import { DashboardRow } from "@/components/features/dashboards/dashboard-row";
import { NewDashboardModal } from "@/components/features/dashboards/new-dashboard-modal";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { ErrorCard } from "@/components/common/error-card";
import type { Dashboard } from "@/types/models/dashboard";

export function DashboardManagementPage() {
  const { data, isLoading, isError } = useDashboards();
  const { data: prefsData } = useHouseholdPreferences();
  const reorderMutation = useReorderDashboards();
  const [modalOpen, setModalOpen] = useState(false);

  const dashboards = useMemo(() => data?.data ?? [], [data]);

  const { householdList, userList } = useMemo(() => {
    const household = sortDashboards(
      dashboards.filter((d) => d.attributes.scope === "household"),
    );
    const user = sortDashboards(
      dashboards.filter((d) => d.attributes.scope === "user"),
    );
    return { householdList: household, userList: user };
  }, [dashboards]);

  const defaultDashboardId =
    prefsData?.data?.[0]?.attributes.defaultDashboardId ?? null;

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 4 } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates }),
  );

  // One reorder call per scope — the backend rejects mixed-scope batches.
  const handleDragEnd = (scope: "household" | "user") => (event: DragEndEvent) => {
    const { active, over } = event;
    if (!over) return;
    const list = scope === "household" ? householdList : userList;
    const entries = computeReorderEntries(list, String(active.id), String(over.id));
    if (!entries) return;
    reorderMutation.mutate(entries);
  };

  const renderSection = (
    title: string,
    list: Dashboard[],
    scope: "household" | "user",
  ) => (
    <section className="space-y-2">
      <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">
        {title}
      </h2>
      <DndContext
        sensors={sensors}
        collisionDetection={closestCenter}
        onDragEnd={handleDragEnd(scope)}
      >
        <SortableContext items={list.map((d) => d.id)} strategy={verticalListSortingStrategy}>
          <div className="space-y-2">
            {list.map((dashboard) => (
              <DashboardRow
                key={dashboard.id}
                dashboard={dashboard}
                defaultDashboardId={defaultDashboardId}
              />
            ))}
          </div>
        </SortableContext>
      </DndContext>
    </section>
  );

  if (isLoading) {
    return (
      <div className="p-4 md:p-6 space-y-4" role="status" aria-label="Loading">
        {Array.from({ length: 3 }).map((_, i) => (
          <Skeleton key={i} className="h-16" />
        ))}
      </div>
    );
  }

  if (isError) {
    return (
      <div className="p-4 md:p-6">
        <ErrorCard message="Failed to load dashboards. Try refreshing the page." />
      </div>
    );
  }

  const isEmpty = householdList.length === 0 && userList.length === 0;

  return (
    <div className="p-4 md:p-6 space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-xl md:text-2xl font-semibold">Dashboards</h1>
        <Button size="sm" onClick={() => setModalOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />New Dashboard
        </Button>
      </div>

      <NewDashboardModal open={modalOpen} onOpenChange={setModalOpen} />

      {isEmpty ? (
        <div className="flex flex-col items-center justify-center py-12 text-center space-y-4">
          <p className="text-muted-foreground">No dashboards yet.</p>
          <Button variant="outline" onClick={() => setModalOpen(true)}>
            <Plus className="mr-2 h-4 w-4" />Create First Dashboard
          </Button>
        </div>
      ) : (
        <div className="space-y-6">
          {householdList.length > 0 &&
            renderSection("Household Dashboards", householdList, "household")}
          {userList.length > 0 && renderSection("My Dashboards", userList, "user")}
        </div>
      )}
    </div>
  );
}
```

- [ ] **Step 4: Run the page test to verify it passes**

Run: `cd frontend && npx vitest run src/pages/__tests__/DashboardManagementPage.test.tsx`
Expected: PASS (6 tests).

- [ ] **Step 5: Type-check**

Run: `cd frontend && npx tsc -b`
Expected: no errors.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/pages/DashboardManagementPage.tsx \
        frontend/src/pages/__tests__/DashboardManagementPage.test.tsx
git commit -m "feat(frontend): add DashboardManagementPage"
```

---

## Task 5: Wire routing, nav entry point, and delete the obsolete redirect

Repoint `/app/dashboards` to the new page, add the "Manage Dashboards" entry to the `management` nav group, and delete `DashboardsIndexRedirect` (only `App.tsx` imports it).

**Files:**
- Modify: `frontend/src/components/features/navigation/nav-config.ts`
- Modify: `frontend/src/App.tsx`
- Delete: `frontend/src/pages/DashboardsIndexRedirect.tsx`

- [ ] **Step 1: Add the "Manage Dashboards" nav item**

In `frontend/src/components/features/navigation/nav-config.ts`:

Add `LayoutGrid` to the `lucide-react` import on line 1:

```typescript
import { Home, CheckSquare, Bell, Calendar, Package, CloudSun, UtensilsCrossed, Carrot, CalendarDays, ShoppingCart, Heart, Settings, Target, Dumbbell, LayoutGrid, type LucideIcon } from "lucide-react";
```

Update the `management` group's `items` (lines 56–59) to add the entry before Households:

```typescript
    items: [
      { to: "/app/dashboards", icon: LayoutGrid, label: "Manage Dashboards" },
      { to: "/app/households", icon: Home, label: "Households", badgeKey: "pendingInvitationCount" },
    ],
```

- [ ] **Step 2: Repoint the route in `App.tsx` and drop the redirect import**

In `frontend/src/App.tsx`:

Remove the import on line 16:

```typescript
import { DashboardsIndexRedirect } from "@/pages/DashboardsIndexRedirect";
```

Add an eager import for the new page alongside the other page imports (e.g. after the `DashboardRedirect` import on line 15):

```typescript
import { DashboardManagementPage } from "@/pages/DashboardManagementPage";
```

Change the route on line 78 from:

```typescript
                  <Route path="dashboards" element={<DashboardsIndexRedirect />} />
```

to:

```typescript
                  <Route path="dashboards" element={<DashboardManagementPage />} />
```

(Leave the `index`, `dashboard`, `dashboards/:dashboardId`, and `:dashboardId/edit` routes untouched.)

- [ ] **Step 3: Delete the obsolete redirect file**

Run: `cd frontend && rm src/pages/DashboardsIndexRedirect.tsx`

- [ ] **Step 4: Verify no dangling references remain**

Run: `cd frontend && grep -rn "DashboardsIndexRedirect" src`
Expected: no output (zero matches).

- [ ] **Step 5: Type-check and run the existing redirect test (must stay green)**

Run: `cd frontend && npx tsc -b`
Expected: no errors.

Run: `cd frontend && npx vitest run src/pages/__tests__/DashboardRedirect.test.tsx`
Expected: PASS (the `DashboardRedirect` component is unaffected).

- [ ] **Step 6: Commit**

```bash
git add frontend/src/components/features/navigation/nav-config.ts frontend/src/App.tsx
git rm frontend/src/pages/DashboardsIndexRedirect.tsx
git commit -m "feat(frontend): route /app/dashboards to management page; add nav entry"
```

---

## Task 6: Strip `DashboardsNavGroup` to navigation-only

Remove the drag/kebab/new-dashboard machinery; render each dashboard as a plain `NavLink` (icon, truncation, active highlight, mobile drawer close intact). Rewrite the component's tests for the new behavior.

**Files:**
- Modify: `frontend/src/components/features/navigation/dashboards-nav-group.tsx`
- Modify: `frontend/src/components/features/navigation/__tests__/dashboards-nav-group.test.tsx`

- [ ] **Step 1: Rewrite the `DashboardsNavGroup` test cases**

In `frontend/src/components/features/navigation/__tests__/dashboards-nav-group.test.tsx`, replace the entire `describe("DashboardsNavGroup", ...)` block with the nav-only assertions below. Also simplify the `vi.mock("@/lib/hooks/api/use-dashboards", ...)` factory (lines 29–37) to only what the nav group still imports — `useDashboards` — and drop the `use-household-preferences` mock (the group no longer reads prefs).

Replace the mock factories (the two `vi.mock(...)` blocks for `use-dashboards` and `use-household-preferences`) with a single mock:

```typescript
const mockUseDashboards = vi.fn();

vi.mock("@/lib/hooks/api/use-dashboards", () => ({
  useDashboards: () => mockUseDashboards(),
}));
```

(Delete the now-unused `mockUseHouseholdPreferences` and `mockReorderMutate` declarations.)

Replace the `describe("DashboardsNavGroup", ...)` block with:

```typescript
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
```

- [ ] **Step 2: Run the test to verify it fails against the current component**

Run: `cd frontend && npx vitest run src/components/features/navigation/__tests__/dashboards-nav-group.test.tsx`
Expected: FAIL — the current component still renders the "New Dashboard" button, kebab triggers, and grips.

- [ ] **Step 3: Rewrite the component to navigation-only**

Replace the entire contents of `frontend/src/components/features/navigation/dashboards-nav-group.tsx` with:

```typescript
import { useMemo } from "react";
import { NavLink } from "react-router-dom";
import { Collapsible as CollapsiblePrimitive } from "@base-ui/react/collapsible";
import { ChevronRight, LayoutDashboard } from "lucide-react";
import { cn } from "@/lib/utils";
import { useDashboards } from "@/lib/hooks/api/use-dashboards";
import { sortDashboards } from "@/lib/dashboard/ordering";
import type { Dashboard } from "@/types/models/dashboard";

interface DashboardsNavGroupProps {
  isOpen: boolean;
  onToggle: () => void;
  onItemClick?: () => void;
  iconSize?: string;
  itemPadding?: string;
}

export function DashboardsNavGroup({
  isOpen,
  onToggle,
  onItemClick,
  iconSize = "h-4 w-4",
  itemPadding = "py-2",
}: DashboardsNavGroupProps) {
  const { data } = useDashboards();

  const dashboards = data?.data ?? [];

  const { householdList, userList } = useMemo(() => {
    const household = sortDashboards(
      dashboards.filter((d) => d.attributes.scope === "household"),
    );
    const user = sortDashboards(
      dashboards.filter((d) => d.attributes.scope === "user"),
    );
    return { householdList: household, userList: user };
  }, [dashboards]);

  const renderLink = (dashboard: Dashboard) => (
    <NavLink
      key={dashboard.id}
      to={`/app/dashboards/${dashboard.id}`}
      onClick={onItemClick}
      className={({ isActive }) =>
        cn(
          "flex items-center gap-3 rounded-md px-3 text-sm font-medium transition-colors",
          itemPadding,
          isActive
            ? "bg-sidebar-accent text-sidebar-accent-foreground"
            : "text-sidebar-foreground hover:bg-sidebar-accent/50",
        )
      }
    >
      <LayoutDashboard className={iconSize} />
      <span className="flex-1 truncate">{dashboard.attributes.name}</span>
    </NavLink>
  );

  return (
    <CollapsiblePrimitive.Root open={isOpen} onOpenChange={() => onToggle()}>
      <CollapsiblePrimitive.Trigger
        className={cn(
          "flex w-full items-center gap-2 rounded-md px-3 py-1.5 text-xs font-semibold uppercase tracking-wider text-muted-foreground transition-colors hover:text-sidebar-foreground",
        )}
      >
        <ChevronRight
          className={cn(
            "h-3 w-3 transition-transform duration-200",
            isOpen && "rotate-90",
          )}
        />
        Dashboards
      </CollapsiblePrimitive.Trigger>
      <CollapsiblePrimitive.Panel className="overflow-hidden transition-all duration-200 data-[state=closed]:animate-collapse data-[state=open]:animate-expand">
        <div className="space-y-2 pl-2">
          {householdList.length > 0 && (
            <div className="space-y-0.5">{householdList.map(renderLink)}</div>
          )}
          {userList.length > 0 && (
            <div className="space-y-0.5">
              <p className="px-3 pt-1 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground/80">
                My Dashboards
              </p>
              {userList.map(renderLink)}
            </div>
          )}
        </div>
      </CollapsiblePrimitive.Panel>
    </CollapsiblePrimitive.Root>
  );
}
```

- [ ] **Step 4: Run the nav-group test to verify it passes**

Run: `cd frontend && npx vitest run src/components/features/navigation/__tests__/dashboards-nav-group.test.tsx`
Expected: PASS (4 tests).

- [ ] **Step 5: Type-check**

Run: `cd frontend && npx tsc -b`
Expected: no errors. (Confirms no remaining importer relied on the removed `sortDashboards`/`computeReorderEntries`/`SortableDashboardRow` exports from this file.)

- [ ] **Step 6: Commit**

```bash
git add frontend/src/components/features/navigation/dashboards-nav-group.tsx \
        frontend/src/components/features/navigation/__tests__/dashboards-nav-group.test.tsx
git commit -m "refactor(frontend): reduce DashboardsNavGroup to navigation-only links"
```

---

## Task 7: Full verification of the affected frontend

Run the complete frontend gate to catch cross-file regressions (app-shell, mobile-drawer, and any suite touching the changed modules).

**Files:** none (verification only).

- [ ] **Step 1: Type-check the whole frontend**

Run: `cd frontend && npx tsc -b`
Expected: no errors.

- [ ] **Step 2: Lint**

Run: `cd frontend && npm run lint`
Expected: no errors. (If lint flags an unused import in a file this plan touched — e.g. a leftover icon import — remove it and re-run.)

- [ ] **Step 3: Run the full frontend test suite**

Run: `cd frontend && npm run test`
Expected: all suites PASS, including `ordering`, `dashboard-row`, `DashboardManagementPage`, `dashboard-kebab-menu`, `dashboards-nav-group`, `DashboardRedirect`, and the navigation/app-shell suites.

- [ ] **Step 4: Commit any verification fixups (only if Steps 1–3 required edits)**

```bash
git add -A
git commit -m "fix(frontend): resolve verification findings for dashboard management page"
```

(Skip this commit if Steps 1–3 were clean.)

---

## Self-Review

**Spec coverage (PRD FR-1 … FR-14, design §):**
- FR-1 (list by scope, sorted): Task 4 page `householdList`/`userList` via `sortDashboards`. ✓
- FR-2 (per-scope reorder, one call, no cross-scope): Task 4 two `DndContext`s + `computeReorderEntries` + one `mutate` per scope. ✓
- FR-3 (rename/set-default/promote/copy/delete + open + edit): Task 3 row hosts reused `DashboardKebabMenu` (full action set) + open link + edit link. ✓
- FR-4 (New Dashboard → `NewDashboardModal`): Task 4 header button + modal host. ✓
- FR-5 (mobile, no-hover controls): Task 2 always-visible kebab + Task 3 full-width card with fixed tap targets. ✓
- FR-6 (loading skeleton, empty scope omitted): Task 4 loading branch + per-scope conditional render. ✓
- FR-7 (promote/copy navigation unchanged): kebab reused as-is (no nav change). ✓
- FR-8–FR-10 (sidebar nav-only, active highlight, mobile close): Task 6 plain `NavLink` with active classes + `onItemClick`. ✓
- FR-11 (Manage Dashboards in `management`, `LayoutGrid`): Task 5. ✓
- FR-12 (route repoint): Task 5. ✓
- FR-13 (index/dashboard/landing untouched): Task 5 leaves those routes alone. ✓
- FR-14 (delete `DashboardsIndexRedirect` if unused): Task 5 deletes after grep-verified single importer. ✓
- Design §3.2 (helpers → `lib/dashboard/ordering.ts`, tests move): Task 1. ✓
- Design §5 (kebab restyle is the only shared-component edit): Task 2. ✓
- A11y (keyboard reorder, labels): `KeyboardSensor` + `sortableKeyboardCoordinates` (Task 4), grip + kebab `aria-label` (Task 3 / reused kebab). ✓

**Placeholder scan:** No TBD/TODO/"similar to"/"add error handling" — every code step shows full content. ✓

**Type consistency:** `sortDashboards`/`computeReorderEntries` signatures identical across Tasks 1, 4, 6 and the original. `DashboardRow` props `{ dashboard: Dashboard; defaultDashboardId: string | null }` consistent between Task 3 (definition + test) and Task 4 (usage). `DashboardKebabMenu` props `{ dashboard, isDefault }` match the existing component. `defaultDashboardId` derived as `prefsData?.data?.[0]?.attributes.defaultDashboardId ?? null` (same access shape as the existing kebab/nav-group). ✓
