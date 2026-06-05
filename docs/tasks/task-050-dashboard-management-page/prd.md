# Dashboard Management Page — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-06-05
---

## 1. Overview

Today, every dashboard-management affordance lives inline inside the sidebar's
`DashboardsNavGroup` component, which is rendered in both the desktop sidebar
(`app-shell.tsx`) and the mobile drawer (`mobile-drawer.tsx`). That single
component carries a lot of weight: it lists every dashboard as a navigation
link, exposes a drag-to-reorder grip on each row, attaches a per-row kebab menu
(rename, set-as-default, promote-to-household, copy-to-mine, delete), splits the
list into "household" and "My Dashboards" sections, and renders a "New
Dashboard" button. On desktop this is dense but usable. On mobile — inside a
narrow drawer — the tiny grip handles and nested kebab pop-overs are awkward to
operate, and the management chrome inflates the nav's vertical footprint,
wasting space and pushing real navigation items down.

This task separates the two concerns. The sidebar keeps doing what it is good
at — letting a user click a dashboard to navigate to it. Everything else
(reordering, renaming, deleting, promoting, copying, setting a default, creating
a new dashboard) moves to a dedicated **Dashboard Management** page reachable
from a "Manage Dashboards" link in the sidebar's "Management" nav group. The
sidebar's per-dashboard rows shed their drag handles, kebab menus, and the "New
Dashboard" button, becoming plain navigation links.

The change is **frontend-only**. Every API hook the new page needs already
exists (`useDashboards`, `useCreateDashboard`, `useUpdateDashboard`,
`useDeleteDashboard`, `useReorderDashboards`, `usePromoteDashboard`,
`useCopyDashboardToMine`, plus `useHouseholdPreferences` /
`useUpdateHouseholdPreferences` for the default-dashboard setting). No backend,
contract, data-model, or migration changes are required.

## 2. Goals

Primary goals:
- Create a dedicated dashboard-management page at `/app/dashboards` that lets a
  user reorder, rename, delete, promote, copy, set-as-default, and create
  dashboards across both the "household" and "user" scopes.
- Add a "Manage Dashboards" link to the sidebar's "Management" nav group
  (desktop sidebar and mobile drawer) that routes to `/app/dashboards`.
- Reduce `DashboardsNavGroup` to a navigation-only role: it still lists every
  dashboard as a clickable link, grouped by scope, but no longer renders drag
  handles, kebab menus, or the "New Dashboard" button.
- Preserve every management capability that exists today — nothing is dropped,
  only relocated.
- Ensure the management page is comfortably usable on mobile (full-width rows,
  touch-friendly controls), resolving the original pain point.

Non-goals:
- Any backend, API contract, data-model, or migration change.
- Changes to the dashboard **designer** (widget editing at
  `/app/dashboards/:dashboardId/edit`) or to widget configuration.
- Changing the semantics of dashboard scopes, the default-dashboard mechanism,
  promote/copy behavior, or `sortOrder` storage.
- Changing how `/app`, `/app/dashboard`, or the post-login landing resolve to
  the default dashboard (only `/app/dashboards` changes meaning).
- Adding any net-new management capability (e.g. bulk actions, archiving) beyond
  what the sidebar offers today.
- Redesigning the sidebar's visual style beyond removing the relocated controls.

## 3. User Stories

- As a user on mobile, I want to manage and reorder my dashboards on a real
  page with full-width, touch-friendly controls, so that I am not fighting tiny
  grip handles and nested menus inside a cramped drawer.
- As a user on desktop, I want the sidebar's Dashboards section to stay compact
  and focused on navigation, so that it doesn't waste vertical space on controls
  I rarely use.
- As a user, I want a single, discoverable "Manage Dashboards" entry point, so
  that I know exactly where to go to rename, reorder, delete, or create a
  dashboard.
- As a user, I want every dashboard action I have today (reorder, rename,
  delete, set default, promote, copy, create) to still be available after the
  move, so that nothing I rely on disappears.
- As a user, I want to click a dashboard in the sidebar and go straight to it,
  without management controls getting in the way.

## 4. Functional Requirements

### 4.1 Dashboard Management Page

A new page component (e.g. `DashboardManagementPage`) rendered at the
`/app/dashboards` route. It must:

- **FR-1.** List all dashboards returned by `useDashboards()`, split into two
  sections by scope: "Household Dashboards" (`scope === "household"`) and "My
  Dashboards" (`scope === "user"`), each sorted by the same rule used today —
  `sortOrder` ascending, then `createdAt` ascending as a stable tiebreaker
  (reuse `sortDashboards`).
