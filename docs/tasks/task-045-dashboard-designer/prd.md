# Dashboard Designer — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-04-23
---

## 1. Overview

Today, `frontend/src/pages/DashboardPage.tsx` is a hand-assembled layout that hard-codes a specific set of widgets in a specific order. Changing the dashboard requires a code deploy. This feature introduces a *dashboard designer* that lets a household compose one or more dashboards from a fixed registry of widgets, saving the layout and per-widget configuration as JSON on the server.

A new backend service, `dashboard-service`, owns the `dashboard.dashboards` table and exposes JSON:API CRUD endpoints under `/api/v1/dashboards`. The frontend gains two user-facing surfaces: a **renderer** that turns dashboard JSON into a rendered page, and a **designer** (edit mode) that lets users add, remove, arrange, resize, and configure widgets on a 12-column responsive grid.

Initial validation is achieved by seeding the existing `DashboardPage` content as a household's first dashboard (a "Home" dashboard) and switching the `/app/dashboard` route over to the new renderer. Once that round-trips, the old hard-coded page is deleted. Households may create additional dashboards (e.g. "Weekend", "Planning") and each user may also create private dashboards visible only to themselves. Every dashboard appears as its own sidebar entry.

## 2. Goals

Primary goals:
- Let households design JSON-backed dashboards without a deploy
- Provide a graphical, grid-based designer (drag, resize, configure, save)
- Support multiple dashboards per household **and** per user, each surfaced as its own sidebar entry
- Seed the current `DashboardPage` as the household's first default dashboard to validate the model
- Promote every widget currently rendered on `DashboardPage` into a reusable registry entry
- Provide per-widget configuration so the same widget type can be placed multiple times with different behavior

Non-goals:
- User-authored / scriptable widgets (registry is fixed in code)
- Cross-household dashboard sharing, templates gallery, or marketplace
- Real-time collaborative editing (last-writer-wins is acceptable)
- Per-breakpoint layout variants stored separately (one layout that responsively stacks)
- Dashboard-level ACLs beyond the household/user scope split (any household member can edit household dashboards; only the owner can see/edit their user dashboards)
- Designer for mobile (editing on narrow viewports is out of scope for v1 — the renderer is responsive, the designer requires a tablet-or-wider viewport)
- Versioning / history / rollback of dashboard edits

## 3. User Stories

- As a household member, I want to design a dashboard that emphasizes the information my household cares about, so the dashboard page matches how we actually live
- As a household member, I want to create more than one dashboard (e.g. "Home", "Weekend", "Trip Planning") so I can switch between contexts without scrolling a huge page
- As an individual, I want to create a private dashboard that only I see, so I can track things I don't need to share with the household
- As a household member, I want each dashboard to appear in the sidebar so switching between them is a single click
- As an editor, I want to drag widgets onto a grid, resize them, and configure them (e.g. "show overdue tasks only") so one widget type can serve multiple purposes
- As an editor, I want a clear "Save" / "Discard" step so I can experiment without committing changes
- As any user, I want to mark one household dashboard and one user dashboard as my default, so `/app/dashboard` opens the right thing on login
- As any user, I want the first load of a brand-new household to show the same dashboard we have today, so the feature is never "empty on day one"

## 4. Functional Requirements

### 4.1 Dashboard Resource

A dashboard has:
- `id` (UUID)
- `tenant_id`
- `household_id` (always present)
- `user_id` (nullable — null = household-scoped, non-null = private to that user)
- `name` (1–80 chars, trimmed)
- `sort_order` (non-negative int; determines sidebar order within its scope for that user)
- `layout` (JSON; structure in §6 and in `data-model.md`)
- `schema_version` (int; starts at 1)
- `created_at`, `updated_at`
- Soft delete is not required — DELETE is immediate (but see §4.7 on the "last dashboard in scope" rule)

`is_default` is NOT stored on the dashboard row. The "default dashboard" is a per-user preference stored separately (see §4.6) so it can point at either a household- or user-scoped dashboard.

### 4.2 Scoping Rules

- Household-scoped dashboard (`user_id = NULL`): visible to and editable by every member of the household.
- User-scoped dashboard (`user_id = <caller>`): visible only to that user within that household. Not transferable between users or households.
- Personal widgets (habits, workout) render the **viewer's** user-scoped data regardless of which scope the dashboard itself belongs to. The dashboard does not "lock in" whose habits it shows.

