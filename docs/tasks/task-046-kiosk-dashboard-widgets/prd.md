# Kiosk Dashboard Widgets — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-05-01
---

## 1. Overview

The `initial-commit` branch contained a standalone `apps/kiosk` Next.js app that rendered a 4-column "what's happening today and tomorrow" dashboard intended for an always-on tablet. That app is not coming back as a separate surface, but the *content ideas* on its six cards are still useful: today's meals (with a multi-day dinner peek), today's tasks (overdue + today), today's weather, today's schedule, a tomorrow preview, and active reminders. Today's main-app dashboard (delivered by `task-045-dashboard-designer`) already has a JSON-backed widget registry, a designer, and a seeded "Home" dashboard, but it doesn't have list-based "today" variants of tasks or reminders, nor a tomorrow-preview concept.

This feature pulls those kiosk content ideas forward as **new widgets in the existing dashboard registry** and **adds a second seeded household dashboard** ("Kiosk") that arranges them in the spirit of the original kiosk's 4-column layout. No new backend service is introduced; we extend the existing `dashboard-service` widget allowlist and the frontend widget registry. The kiosk app itself, its device-context plumbing, polling/wake-lock hooks, and any fullscreen / always-on behaviors are explicitly out of scope.

The new widgets are intentionally **read-only** (in line with the task-034 dashboard philosophy that widgets surface state and link out for interaction) and **composable** (the kiosk's compound "Tomorrow" card splits into three independent widgets the user can place however they like). One existing widget (`meal-plan-today`) is extended via a new config knob to provide the kiosk's "today's full B/L/D + N-day dinner peek" view rather than introducing a parallel widget.

## 2. Goals

Primary goals:
- Add three new widgets to the dashboard registry: `tasks-today`, `reminders-today`, and three small tomorrow-preview widgets (`weather-tomorrow`, `calendar-tomorrow`, `tasks-tomorrow`)
- Extend the existing `meal-plan-today` widget with a `view` config option that supports a kiosk-style "today detail + N-day dinner peek" rendering
- Seed a second household dashboard named "Kiosk" on first-run alongside "Home", so brand-new households see both dashboards in their sidebar without doing anything
- Make the seeded "Kiosk" layout faithful in spirit to the original kiosk's 4-column composition (full-height columns of widgets)
- Keep all new widgets read-only — clicking through navigates to the relevant feature page

Non-goals:
- Reviving `apps/kiosk`, its `DeviceContext`, `useScreenWakeLock`, or any kiosk-mode runtime behaviors
- A separate `/app/kiosk` route, full-screen toggle, or device pairing
- Inline mutating actions on widgets (no inline task complete, no inline reminder snooze/dismiss). These were in the kiosk and are intentionally dropped — see §9 for the rationale
- Per-user theming, font-size, or "presentation mode" knobs
- Push or real-time updates; standard React Query refetch on mount/window-focus is sufficient
- A standalone `meal-plan-rolling` widget — the multi-day-dinner-peek view ships as a config of `meal-plan-today`

## 3. User Stories

- As a household member, I want a "Kiosk" dashboard already in my sidebar on day one, so I can see a glance-style layout without having to assemble it myself
- As a household member, I want a tasks widget that shows today's incomplete tasks and any overdue ones together, so I can see what's actually demanding attention without scrolling a count card
- As a household member, I want an active-reminders list widget so I can see what's queued, separate from the count-only summary card
- As a household member, I want to see "what's tomorrow shaping up to be?" — weather, events, and pending tasks — without having to switch to those pages
- As a household member, I want the meal-plan widget to show today's full breakfast/lunch/dinner *and* the next few days' dinners so I can plan groceries
- As a designer (in edit mode), I want each tomorrow-preview widget to be its own draggable tile so I can mix-and-match (e.g., drop just `weather-tomorrow` next to today's weather)
- As a household member, when I delete the seeded Kiosk dashboard, I don't want it to keep coming back

## 4. Functional Requirements

### 4.1 New Widget: `tasks-today`

Renders an at-a-glance list of the caller's tasks relevant to today.

- **Data source**: existing productivity-service endpoints already used by the dashboard (`GET /api/v1/tasks` filtered for the household; client filters in line with how `overdue-summary` and `tasks-summary` work today). No new backend.
- **Rendering**:
    - Section "Overdue (N)" listing every incomplete task with `due_at < today`. Hidden if N=0.
    - Section "Today" listing every incomplete task with `due_at` falling within the caller's local day. If empty *and* there are completed-today tasks, show "All tasks completed!" plus the completed list (visually de-emphasized). If empty and no completed-today, show "No tasks for today".
    - Each row shows: task title, the assignee's display name (if multi-user household), and an overdue indicator when applicable.
    - Each row is a link to `/app/tasks` (or to the task's detail route if one exists; if not, the list page is acceptable).
    - The widget header is a link to `/app/tasks`.
- **Interactivity**: read-only. No inline complete-toggle.
- **Config schema** (Zod):
    - `title?: string (max 80)` — overrides the default "Today's Tasks" header.
    - `includeCompleted?: boolean` (default `true`) — when `false`, the "all completed" fallback is suppressed and the widget shows "No tasks for today" instead of listing completed-today tasks.
- **Default size**: `w: 4, h: 4`. **Min**: `w: 3, h: 2`. **Max**: `w: 6, h: 8`.
- **Data scope**: `household` (any member sees the same household-wide task list, consistent with existing dashboard semantics).
- **Empty / error / loading states**: consistent with existing dashboard widgets — skeleton on load, inline error banner on failure that doesn't crash neighbors, friendly empty copy.

### 4.2 New Widget: `reminders-today`

Renders an at-a-glance list of the caller's currently active reminders.

- **Data source**: existing productivity-service reminders endpoint (the same one `reminders-summary` uses). Filter to reminders with `status = active`.
- **Rendering**:
    - Each row shows: reminder name, optional description, and a relative time indicator (e.g., "in 25 min", "Now").
    - List is sorted ascending by `remindAt`.
    - The widget header is a link to `/app/reminders` (or whatever route the existing reminders feature uses — match `reminders-summary`).
- **Interactivity**: read-only. No inline snooze/dismiss.
- **Config schema** (Zod):
    - `title?: string (max 80)`
    - `limit?: number (1–10, default 5)` — caps how many reminders render.
- **Default size**: `w: 3, h: 4`. **Min**: `w: 3, h: 2`. **Max**: `w: 6, h: 8`.
- **Data scope**: `household` (matches `reminders-summary`).
- **Empty state**: bell icon + "No active reminders".

### 4.3 New Widget: `weather-tomorrow`

A small read-only card showing tomorrow's high/low.

- **Data source**: same weather endpoint as the existing `weather` widget. Pick the daily forecast entry whose `date` equals tomorrow's local date.
- **Rendering**: cloud-icon, "Tomorrow" header with the formatted date, high/low temperatures formatted in the household's preferred unit (resolved exactly the way `weather` does — see `weather-adapter.tsx`).
- **Config schema** (Zod):
    - `units?: "imperial" | "metric"` — defaults to `null` meaning "inherit from household preference" (mirrors `weather` widget). If `weather` widget already has a different default, match that.
- **Default size**: `w: 3, h: 2`. **Min**: `w: 2, h: 2`. **Max**: `w: 6, h: 3`.
- **Data scope**: `household`.
- **Empty / error**: if the daily forecast doesn't include tomorrow (no data yet), show "Tomorrow's forecast not available". Errors render the existing weather error pattern.

### 4.4 New Widget: `calendar-tomorrow`

Tomorrow's calendar events.

- **Data source**: existing `GET /api/v1/calendar/events` filtered to tomorrow's local-day window. Reuse the same query plumbing `calendar-today` uses; this widget is essentially `calendar-today` shifted by one day with no `horizonDays` knob.
- **Rendering**: list events sorted by start time; all-day events first; show title, start–end (or "All day"), and the event's calendar/owner indicator (consistent with `calendar-today`). Header links to `/app/calendar`.
- **Config schema** (Zod):
    - `includeAllDay?: boolean` (default `true`)
    - `limit?: number (1–10, default 5)` — caps rendered count; if more, show "+N more".
- **Default size**: `w: 4, h: 3`. **Min**: `w: 3, h: 2`. **Max**: `w: 12, h: 6`.
- **Data scope**: `household`.
- **Empty state**: "No events tomorrow".

### 4.5 New Widget: `tasks-tomorrow`

Tomorrow's pending tasks.

- **Data source**: same task source as `tasks-today`. Client filters to incomplete tasks where `due_at` falls in tomorrow's local-day window.
- **Rendering**: simple list of titles (and assignee name when multi-user). Header links to `/app/tasks`.
- **Config schema** (Zod):
    - `title?: string (max 80)`
    - `limit?: number (1–10, default 5)` — caps rendered count; if more, show "+N more".
- **Default size**: `w: 3, h: 3`. **Min**: `w: 3, h: 2`. **Max**: `w: 6, h: 6`.
- **Data scope**: `household`.
- **Empty state**: "No tasks for tomorrow".

### 4.6 Extension: `meal-plan-today` gains a `view` config

Today, `meal-plan-today` accepts `horizonDays: 1 | 3 | 7`. The kiosk's MealsCard split into two visual sections: "Today" (full B/L/D) and "This Week" (next 3 days, dinners only). To support that view without spawning a duplicate widget:

- **New config field**: `view?: "list" | "today-detail" (default "list")`.
    - `list` — preserves today's behavior unchanged (uses `horizonDays` to flatten the list).
    - `today-detail` — renders two visual sections: "Today" with all populated slots (breakfast / lunch / dinner / snack / side, in canonical order), and "Next N days" with one row per day showing only the dinner. `N` is taken from `horizonDays` (so `horizonDays: 3` gives 3 follow-up days; `horizonDays: 1` collapses to just the today section; `horizonDays: 7` shows 7).
- **Backwards compat**: existing layouts that don't specify `view` continue to render as `list`. The Zod schema keeps `view` optional with a default. The seed-layout fixture for "Home" remains unchanged. The validator's per-widget config size cap (4 KB) is unaffected.
- **Default values stay the same** (`horizonDays: 1`, `view: "list"`). A migration of existing rows is **not** required.

### 4.7 Backend Allowlist Update

The shared widget-type allowlist gains four new entries. The fifth entry is the extended `meal-plan-today` whose key doesn't change.

| File | Change |
|---|---|
| `shared/go/dashboard/types.go` | Add `tasks-today`, `reminders-today`, `weather-tomorrow`, `calendar-tomorrow`, `tasks-tomorrow` to `WidgetTypes` map |
| `frontend/src/lib/dashboard/widget-types.ts` | Add the same five strings to `WIDGET_TYPES` |
| `shared/go/dashboard/fixtures/widget-types.json` (or wherever the parity fixture lives) | Mirror the additions so the parity test passes |

No other backend code changes are required: the validator delegates to `IsKnownWidgetType`, and `config` is opaque from the backend's perspective (per the existing dashboard-designer contract).

### 4.8 Frontend Widget Registry Updates

For each new widget, add a file under `frontend/src/lib/dashboard/widgets/<name>.ts` exporting a `WidgetDefinition<TConfig>`, plus a corresponding adapter component under `frontend/src/components/features/dashboard-widgets/<name>-adapter.tsx`. Wire each into `widget-registry.ts`'s `widgetRegistry` array.

Each adapter follows the patterns set by the existing nine adapters (`weather-adapter.tsx`, `meal-plan-adapter.tsx`, etc.):
- Use TanStack React Query hooks already in `frontend/src/services/api/`. New hooks are added only when no equivalent already exists.
- Use the standard skeleton, empty-state, and error UI patterns. One widget's failure must not affect siblings.
- Respect the household-vs-user data-scope contract from the dashboard-designer PRD.

### 4.9 Seeded "Kiosk" Dashboard

A second household-scoped dashboard named **"Kiosk"** is seeded alongside the existing "Home" dashboard on first-run.

#### 4.9.1 Layout Shape

Faithful in spirit to the original kiosk: a 4-column-wide composition where each column is a "full-height" stack of related widgets. Concretely, on the 12-column grid:

| Column (3 grid units) | Widgets stacked top-to-bottom |
|---|---|
| Col 1 (`x: 0, w: 3`) | `weather`, `meal-plan-today` (`view: "today-detail"`, `horizonDays: 3`), `tasks-today` |
| Col 2 (`x: 3, w: 3`) | `calendar-today` (full vertical extent of the column) |
| Col 3 (`x: 6, w: 3`) | `weather-tomorrow`, `calendar-tomorrow`, `tasks-tomorrow` |
| Col 4 (`x: 9, w: 3`) | `reminders-today` (full vertical extent of the column) |

"Full-height" means each column's widgets together fill a target height (proposed: `h` totals of 12 grid units per column; per-widget heights are tuned during design). The exact `h` tuning is finalized in `design.md`. Mobile / narrow viewports inherit the renderer's existing single-column row-major collapse from task-045 — no special handling.

#### 4.9.2 Seeding Mechanism

The current `POST /api/v1/dashboards/seed` endpoint creates a single household-scoped dashboard and is a no-op when *any* household-scoped dashboard already exists. Extending it to a second seeded dashboard requires a behavior change. Two acceptable mechanisms (final selection deferred to `design.md`):

1. **Per-key seeding (preferred)**: add an optional `key: string` field to the seed request body and persist it on the row (or on a tiny `seeded_keys` table). The endpoint becomes idempotent per `(tenant, household, key)`. The frontend issues two seed calls on first-run: `key=home` and `key=kiosk`. Re-deletion of either is permanent.
2. **Frontend-only orchestration**: keep `/seed` as-is for "Home"; on first-run the frontend additionally calls `POST /api/v1/dashboards` (regular create) for "Kiosk" only when no dashboard named "Kiosk" exists in the visible list. To avoid the "delete-and-it-comes-back" loop, the frontend records a per-(user, household) preference in `account-service` (`preferences.kiosk_dashboard_seeded: true`) the first time it sees any dashboard list returned (whether it created one or not). This means deletion is permanent because the preference flag stays set.

Both options keep the kiosk dashboard fully editable and deletable just like Home. The PRD constrains *behavior* (idempotent first-run, deletion is permanent, brand-new households get both dashboards in the sidebar); the implementation choice is a design decision.

#### 4.9.3 Brownfield Households

Households that already exist (already have a "Home" dashboard) when this feature ships **must** still receive the seeded "Kiosk" dashboard the next time any member loads the app. This is non-negotiable — otherwise the feature is invisible to everyone except brand-new tenants.

If option 1 is chosen, this falls out naturally: the frontend issues both seed calls on every load; the second one creates the row on existing households. If option 2 is chosen, the frontend's "no Kiosk dashboard exists yet AND preference flag is unset" check works for both new and existing households.

#### 4.9.4 Sidebar Order

The seeded Kiosk dashboard's `sort_order` is `1` (Home stays at `0`). It appears second in the household-scope sidebar group on first render. Users may reorder freely via the existing drag-reorder mechanism from task-045.

### 4.10 Validation

All new widgets pass the existing layout validator (it's opaque-config — backend only checks `type` is in the allowlist and the widget instance fits grid bounds). Frontend Zod schemas for new widgets are unit-tested with the same harness used for existing ones.

### 4.11 Telemetry / Analytics

None required beyond what the existing dashboard system emits. If existing dashboard events log widget types, the new five types will appear automatically.

## 5. API Surface

No new resource type. The only API contract change is to `POST /api/v1/dashboards/seed`, *if* design.md selects mechanism (1) above:

- **Request body** gains an optional `key: string` field (1–40 chars, lowercase + hyphen, e.g. `"home"`, `"kiosk"`).
- **Behavior**: idempotent per `(tenant, household, key)`. Omitting `key` keeps the existing "any-household-scoped-dashboard counts" semantics for backward compatibility with currently-deployed clients.
- **Response shape unchanged**: same `SeedResult`.

If design.md selects mechanism (2), there is no API change, only a new preference key in `account-service`.

Detailed request/response shapes for the seed endpoint adjustment go in `api-contracts.md`.

## 6. Data Model

No new entities. Possible adjustments depending on the seeding mechanism chosen in design:

- **If mechanism (1)**: add a nullable `seed_key VARCHAR(40)` column to `dashboard.dashboards` with a partial unique index `(tenant_id, household_id, seed_key) WHERE seed_key IS NOT NULL`. Existing rows have `seed_key = NULL`. Migration: add column + index, no backfill.
- **If mechanism (2)**: add a string preference key `kiosk_dashboard_seeded` to the `account-service` preferences key registry — no schema change since preferences are already key/value.

The widget `config` payloads for the new widgets fit comfortably under the per-widget 4 KB cap.

## 7. Service Impact

| Service | Change |
|---|---|
| **Frontend** | New widget definitions + adapters; updated `widget-registry.ts`; updated `seed-layout.ts` (or a parallel `kiosk-seed-layout.ts`); updated `meal-plan-today` Zod schema and its adapter to support `view: "today-detail"`; new query hooks where missing (likely tomorrow's calendar and tomorrow's tasks); minor seed-orchestration logic per §4.9.2 |
| **dashboard-service** | Allowlist additions in `shared/go/dashboard/types.go`; possibly a `seed_key` column + migration + processor change (if mechanism 1) |
| **account-service** | Possibly a new preference key (if mechanism 2) |
| **All other services** | None |

## 8. Non-Functional Requirements

- **Performance**: each new widget fetches independently with React Query; reuse existing keys where the data overlaps (e.g., `tasks-today` and `tasks-tomorrow` should share the same task list query rather than firing two separate requests).
- **Multi-tenancy**: all data fetches go through existing services that already enforce tenant scoping; no new tenant plumbing.
- **Accessibility**: widget headers are real links (not divs), list items have proper semantics, empty states have descriptive copy. Match the patterns established by existing widgets.
- **Mobile**: no special handling — the dashboard renderer already collapses to a single column on narrow viewports per task-045.
- **Security**: no new endpoints, no new data exposure. Widgets render only what existing endpoints already authorize the caller to see.
- **Observability**: dashboard-service logs unknown-widget-type rejections; the new types must not appear there once added to the allowlist (parity test enforces this).
- **Backward compatibility**: existing dashboards (including the current "Home" seed) continue to render unchanged. The `meal-plan-today` extension is additive: layouts with no `view` field render exactly as before.

## 9. Open Questions

These are deferred to `design.md`:

1. **Seeding mechanism**: per-key on the seed endpoint (4.9.2 option 1) vs. frontend-orchestrated with a preference flag (option 2). Trade-offs: option 1 is cleaner long-term and supports future seeded templates; option 2 ships with zero backend changes but introduces a frontend-side guard against re-creation.
2. **Exact column heights** for the seeded Kiosk layout (target column total height in grid units).
3. **Whether to share query plumbing**: should `tasks-today` and `tasks-tomorrow` use a single React Query key + client-side filter, or separate keys? Same question for `calendar-today` vs. `calendar-tomorrow`.
4. **Tomorrow widget date semantics**: "tomorrow" relative to *what timezone*? Use the same logic as the existing `calendar-today` and meal-plan-today widgets (almost certainly the household preferred timezone, falling back to browser local). `design.md` confirms by reading the existing helpers.
5. **Inline interactivity reconfirmation**: this PRD locks in read-only per the user's direction, but if `design.md` surfaces a strong reason to reintroduce inline complete/dismiss, that's a scope-change conversation, not a silent decision.

## 10. Acceptance Criteria

- [ ] `tasks-today` widget appears in the designer palette, can be added to a dashboard, fetches and renders today's incomplete tasks plus overdue tasks, with read-only rows and a header link to `/app/tasks`
- [ ] `reminders-today` widget appears in the palette, renders the active reminders list capped by `limit`, with a header link to the reminders feature page
- [ ] `weather-tomorrow` widget renders tomorrow's daily high/low using the household's preferred temperature unit; falls back gracefully when tomorrow's forecast isn't yet available
- [ ] `calendar-tomorrow` widget renders tomorrow's events sorted by start time with all-day events first; capped by `limit` with a "+N more" indicator
- [ ] `tasks-tomorrow` widget renders incomplete tasks with `due_at` falling in tomorrow's local-day window; capped by `limit`; assignee shown for multi-user households
- [ ] `meal-plan-today` accepts `view: "today-detail"` and renders today's full B/L/D + an N-day dinner peek (N = `horizonDays`); `view: "list"` (and absence of `view`) preserves the current behavior byte-for-byte
- [ ] All five new widget types are added to both `shared/go/dashboard/types.go` and `frontend/src/lib/dashboard/widget-types.ts`; the parity fixture is updated; the parity test passes
- [ ] Brand-new households see two dashboards in the sidebar after first load: "Home" and "Kiosk"
- [ ] Existing households (those that already have a "Home" dashboard at deploy-time) receive the "Kiosk" dashboard the next time any member loads the app
- [ ] Deleting the seeded "Kiosk" dashboard is permanent — it does not re-seed on subsequent loads
- [ ] All new widgets handle loading, empty, and error states independently — one widget failing does not affect siblings
- [ ] All new widgets and the `meal-plan-today` extension have unit tests covering the Zod schema, default config, and rendering of empty / loaded / error states
- [ ] The dashboard-renderer-parity test passes against the updated seed layouts
- [ ] No regressions on the existing "Home" dashboard rendering or layout