- **FR-2.** Within each scope section, support drag-to-reorder using the same
  dnd-kit primitives already in use. Reordering issues **one
  `useReorderDashboards` call per scope** with 0-indexed `sortOrder` entries
  (reuse `computeReorderEntries`); the backend rejects mixed-scope batches, so a
  drag must never move an item across scope sections.
- **FR-3.** Each dashboard row exposes the full action set currently in
  `DashboardKebabMenu`:
  - **Rename** — opens a dialog with a name field validated by the existing
    `dashboardNameSchema`; submits via `useUpdateDashboard`.
  - **Set as my default** — calls `useUpdateHouseholdPreferences` to set
    `defaultDashboardId`; disabled / labeled "Already default" when the row is
    the current default (`useHouseholdPreferences`).
  - **Promote to household** — shown only for `user`-scope rows; calls
    `usePromoteDashboard`.
  - **Copy to mine** — shown only for `household`-scope rows; calls
    `useCopyDashboardToMine`.
  - **Delete** — opens a confirmation dialog; calls `useDeleteDashboard`.
  - Each row also lets the user navigate to the dashboard (open it) and to its
    designer (`/app/dashboards/:id/edit`).
- **FR-4.** Provide a "New Dashboard" action on the page that opens the existing
  `NewDashboardModal` (via `useCreateDashboard`).
- **FR-5.** The page must be responsive: on mobile, rows are full-width with
  touch-friendly hit targets; reorder controls and per-row actions are operable
  without hover (controls visible/accessible without requiring a hover state,
  unlike the current hover-reveal sidebar grips and kebab triggers).