### 4.3 Listing & Sidebar Integration

- `GET /api/v1/dashboards` returns all dashboards the caller can see for the current household: every household-scoped dashboard plus the caller's own user-scoped dashboards.
- The frontend sidebar renders one link per dashboard, grouped under a single "Dashboards" section. Household dashboards are listed first, then user dashboards. Within each group they are ordered by `sort_order` then `created_at`.
- A "+ New dashboard" action appears at the bottom of the section and opens a modal asking for `name` and `scope` (Household / Mine).
- The currently active dashboard is highlighted in the sidebar.

### 4.4 Rendering

- Route: `/app/dashboards/{dashboardId}` renders a dashboard in read mode.
- Legacy route `/app/dashboard` redirects to the caller's default dashboard (see §4.6).
- The renderer consumes `layout.widgets[]`, mounts the corresponding widget component from the frontend widget registry, and passes each instance's `config` to the component.
- Unknown widget types (e.g. removed from the registry after a downgrade) render a clearly labeled placeholder ("Widget type 'foo' is no longer available") rather than crashing.
- Widgets that fail to fetch data show their existing per-widget error UI; one widget's failure does not affect others.
- Pull-to-refresh at the page level invalidates every widget's query and refetches (parity with today's dashboard).
- The renderer is responsive: at viewports below the grid's small breakpoint, widgets collapse to a single full-width column in row-major (top-to-bottom, left-to-right) order derived from their (x, y) placement.

### 4.5 Designer (Edit Mode)

- Entry: an "Edit" button in the dashboard page header, visible to any user who can edit this dashboard (all household members for household dashboards, the owner for user dashboards).
- Exit: "Save" persists the new layout; "Discard" reverts to server state; "Exit edit mode" warns if there are unsaved changes.
- On entering edit mode:
    - Widgets display a border, drag handle, resize handles (right, bottom, bottom-right corner), a gear icon to open config, and a trash icon to remove.
    - A widget palette drawer slides in from one side and lists every widget in the registry with name + short description. Dragging a palette item onto the grid adds it at the cursor's cell with the widget's default size and default config.
- Grid behavior:
    - 12 columns wide on desktop.
    - Widgets have integer `x`, `y`, `w`, `h` in grid cells.
    - Each widget type declares `minW`, `minH`, `maxW`, `maxH`, and `defaultW`, `defaultH` in the frontend registry; the designer enforces those bounds during resize.
    - Overlaps are not allowed; the grid engine (react-grid-layout or equivalent) handles collision resolution by shifting widgets downward.
    - Empty rows between widgets are preserved.
- Per-widget config:
    - Clicking the gear icon on a widget opens a side panel.
    - The config form is rendered from that widget's Zod schema (frontend-only); inputs are labeled and validated inline.
    - Closing the panel with "Apply" updates the in-memory layout; closing without "Apply" discards config changes for that instance only.
- Dashboard-level controls in edit mode: rename the dashboard, change its scope (Household ↔ Mine — only the owner of a user dashboard may convert to household; any member may convert a household dashboard to their own "Mine" *copy*, which creates a new dashboard rather than moving the original), delete the dashboard.
- Save is a single PATCH that writes the entire new `layout` and (if changed) `name`. No autosave.
- Editing is a disabled on viewports below a tablet breakpoint (e.g. < 768px); the renderer still works. A toast explains: "Editing is only supported on larger screens."

### 4.6 Default Dashboard Preference

- A per-user, per-household preference stored in `account-service` as `preferences.default_dashboard_id` (scoped by `tenant_id` + `user_id` + `household_id`).
- Accessed via existing preferences infrastructure — no new endpoint is required. A string field suffices.
- Semantics:
    - If set and the referenced dashboard still exists and is visible to the caller, `/app/dashboard` redirects there.
    - If unset or the referenced id is gone, the frontend falls back to the first household-scoped dashboard the user can see (by `sort_order`), then the first user-scoped dashboard. If none exist, the seeding flow in §4.8 runs and the result becomes the redirect target.
- Each dashboard row has a "Set as my default" action.

### 4.7 Reordering

- The sidebar supports drag reorder within each scope (household group, user group). Reordering persists by calling a bulk endpoint `PATCH /api/v1/dashboards/order` with `[{id, sort_order}, ...]`.
- Reordering household dashboards affects all household members; reordering user dashboards affects only the owner.

