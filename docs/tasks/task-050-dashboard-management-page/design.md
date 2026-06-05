# Dashboard Management Page — Design

Status: Approved (design phase)
Created: 2026-06-05
PRD: `./prd.md`
Scope: frontend-only

---

## 1. Summary

Relocate every dashboard-management affordance out of the sidebar's
`DashboardsNavGroup` and into a dedicated **Dashboard Management** page rendered
at `/app/dashboards`. The sidebar group becomes navigation-only (plain
`NavLink`s grouped by scope). A new "Manage Dashboards" entry in the
`management` nav group routes to the page. No backend, contract, or data-model
changes — the page composes the existing dashboard hooks.

This is a **move + extract** refactor, not new behavior: the action set, the
reorder contract, the sort rule, and the modal are all reused as-is. The design
work is mostly about *where the existing pieces live* and *how the row renders
on a full page instead of a cramped drawer*.

## 2. Resolved design decisions

The PRD deferred four open questions to this phase. Decisions (with the user):

| # | Question | Decision | Rationale |
|---|----------|----------|-----------|
| Q1 | Row action affordance | **Always-visible kebab** (no hover-reveal) | Keeps parity with today's action set and the existing `DashboardKebabMenu` UI; making the trigger always visible (drop the `opacity-0 group-hover/row:opacity-100` classes) satisfies FR-5's "operable without hover" on touch. |
| Q2 | Reorder mechanism | **Drag-only** (dnd-kit, keyboard-operable) | Satisfies FR-2 with the least new surface area; reuses the exact `DndContext`/`SortableContext`/`useSortable` wiring already proven in the sidebar. On a full-width page the grip is a comfortably sized hit target, unlike the cramped drawer. No move-up/down buttons. |
| Q3 | Page header | **Standard management header** — `h1` "Dashboards" + primary "New Dashboard" button, no description line | Matches `HouseholdsPage` (`p-4 md:p-6 space-y-4`, `flex items-center justify-between`, `h1 text-xl md:text-2xl`). Consistency over novelty. |
| Q4 | Nav icon | **`LayoutGrid`** | Distinct from the per-dashboard `LayoutDashboard` icon while still reading as "a set of dashboards." |

## 3. Architecture

### 3.1 Component inventory

```
pages/
  DashboardManagementPage.tsx        NEW  — the page (header, scope sections, modal host)
  DashboardsIndexRedirect.tsx        DELETE (if no other importer) — only App.tsx imports it

components/features/dashboards/
  dashboard-row.tsx                  NEW  — one draggable row: grip + name + open/edit links + kebab
  dashboard-kebab-menu.tsx           REUSE — action menu; trigger restyled to always-visible
  new-dashboard-modal.tsx            REUSE as-is (no changes)

components/features/navigation/
  dashboards-nav-group.tsx           MODIFY — strip drag/kebab/new; plain NavLinks
  nav-config.ts                      MODIFY — add "Manage Dashboards" to `management` group

lib/dashboard/ (or kept in nav group, re-exported)
  sort + reorder helpers             REUSE — sortDashboards, computeReorderEntries

App.tsx                              MODIFY — /app/dashboards → DashboardManagementPage
```

### 3.2 Where the shared helpers live

`sortDashboards` and `computeReorderEntries` are pure functions currently
**exported from `dashboards-nav-group.tsx`** and unit-tested there. The
management page needs both. Two options:

- **(A) Move them to a neutral module** (e.g.
  `lib/dashboard/ordering.ts`) and import from both the nav group and the page.
- (B) Import them from the nav group into the page.

**Decision: (A).** After this task the nav group no longer reorders, so it is
the wrong home for `computeReorderEntries`. Moving both helpers (and their unit
tests) to `lib/dashboard/ordering.ts` keeps the page and any future consumer
from importing a navigation component for pure logic, and respects the project's
"don't break service boundaries / keep abstractions clean" guidance. The nav
group keeps importing `sortDashboards` from the new module. The existing
`sortDashboards`/`computeReorderEntries` describe blocks in
`dashboards-nav-group.test.tsx` move to `lib/dashboard/__tests__/ordering.test.ts`
unchanged.

### 3.3 The row component (`DashboardRow`)

A single presentational+wiring unit so the page body stays readable and the row
can be tested in isolation. Responsibilities:

- Be a dnd-kit sortable item (`useSortable({ id })`) — owns the grip button,
  drag listeners, and drag styling (transform/opacity), lifted from the
  sidebar's `SortableDashboardRow`.
- Render the dashboard name as a link to **open** it (`/app/dashboards/:id`) and
  a secondary affordance to **edit** in the designer (`/app/dashboards/:id/edit`)
  — FR-3's "navigate to the dashboard and to its designer." On the page these
  become explicit controls (e.g. the name links to open; a small "Edit" / pencil
  link or a kebab item routes to the designer) rather than the sidebar's
  single nav link.
- Host the `DashboardKebabMenu` (always-visible trigger) for rename / set-default
  / promote|copy / delete.
- Show a "Default" badge when `dashboard.id === defaultDashboardId` (read-only
  affordance; the actual set-default action stays in the kebab, disabled/"Already
  default" when current — unchanged behavior).

Props (mirrors what the sidebar passed, minus sidebar-only sizing):
`{ dashboard: Dashboard; defaultDashboardId: string | null }`.

Full-width, touch-first styling: `flex items-center gap-2 rounded-md border p-3`
(card-like rows like `HouseholdsPage`'s invitation rows), grip and kebab as
fixed-size tap targets, name truncates.

### 3.4 The page (`DashboardManagementPage`)

Composition, top to bottom:

1. **Header** — `h1` "Dashboards" + `Button` "New Dashboard" (opens
   `NewDashboardModal`), per Q3.
2. **Loading state** — `useDashboards().isLoading` → skeleton rows
   (`Skeleton` ×3, `role="status" aria-label="Loading"`), matching
   `HouseholdsPage` (FR-6).
3. **Error state** — `isError` → `ErrorCard` (matches `HouseholdsPage`).
4. **Household section** — heading "Household Dashboards", rendered only when
   `householdList.length > 0`; a `DndContext` + `SortableContext` wrapping the
   household rows, scope `"household"` (FR-1, FR-6).
5. **User section** — heading "My Dashboards", rendered only when
   `userList.length > 0`; its own `DndContext` + `SortableContext`, scope
   `"user"`.
6. **Empty state** — when there are zero dashboards of any scope, show a
   centered empty prompt with a "New Dashboard" affordance (mirrors
   `HouseholdsPage`'s empty block). Per-scope empties are simply omitted
   (no section), matching today.
7. **Modal host** — one `NewDashboardModal` controlled by page state.

The page reads `useDashboards()` and `useHouseholdPreferences()` once and
derives `householdList`/`userList` with `useMemo` + `sortDashboards`, exactly as
the sidebar does today (FR-1). It owns `useReorderDashboards` and a per-scope
`handleDragEnd` identical to the current sidebar handler (one mutate call per
scope via `computeReorderEntries`) (FR-2). No new network round-trips — the page
hits the same `useDashboards` cache the sidebar already populates (NFR
performance).

### 3.5 Two `DndContext`s, one per scope

Keep the existing pattern: a separate `DndContext`/`SortableContext` per scope
section. This makes cross-scope drags structurally impossible (an item in the
household context can never drop into the user context), which is exactly what
FR-2 requires (the backend rejects mixed-scope batches). `handleDragEnd` is
curried by scope and selects the matching list, unchanged from the sidebar.

### 3.6 Sidebar after the change (`DashboardsNavGroup`)

Reduce to navigation-only (FR-8 – FR-10):

- Remove imports/usage of `useReorderDashboards`, dnd-kit, `DashboardKebabMenu`,
  `NewDashboardModal`, `Plus`/`GripVertical`, and the `modalOpen` state.
- Replace `SortableDashboardRow` with a plain `NavLink` row (the existing
  `NavLink` markup minus the grip button and kebab) — keep `LayoutDashboard`
  icon, truncation, active-highlight classes, and `onItemClick` (mobile drawer
  close) (FR-10).
- Keep the collapsible group, the household/"My Dashboards" split, and
  `sortDashboards` (imported from `lib/dashboard/ordering.ts`).
- `useHouseholdPreferences` is no longer needed here (default badge/marking was
  only consumed by the kebab) — remove it from the nav group. Confirm nothing
  else in the group reads prefs before deleting.

### 3.7 Routing & entry point

- `App.tsx`: change `<Route path="dashboards" element={<DashboardsIndexRedirect />} />`
  to `element={<DashboardManagementPage />}` (FR-12). Leave the index (`/app`),
  `dashboard`, `dashboards/:dashboardId`, and `:dashboardId/edit` routes
  untouched (FR-13). `DashboardManagementPage` is imported eagerly like the other
  page components (the designer alone is `React.lazy`; the management page is
  lightweight and matches `HouseholdsPage`'s eager import).
- Delete `DashboardsIndexRedirect.tsx` and its `App.tsx` import once the route no
  longer references it (FR-14). Verified: the only importer is `App.tsx`
  (`grep` confirms `DashboardsIndexRedirect` appears only in `App.tsx` and its
  own definition). `DashboardRedirect` itself is unaffected.
- `nav-config.ts`: add to the `management` group, before or after Households,
  `{ to: "/app/dashboards", icon: LayoutGrid, label: "Manage Dashboards" }`, and
  import `LayoutGrid` from `lucide-react` (FR-11). Both the desktop sidebar and
  mobile drawer render the `management` group from this shared config, so the
  entry appears in both automatically.

## 4. Data flow

```
useDashboards() ─┐
                 ├─► useMemo(sortDashboards) ─► householdList / userList ─► sections ─► DashboardRow×n
useHouseholdPreferences() ─► defaultDashboardId ──────────────────────────────────────────┘ (Default badge + kebab)

DashboardRow drag end ─► page handleDragEnd(scope) ─► computeReorderEntries ─► useReorderDashboards.mutate(entries)
DashboardRow kebab ─► DashboardKebabMenu ─► useUpdateDashboard / useDeleteDashboard /
                                            usePromoteDashboard / useCopyDashboardToMine /
                                            useUpdateHouseholdPreferences
Header "New Dashboard" ─► NewDashboardModal ─► useCreateDashboard (+ useCopyDashboardToMine for copy-of)
```

All mutations invalidate `dashboardKeys.all(tenant, household)` (existing hook
behavior), so the page's `useDashboards` list refetches and re-sorts after any
action. Tenant/household scoping is entirely inside the hooks — the page passes
no tenant ids (NFR multi-tenancy).

### Promote / copy navigation (FR-7)

`DashboardKebabMenu` already navigates to the resulting dashboard on promote/copy
success (`navigate('/app/dashboards/:id')`). Keeping the kebab as-is means
promote/copy from the management page navigates **away** to the new dashboard —
which FR-7 explicitly accepts ("navigating away to the new dashboard is
acceptable and matches existing behavior"). No change to that behavior; the
resulting dashboard remains discoverable (it's in the list on return).

## 5. Reuse & extraction summary

| Piece | Action | Notes |
|-------|--------|-------|
| `sortDashboards`, `computeReorderEntries` | **Move** to `lib/dashboard/ordering.ts` | Pure logic; tests move with them. Nav group + page import from here. |
| `SortableDashboardRow` markup | **Extract → `DashboardRow`** | Page-styled (full-width, card row), always-visible kebab, adds open/edit affordances. |
| `DashboardKebabMenu` | **Reuse**, restyle trigger | Drop `opacity-0 … group-hover/row:opacity-100` so the trigger is always visible (Q1). Action logic unchanged. |
| `NewDashboardModal` | **Reuse as-is** | No edits. |
| dnd-kit wiring (`DndContext`/`SortableContext`/sensors) | **Reuse**, two contexts | One per scope; identical sensors (`PointerSensor` + `KeyboardSensor` w/ `sortableKeyboardCoordinates`) for keyboard reorder (NFR a11y). |
| `dashboardNameSchema` | **Reuse** | Via the kebab's rename dialog. |

The kebab restyle is the one shared-component edit. Risk: the sidebar no longer
renders the kebab after this task, so the restyle only affects the page — but
verify no other component imports `DashboardKebabMenu` (grep: only the nav group
and its test reference it today; the nav group drops it).

## 6. Accessibility

- **Keyboard reorder:** preserved via dnd-kit `KeyboardSensor` +
  `sortableKeyboardCoordinates` (same as today). Grip button keeps an
  `aria-label` ("Drag {name} to reorder").
- **Action trigger:** kebab keeps `aria-label` ("Dashboard actions for {name}").
- **Always-visible (not hover-gated):** the page's grip and kebab are visible
  without hover, satisfying FR-5 / NFR responsiveness on touch.
- **Dialogs:** rename + delete dialogs reuse the existing `Dialog` component's
  focus management (unchanged).
- **Loading:** skeleton container uses `role="status" aria-label="Loading"`.

## 7. Testing strategy

New / updated tests (Vitest + Testing Library), matching existing conventions:

- **`lib/dashboard/__tests__/ordering.test.ts`** — the relocated
  `sortDashboards` / `computeReorderEntries` describe blocks, unchanged
  assertions.
- **`pages/__tests__/DashboardManagementPage.test.tsx`** (new):
  - renders household + user sections sorted, omits an empty scope section;
  - renders the page-level empty state when there are no dashboards;
  - "New Dashboard" button opens the modal;
  - each row exposes the kebab (actions) and an open link to
    `/app/dashboards/:id`;
  - loading state shows skeletons.
  Mocks the dashboard + preferences hooks like the existing nav-group test does.
- **`components/features/dashboards/__tests__/dashboard-row.test.tsx`** (new,
  optional but recommended): row renders name link, edit affordance, kebab, and
  Default badge when id matches.
- **`dashboards-nav-group.test.tsx`** (update): drop the `sortDashboards` /
  `computeReorderEntries` describe blocks (moved); update the
  `DashboardsNavGroup` cases — assert rows are plain `NavLink`s, assert **no**
  "New Dashboard" button, no kebab trigger, no grip button. The current
  "renders only new-dashboard button when list is empty" test is replaced by an
  assertion that the empty group renders no links and no new-dashboard button.
  Update the `vi.mock` for `use-dashboards` to drop now-unused mutation mocks.
- **`DashboardRedirect.test.tsx`** — unaffected (the redirect component stays).
  Confirm no test imports `DashboardsIndexRedirect`.

## 8. Verification checklist (maps to acceptance criteria)

- `/app/dashboards` renders the management page (no redirect). `/app`,
  `/app/dashboard`, post-login landing still resolve to default dashboard.
- "Manage Dashboards" appears in the Management group (desktop + mobile drawer)
  and routes to `/app/dashboards`.
- Page lists household + user sections, each sorted by `sortOrder` then
  `createdAt`.
- Drag within a scope reorders and persists via one `useReorderDashboards` call;
  cross-scope moves impossible (separate contexts).
- Rename / delete (confirm) / set-default (disabled when current) / promote
  (user scope) / copy (household scope) all work from the row kebab.
- "New Dashboard" opens `NewDashboardModal` and creates.
- Reorder keyboard-operable; grip + kebab have accessible labels.
- Mobile: full-width rows, controls operable without hover.
- Sidebar `DashboardsNavGroup` renders nav links only — no grips, kebabs, or
  new-dashboard button; active highlight + mobile drawer close-on-tap intact.
- `DashboardsIndexRedirect` removed (only `App.tsx` imported it).
- No backend / contract / migration / Go changes.
- Frontend type-check, lint, and affected suites
  (`ordering`, `DashboardManagementPage`, `dashboard-row`,
  `dashboards-nav-group`) pass.

## 9. Risks & mitigations

- **Shared-helper move breaks imports.** Mitigate: update the nav group import
  and the test path in the same change; type-check catches stragglers.
- **Kebab restyle leaks to another consumer.** Mitigate: grep confirms the nav
  group is the only non-test importer; it drops the kebab, so the always-visible
  trigger is page-only in practice.
- **Two `DndContext`s + drag styling regress on the page's card layout.**
  Mitigate: row-level `dashboard-row` test + manual mobile-width check; reuse the
  identical sensor/strategy config from the working sidebar.
- **Forgetting an FR action on the row.** Mitigate: the kebab is reused wholesale,
  so the full action set comes along; the verification checklist enumerates each.

## 10. Out of scope (from PRD non-goals)

No backend/contract/data-model/migration changes; no designer or widget-config
changes; no scope/default/promote/copy/`sortOrder` semantic changes; no change
to how `/app`, `/app/dashboard`, or landing resolve; no net-new management
capability (bulk/archiving); no sidebar visual redesign beyond removing the
relocated controls.