- **FR-6.** Loading and empty states: while dashboards are loading, show an
  appropriate skeleton/placeholder; when a scope section has no dashboards, omit
  the section (matching today's behavior where empty lists are not rendered).
- **FR-7.** After promote or copy, navigation behavior should match today
  (promote/copy currently navigates to the resulting dashboard). On the
  management page, the resulting dashboard should remain discoverable; navigating
  away to the new dashboard is acceptable and matches existing behavior.

### 4.2 Sidebar Navigation Group (`DashboardsNavGroup`)

- **FR-8.** Continue to render the collapsible "Dashboards" group listing every
  dashboard as a `NavLink` to `/app/dashboards/:id`, grouped into household and
  "My Dashboards" sections, sorted identically to today.
- **FR-9.** Remove from the sidebar rows: the drag-to-reorder grip handle and
  dnd-kit wiring, the per-row `DashboardKebabMenu`, and the "New Dashboard"
  button. These move to the management page.
- **FR-10.** The active-route highlight, truncation, icon, and click-to-navigate
  behavior of each row are preserved. On mobile, tapping a dashboard still
  closes the drawer (`onItemClick`).

### 4.3 Entry Point

- **FR-11.** Add a "Manage Dashboards" `NavItem` to the `management` group in
  `nav-config.ts` pointing to `/app/dashboards`, with an appropriate icon
  (e.g. `LayoutDashboard` or `Settings2`/`SlidersHorizontal`). It appears in
  both the desktop sidebar and the mobile drawer because both render the
  `management` group from the shared config.

### 4.4 Routing

- **FR-12.** Change the `/app/dashboards` route in `App.tsx` to render the new
  `DashboardManagementPage` instead of `DashboardsIndexRedirect`.
- **FR-13.** `/app` (index) and `/app/dashboard` continue to render
  `DashboardRedirect` and resolve to the user's default dashboard — unchanged.
- **FR-14.** If `DashboardsIndexRedirect` becomes unused after FR-12, remove it
  (it is a thin re-export of `DashboardRedirect`); verify no other importer
  depends on it before deletion.

## 5. API Surface

No new or modified endpoints. The page composes existing dashboard-service
operations through existing hooks:

| Capability        | Hook                            |
|-------------------|---------------------------------|
| List dashboards   | `useDashboards`                 |
| Create            | `useCreateDashboard`            |
| Rename / update   | `useUpdateDashboard`            |
| Delete            | `useDeleteDashboard`            |
| Reorder (by scope)| `useReorderDashboards`          |
| Promote to hh     | `usePromoteDashboard`           |
| Copy to mine      | `useCopyDashboardToMine`        |
| Read default      | `useHouseholdPreferences`       |
| Set default       | `useUpdateHouseholdPreferences` |

Reorder contract (unchanged): one call per scope, payload is an array of
`{ id, sortOrder }` with 0-indexed `sortOrder`; mixed-scope batches are rejected
by the backend.

## 6. Data Model

No changes. Relevant existing fields on the `Dashboard` model:
`attributes.name`, `attributes.scope` (`"household" | "user"`),
`attributes.sortOrder`, `attributes.createdAt`. The default dashboard is stored
on household preferences as `attributes.defaultDashboardId`. All tenant scoping
is handled by the existing hooks/services; no new `tenant_id` handling is
introduced here.

## 7. Service Impact

- **frontend** (only affected service):
  - New: `frontend/src/pages/DashboardManagementPage.tsx` (+ tests).
  - Modified: `frontend/src/components/features/navigation/dashboards-nav-group.tsx`
    — strip drag/kebab/new affordances down to navigation links.
  - Modified: `frontend/src/components/features/navigation/nav-config.ts` — add
    "Manage Dashboards" item to the `management` group.
  - Modified: `frontend/src/App.tsx` — repoint `/app/dashboards` to the new page.
  - Reused / possibly relocated: `DashboardKebabMenu` and `NewDashboardModal`
    (the kebab logic may be refactored into a row-action component shared by /
    moved to the page; the modal is reused as-is).
  - Removed (if unused): `frontend/src/pages/DashboardsIndexRedirect.tsx`.
  - Updated tests: `dashboards-nav-group.test.tsx` (assertions about
    drag/kebab/new in the sidebar), plus any test importing the removed redirect.
- No Go services, shared libraries, Docker builds, or migrations are touched.

## 8. Non-Functional Requirements

- **Multi-tenancy:** Inherited from existing hooks/services; the page must not
  bypass them or introduce direct cross-service calls.
- **Accessibility:** Reorder must remain keyboard-operable (dnd-kit
  `KeyboardSensor` + `sortableKeyboardCoordinates`, as today). Drag handles and
  action triggers need accessible labels. Dialogs preserve focus management.
- **Responsiveness:** Page is usable from the smallest supported mobile width up
  through desktop; controls do not depend on hover to be discoverable on touch.
- **Performance:** No additional network round-trips beyond what the sidebar
  does today; the page reads the same `useDashboards` cache.
- **Consistency:** Reuse existing primitives (`sortDashboards`,
  `computeReorderEntries`, `NewDashboardModal`, dialog/menu UI components,
  `dashboardNameSchema`) rather than reimplementing.
- **No regressions:** Sidebar navigation, active highlighting, and mobile
  drawer close-on-navigate continue to work.

## 9. Open Questions

- **Q1 (row action affordance):** On the management page, should per-row actions
  use a kebab menu (consistent with today) or surface the most common actions
  (rename, reorder, set-default) as inline buttons with the rest behind an
  overflow menu? — Defer to the design phase; either satisfies the FRs.
- **Q2 (icon choice):** Which Lucide icon best represents "Manage Dashboards" in
  the Management group without clashing with the per-dashboard
  `LayoutDashboard` icon? — Design-phase detail.
- **Q3 (page header / breadcrumb):** Should the page show a title/description
  header consistent with other management pages (e.g. Households, Data
  Retention)? Assumed yes; confirm pattern during design.
- **Q4 (reorder UX on the page):** Keep drag-and-drop as the only reorder
  mechanism, or also offer move-up/move-down buttons for better touch/keyboard
  ergonomics? — Design-phase decision; drag-and-drop alone satisfies FR-2.

## 10. Acceptance Criteria

- [ ] Navigating to `/app/dashboards` renders the new Dashboard Management page
      (no longer redirects to the default dashboard).
- [ ] `/app`, `/app/dashboard`, and post-login landing still resolve to the
      user's default dashboard (unchanged).
- [ ] The Management nav group (desktop sidebar + mobile drawer) shows a "Manage
      Dashboards" link that routes to `/app/dashboards`.
- [ ] The management page lists household and user dashboards in separate
      sections, each sorted by `sortOrder` then `createdAt`.
- [ ] On the page, dragging within a scope reorders and persists via a single
      `useReorderDashboards` call for that scope; cross-scope moves are not
      possible.
- [ ] On the page, each dashboard can be renamed, deleted (with confirmation),
      set as default (disabled when already default), and — per scope —
      promoted to household (user scope) or copied to mine (household scope).
- [ ] The page provides a "New Dashboard" action that opens `NewDashboardModal`
      and creates a dashboard.
- [ ] Reorder is keyboard-operable and all interactive controls have accessible
      labels.
- [ ] The page is usable on a mobile-width viewport: full-width rows, controls
      operable without hover.
- [ ] The sidebar `DashboardsNavGroup` no longer renders drag handles, kebab
      menus, or a "New Dashboard" button; it lists dashboards as navigation
      links only, with active highlighting and mobile drawer close-on-tap intact.
- [ ] `DashboardsIndexRedirect` is removed if and only if it is no longer
      imported anywhere.
- [ ] No backend, contract, migration, or Go-service changes are introduced.
- [ ] Frontend type-check, lint, and the affected test suites
      (`dashboards-nav-group`, the new page, navigation/app-shell) pass.