### 4.8 Seeding (First-Run Behavior)

- When the current user's `GET /api/v1/dashboards` returns zero household-scoped dashboards for the caller's household, the frontend calls `POST /api/v1/dashboards/seed` (described in §5) to create a household-scoped dashboard named "Home" whose `layout` replicates the current `DashboardPage`.
- Seeding is idempotent per household: if any household-scoped dashboard already exists, the endpoint is a no-op and returns 200 with the existing list.
- The seed layout is defined in code on the **frontend** (it is a registry-referencing JSON document — the backend does not need to know about individual widget types beyond the validation rules in §4.9). The frontend sends the seed layout to the backend as the body of the seed POST.
- The seeded "Home" dashboard is a normal, fully editable, fully deletable dashboard.
- Deleting the last household-scoped dashboard is allowed; the next visit to `/app/dashboard` re-runs seeding.

### 4.9 Validation Rules

Backend-enforced on POST and PATCH:
- `name`: required, 1–80 chars after trim.
- `scope`: at creation, `scope=household` → `user_id = NULL`; `scope=user` → `user_id = <caller>`. Cannot specify another user.
- `layout.version`: integer, currently must equal `1`.
- `layout.widgets`: array, 0–40 entries.
- Each widget instance:
    - `id`: client-generated UUID; unique within the layout; required on persistence.
    - `type`: must be a string from a backend allowlist (§6.4). Unknown types are rejected at 422.
    - `x`, `y`: integers ≥ 0.
    - `w`, `h`: integers ≥ 1.
    - `x + w ≤ 12`.
    - `config`: object, ≤ 4 KB JSON, nesting depth ≤ 5. Content is **not** type-checked by the backend — the frontend Zod schemas own that. The backend treats `config` as opaque but bounded.
- Overlapping widgets are **not** validated by the backend (the frontend grid engine prevents them). Document this as intentional — the backend stays registry-agnostic.
- Total persisted `layout` JSON payload must be ≤ 64 KB.

### 4.10 Permissions

- List/read: any member of the household can read household dashboards; only the owner can read their user dashboards.
- Create household: any member.
- Create user: any member (for themselves).
- Update household: any member.
- Update user: owner only.
- Delete: same as update.
- Change scope household → mine: creates a copy owned by the caller; original is untouched.
- Change scope mine → household: owner only; promotes the row (clears `user_id`).

## 5. API Surface

All endpoints live on `dashboard-service`, under `/api/v1/dashboards`, and follow the existing JSON:API conventions (see other services' `docs/rest.md`).

Resource type: `dashboards`.

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/api/v1/dashboards` | List dashboards visible to caller for current household |
| `GET` | `/api/v1/dashboards/{id}` | Fetch one |
| `POST` | `/api/v1/dashboards` | Create |
| `PATCH` | `/api/v1/dashboards/{id}` | Update (name, layout, sort_order) |
| `DELETE` | `/api/v1/dashboards/{id}` | Delete |
| `PATCH` | `/api/v1/dashboards/order` | Bulk reorder within a scope |
| `POST` | `/api/v1/dashboards/{id}/promote` | Owner-only: convert mine → household |
| `POST` | `/api/v1/dashboards/{id}/copy-to-mine` | Copy household dashboard into caller's user scope |
| `POST` | `/api/v1/dashboards/seed` | Idempotent seed of "Home" household dashboard |

Detailed request/response shapes, including JSON:API resource mappings and error codes, are in `api-contracts.md`.

The default-dashboard preference uses existing `account-service` preferences endpoints (`GET /api/v1/preferences`, `PATCH /api/v1/preferences`). A new preference key `default_dashboard_id` (string) is added — no new endpoint is required.

## 6. Data Model

See `data-model.md` for the full schema including the GORM entity and the layout JSON contract. Summary:

### 6.1 `dashboard.dashboards` table
- `id` UUID PK
- `tenant_id` UUID, indexed
- `household_id` UUID, indexed
- `user_id` UUID NULL, indexed
- `name` varchar(80)
- `sort_order` int, default 0
- `layout` jsonb
- `schema_version` int, default 1
- `created_at`, `updated_at` timestamptz

Indexes:
- `(tenant_id, household_id, user_id)` covering list queries
- `(tenant_id, household_id) WHERE user_id IS NULL` for household fast-path

### 6.2 Account preferences addition
- New preference key `default_dashboard_id` (string, nullable) per the existing key/value preferences model in `account-service`. No migration beyond a documented key.

### 6.3 Layout JSON
```json
{
  "version": 1,
  "widgets": [
    {
      "id": "b9f0...",
      "type": "tasks-summary",
      "x": 0, "y": 0, "w": 4, "h": 2,
      "config": { "status": "pending", "title": "Pending Tasks" }
    }
  ]
}
```

### 6.4 Widget registry (v1)

The allowlist below is enforced by backend validation. Per-widget config schemas are owned by the frontend (Zod). See `data-model.md` for the field-by-field config shape; the summaries here are non-exhaustive.

| Type id | Scope | Default size (w×h) | Configurable |
|---|---|---|---|
| `weather` | household | 12×3 | units (`imperial`\|`metric`), optional location override |
| `tasks-summary` | household | 4×2 | `status` filter (`pending`\|`overdue`\|`completed`), optional title |
| `reminders-summary` | household | 4×2 | `filter` (`active`\|`snoozed`\|`upcoming`), optional title |
| `overdue-summary` | household | 4×2 | optional title (thin wrapper over tasks-summary with status=overdue — kept as a distinct type to preserve the existing card semantics) |
| `meal-plan-today` | household | 4×3 | `horizonDays` (1\|3\|7) |
| `calendar-today` | household | 6×3 | `horizonDays` (1\|3\|7), `includeAllDay` (bool) |
| `packages-summary` | household | 4×3 | optional title |
| `habits-today` | user-render | 4×3 | optional title |
| `workout-today` | user-render | 4×3 | optional title |

"Scope" describes *where the widget's data comes from*: `household` widgets fetch household-scoped data; `user-render` widgets fetch the viewer's personal data regardless of the dashboard's scope.

## 7. Service Impact

| Service | Change |
|---|---|
| **dashboard-service** (new) | New Go service following the `services/<svc>/internal/<domain>` pattern. Owns `dashboard.dashboards`. Implements the endpoints in §5 plus retention-framework compatibility (see §8.4). Registers routes under `/api/v1/dashboards`. |
| **account-service** | Adds `default_dashboard_id` as a recognized preference key (documented; enforced only as "string"). No schema change if preferences are already key/value; otherwise minimal. |
| **frontend** | New `DashboardRendererPage` replacing the hard-coded `DashboardPage`; new `DashboardDesignerPage` (or edit-mode toggle on the renderer); widget registry module mapping `type` strings to React components + Zod config schemas; sidebar integration; seeding bootstrap; sidebar reorder; new dashboard modal; default-dashboard preference wiring. `DashboardPage.tsx` is deleted once parity is demonstrated. |
| **ingress / nginx** | Route `/api/v1/dashboards` to the new service. |
| **docker-compose / k8s** | Add the new service definition, image, env, and DB schema entry. |

No changes required to the services that back existing widgets — all widget data fetches continue to hit their current endpoints.

## 8. Non-Functional Requirements

### 8.1 Performance
- `GET /api/v1/dashboards` must return within 200 ms p95 for a household with ≤ 20 dashboards (the hard cap per household).
- The renderer must not block on dashboard metadata: widgets should begin fetching their own data as soon as their component mounts, in parallel (already the current behavior).

### 8.2 Security / Multi-tenancy
- Every query scoped by `tenant_id` + `household_id`. User-scoped rows additionally filtered by `user_id`.
- Row-level authorization enforced in the service layer (no reliance on frontend filtering).
- Layout JSON size cap (64 KB) and widget count cap (40) prevent abuse.
- No raw HTML or scripts are permitted in any widget config field; string fields have length caps (≤ 80 chars for title, ≤ 200 chars for location label).

### 8.3 Observability
- Structured logs with `tenant_id`, `household_id`, `user_id`, `dashboard_id` on every mutation.
- Metrics: `dashboard_save_total`, `dashboard_save_duration_seconds`, `dashboard_validation_failure_total` (labeled by reason).

### 8.4 Retention & Account Deletion
- Dashboards are user/household content and fall under the retention framework. Add a category (e.g. `dashboards`) with a conservative default (no automatic deletion — retention acts only on manual household deletion cleanup). Implementation follows the `shared/go/retention` pattern used by other services (see `docs/architecture.md` §19). Retention coverage is a v1 requirement; specific windows can be zero for v1 (no automatic purge) and revisited later.
- Account deletion hard-deletes every user-scoped dashboard owned by the deleted user across every household they belonged to, in the same flow that tears down other user-scoped data. As part of the same flow, any `default_dashboard_id` preference whose value is one of the deleted dashboards is cleared (set to null) so no stale references survive. Household-scoped dashboards are untouched when a single member is removed; they are only removed when the household itself is deleted.

### 8.5 Compatibility
- `schema_version` is persisted on every row and every layout document so future breaking changes to the layout format can be migrated without data loss.
- The frontend refuses to render a dashboard whose `schema_version` is newer than it understands, and shows a "please refresh your browser" empty state.

## 9. Resolved Decisions

- **Account-delete cascade.** When a user account is deleted, every user-scoped dashboard they own (across every household) is hard-deleted immediately as part of the account-delete flow. Any `default_dashboard_id` preference rows still pointing at a deleted dashboard id are nulled out in the same transaction/flow so no stale references remain.
- **`copy-to-mine` placement.** The new user-scoped copy is appended at the end of the caller's user-scope list (`sort_order = max(sort_order) + 1` within that scope). Users reorder via drag.
- **Duplicate-in-place.** Out of scope for v1. `copy-to-mine` covers the "fork this dashboard" case; a same-scope duplicate can be added later if demand appears.

## 10. Acceptance Criteria

- [ ] A new `dashboard-service` exists with the schema in §6 and the endpoints in §5, fully tested with unit and integration tests following the existing service pattern.
- [ ] `dashboard-service` is wired into docker-compose, ingress, and the build pipeline; `scripts/local-up.sh` brings it up.
- [ ] `GET /api/v1/dashboards` on a brand-new household returns an empty list; the frontend then calls `POST /api/v1/dashboards/seed` and receives the seeded "Home" dashboard.
- [ ] The seeded "Home" dashboard renders the same widgets in the same positions as the current `DashboardPage` produces today (manual visual parity plus automated snapshot of the renderer given the seed JSON).
- [ ] `DashboardPage.tsx` is deleted and `/app/dashboard` redirects to the user's default dashboard (or the first household dashboard if unset).
- [ ] Users can create additional dashboards at household scope and at user scope via a "+ New dashboard" modal in the sidebar.
- [ ] The sidebar lists every dashboard visible to the caller as its own entry, grouped by scope, in `sort_order`.
- [ ] Drag-to-reorder in the sidebar persists via `PATCH /dashboards/order`.
- [ ] Entering edit mode on a dashboard (tablet viewport or larger) shows drag handles, resize handles, a gear config icon, and a remove icon on every widget.
- [ ] Adding a widget from the palette places it on the grid at default size with default config.
- [ ] Resizing enforces each widget's `minW`/`minH`/`maxW`/`maxH` and the 12-column grid.
- [ ] Each widget's config panel renders from its Zod schema and rejects invalid input inline.
- [ ] "Save" persists the layout and config via a single PATCH; "Discard" reverts cleanly; closing with unsaved changes warns the user.
- [ ] Each of the nine v1 widgets renders correctly through the registry-driven renderer and through the designer.
- [ ] User-scoped dashboards are not visible to other household members (verified by auth test fixtures).
- [ ] Household-scoped dashboards are editable by any household member (verified).
- [ ] Deleting a dashboard that is the user's default clears the preference; visiting `/app/dashboard` falls back to the seeding/first-dashboard flow.
- [ ] "Copy to mine" creates a deep copy in the caller's user scope with a new id, appended to the end of their user scope `sort_order` (`max + 1`).
- [ ] Account deletion hard-deletes the deleted user's user-scoped dashboards in every household and clears any `default_dashboard_id` preference rows that pointed at them.
- [ ] Below the tablet breakpoint, the renderer stacks widgets vertically in row-major order and the "Edit" action is disabled with a clear tooltip/toast.
- [ ] Unknown widget types render a labeled placeholder rather than crashing.
- [ ] Backend rejects: unknown widget `type`, `x+w > 12`, `w < 1` or `h < 1`, `name` out of range, `layout` > 64 KB, > 40 widgets.
- [ ] Backend accepts any well-formed `config` object within size/depth caps without caring about its fields.
- [ ] Retention category registered for `dashboards` (window can be 0/unlimited for v1 but the plumbing exists).
- [ ] `services/dashboard-service/docs/{domain.md,rest.md,storage.md}` exist per the DOCS.md contract.
