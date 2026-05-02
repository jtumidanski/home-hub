# Kiosk Dashboard Widgets — Design

Version: v1
Status: Approved (decisions for §9 questions)
Created: 2026-05-01

This document records the design decisions made on top of `prd.md`, `api-contracts.md`, and `data-model.md`. The PRD's open questions in §9 are all closed; values are restated here as the source of truth for implementation.

## 1. Decisions Locked from PRD §9

| # | Question | Decision |
|---|---|---|
| 1 | Seeding mechanism (per-key seed vs frontend-orchestrated preference) | **Per-key seed for race-safe creation, plus a one-bit preference flag (`kiosk_dashboard_seeded`) as the "stop trying" gate.** PRD mechanism 1 is the API contract; a single preference flag prevents resurrection-after-delete (see §6.4). |
| 2 | Shared query plumbing for today/tomorrow widget pairs | **Share queries, client-filter by date.** Tasks share `useTasks()`; calendar shares one query keyed by date. |
| 3 | "Tomorrow" timezone semantics | Household timezone first, browser local fallback (matches `useLocalDate`). New `useLocalDateOffset(tz, n)` hook for any future "today + n days" needs. |
| 4 | Exact column heights for the seeded Kiosk layout | Each column totals `h: 12`. Per-widget heights detailed in §6.2. |
| 5 | Read-only confirmation | Read-only by **convention**, not machinery. Adapters import only query hooks; no registry-level flag, no runtime guard. |

## 2. Architecture Overview

The feature splits across three edit zones, listed in increasing blast radius:

1. **Shared widget allowlist** — five new strings appended to `shared/go/dashboard/types.go`, the same five appended to `frontend/src/lib/dashboard/widget-types.ts`, and mirrored in the parity fixture. The validator already delegates to `IsKnownWidgetType`, so no validator changes.
2. **Frontend widgets and seed orchestration** — five new `WidgetDefinition` files plus five adapter components, a `view` extension on `meal-plan-today`, a new `kiosk-seed-layout.ts`, and a small change to `DashboardRedirect.tsx` to issue two seed calls instead of one. One new hook (`useLocalDateOffset`).
3. **`dashboard-service` per-key seed** — additive `seed_key VARCHAR(40)` column on `dashboards`, a partial unique index, an `Entity` field, a one-time backfill in the same migration, and a reworked `Processor.Seed` whose idempotency key is `(tenant, household, seed_key)` rather than "any household-scoped row exists".

No new services. No new endpoints (the existing `POST /api/v1/dashboards/seed` extends its body, not its URL). One additive preference key in `account-service` (`kiosk_dashboard_seeded`) — no schema change, the preferences store is already a key/value bag. See §6.4 for why this flag is needed alongside the per-key seed.

## 3. dashboard-service Changes

### 3.1 Schema and Migration

`services/dashboard-service/internal/dashboard/entity.go`:

```go
type Entity struct {
    // existing fields...
    SeedKey *string `gorm:"column:seed_key;type:varchar(40)"`
}

func Migration(db *gorm.DB) error {
    if err := db.AutoMigrate(&Entity{}); err != nil {
        return err
    }
    // existing partial index for household scope
    if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_dashboards_household_partial
        ON dashboards (tenant_id, household_id) WHERE user_id IS NULL`).Error; err != nil {
        return err
    }
    // new partial unique index for seed keys
    if err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_dashboards_seed_key
        ON dashboards (tenant_id, household_id, seed_key) WHERE seed_key IS NOT NULL`).Error; err != nil {
        return err
    }
    // brownfield backfill — claim existing seeded "Home" rows so the
    // new client's idempotent home-seed call is a no-op for them
    return db.Exec(`UPDATE dashboards
        SET seed_key = 'home'
        WHERE seed_key IS NULL
          AND user_id IS NULL
          AND sort_order = 0
          AND name = 'Home'`).Error
}
```

**Backfill heuristic.** The `WHERE` clause matches the exact signature of the original seed (`name = 'Home'`, `sort_order = 0`, `user_id IS NULL`). Households where a member has renamed the Home dashboard or moved it down the sidebar will not be backfilled. The frontend's existing first-run check (any household-scoped dashboard exists?) suppresses the home seed call for those households, so no new "Home" row gets created. The renamed pre-existing dashboard remains untouched. Acceptable trade-off: the seed key represents "this is the row I'd resurrect if you delete it", and a renamed row already opted out of that behavior by being renamed.

### 3.2 Processor

`Processor.Seed` rewritten:

```go
func (p *Processor) Seed(
    tenantID, householdID, callerUserID uuid.UUID,
    name string,
    seedKey *string,             // NEW — nil means "legacy any-row check"
    layoutJSON json.RawMessage,
) (SeedResult, error)
```

Behavior:

- **`seedKey != nil`** — acquire advisory lock on `(tenant, household, seedKey)`; SELECT for a row with that `seed_key`; if found, return `Created: false` with that row in `Existing`; otherwise INSERT with `SeedKey = seedKey` and return `Created: true`. The partial unique index is the backstop against races the advisory lock didn't catch.
- **`seedKey == nil`** — preserves the existing behavior verbatim: lock on `(tenant, household)`, count any household-scoped rows, no-op if `count > 0`. This branch is what currently-deployed clients hit before the frontend update lands.

Lock-key derivation extends the existing `seedLockKey` to mix in the seed key bytes:

```go
func seedLockKeyForKey(tenantID, householdID uuid.UUID, key string) int64 {
    var combined [32 + 40]byte
    copy(combined[:16], tenantID[:])
    copy(combined[16:32], householdID[:])
    copy(combined[32:], []byte(key))
    sum := sha256.Sum256(combined[:])
    return int64(binary.BigEndian.Uint64(sum[:8]))
}
```

`SeedResult` is unchanged.

### 3.3 REST Handler and Validation

`services/dashboard-service/internal/dashboard/resource.go` — the seed handler reads optional `attributes.key` from the JSON:API request body. Validation:

- `key` regex `^[a-z][a-z0-9-]{0,39}$` (matches PRD `api-contracts.md`).
- Empty string is treated as omitted (defensive), not as a validation error, but a present-but-malformed string returns `422 validation.invalid_field` with pointer `/data/attributes/key`.

The handler passes a `*string` (nil when absent) to `Processor.Seed`.

### 3.4 Allowlist Additions

`shared/go/dashboard/types.go` — five strings appended to `WidgetTypes`:

```go
"tasks-today":       {},
"reminders-today":   {},
"weather-tomorrow":  {},
"calendar-tomorrow": {},
"tasks-tomorrow":    {},
```

`shared/go/dashboard/fixtures/widget-types.json` — same five entries appended. The parity test in `types_test.go` enforces the Go map matches the fixture.

## 4. Frontend Widget Definitions

Each new widget gets two files following the existing pattern:

| Definition | Adapter |
|---|---|
| `frontend/src/lib/dashboard/widgets/tasks-today.ts` | `frontend/src/components/features/dashboard-widgets/tasks-today-adapter.tsx` |
| `frontend/src/lib/dashboard/widgets/reminders-today.ts` | `.../reminders-today-adapter.tsx` |
| `frontend/src/lib/dashboard/widgets/weather-tomorrow.ts` | `.../weather-tomorrow-adapter.tsx` |
| `frontend/src/lib/dashboard/widgets/calendar-tomorrow.ts` | `.../calendar-tomorrow-adapter.tsx` |
| `frontend/src/lib/dashboard/widgets/tasks-tomorrow.ts` | `.../tasks-tomorrow-adapter.tsx` |

Each definition exports `WidgetDefinition<TConfig>` with the Zod schema, default size, min/max, and the adapter component. All five are registered in `widget-registry.ts`.

### 4.1 Zod Schemas

Lifted unchanged from `data-model.md`:

```ts
// tasks-today
z.object({
  title: z.string().max(80).optional(),
  includeCompleted: z.boolean().default(true),
});

// reminders-today
z.object({
  title: z.string().max(80).optional(),
  limit: z.number().int().min(1).max(10).default(5),
});

// weather-tomorrow
z.object({
  units: z.enum(["imperial", "metric"]).nullable().default(null),
});

// calendar-tomorrow
z.object({
  includeAllDay: z.boolean().default(true),
  limit: z.number().int().min(1).max(10).default(5),
});

// tasks-tomorrow
z.object({
  title: z.string().max(80).optional(),
  limit: z.number().int().min(1).max(10).default(5),
});

// meal-plan-today (extension)
z.object({
  horizonDays: z.union([z.literal(1), z.literal(3), z.literal(7)]).default(1),
  view: z.enum(["list", "today-detail"]).default("list"),  // NEW
});
```

### 4.2 Adapters: Read-Only Convention

Each adapter's first non-import line is the comment:

```ts
// Read-only widget — no mutations. See PRD §4.x.
```

Adapters import only query hooks (`useTasks`, `useReminders`, `useWeather`, `useCalendarEventsForDate`, etc.) — never `useCreate*`, `useUpdate*`, `useDelete*`, or any mutation. Header rows are real `<Link>` components, never `<button onClick={...}>`. Empty states are presentational; no inline action buttons.

This is enforced by code review, not by machinery. The PRD locks read-only and §9 confirms it; if a future task wants inline actions, that's a scope-change conversation per PRD §9.5.

## 5. Frontend Data Flow

### 5.1 Tasks (today + tomorrow share one query)

`useTasks()` already returns the full household task list keyed on `["tasks", tenant, household, "list"]` with 5-min `staleTime`. Both `tasks-today-adapter` and `tasks-tomorrow-adapter` call `useTasks()` and slice client-side:

```ts
// tasks-today-adapter.tsx
const { household } = useTenant();
const today = useLocalDate(household?.attributes.timezone);
const { data } = useTasks();
const tasks = (data?.data ?? []);

const overdue = tasks.filter(t => !t.attributes.completed && t.attributes.dueOn && t.attributes.dueOn < today);
const todayTasks = tasks.filter(t => !t.attributes.completed && t.attributes.dueOn === today);
const completedToday = tasks.filter(t => t.attributes.completed && t.attributes.completedOn === today);
```

```ts
// tasks-tomorrow-adapter.tsx
const tomorrow = useLocalDateOffset(household?.attributes.timezone, 1);
const { data } = useTasks();
const tomorrowTasks = (data?.data ?? [])
  .filter(t => !t.attributes.completed && t.attributes.dueOn === tomorrow)
  .slice(0, config.limit);
```

When both widgets render on the kiosk dashboard, React Query dedupes — one network request serves both.

### 5.2 Calendar (today + tomorrow share keyed-by-date)

`calendar-today` currently goes through `CalendarWidget`, which has its own internal query. Both today and tomorrow will be parameterized by date and share React Query cache shape. Concretely:

- Introduce (or extract) `useCalendarEventsForDate(date: string)` keyed on `["calendar", tenant, household, "events", date]`.
- `CalendarWidget` (the today renderer) is refactored to consume this hook for its own date.
- `calendar-tomorrow-adapter` calls the same hook with the tomorrow date string.

The two queries are distinct cache entries (one per date string), but both are pure functions of the date. This is the simplest model — no over-fetching, no cross-day cache subtraction.

The PRD §4.4 explicitly says calendar-tomorrow is "essentially `calendar-today` shifted by one day." This refactor unifies that promise in code.

### 5.3 Reminders

`reminders-today-adapter` uses the existing reminders list query (the same one `reminders-summary` reads from). Filter to `status === "active"`, sort ascending by `remindAt`, slice to `config.limit`. No new hook unless one doesn't exist; if `useReminders()` already returns the active list, that's the call site.

Relative-time labels ("in 25 min", "Now") use the existing `formatRelativeTime` helper if present, or are computed inline from `remindAt - Date.now()`. The widget polls via React Query default `refetchOnMount`/`refetchOnWindowFocus`; no new interval required (PRD non-goal: real-time).

### 5.4 Weather

`weather-tomorrow-adapter` calls the same weather endpoint as `weather-adapter` (presumably `useWeather()` or equivalent), reads the daily forecast array, and picks the entry whose `date` equals the tomorrow date string. If absent, render "Tomorrow's forecast not available". Unit resolution mirrors `weather-adapter` exactly (PRD §4.3).

### 5.5 Date Hook

`frontend/src/lib/hooks/use-local-date-offset.ts`:

```ts
export function useLocalDateOffset(tz: string | undefined, offsetDays: number): string {
  const subscribe = useCallback((notify: () => void) => {
    const id = window.setInterval(notify, 60_000);
    return () => window.clearInterval(id);
  }, []);
  const getSnapshot = useCallback(() => getLocalDateStrOffset(tz, offsetDays), [tz, offsetDays]);
  return useSyncExternalStore(subscribe, getSnapshot);
}
```

A sibling helper `getLocalDateStrOffset(tz, n)` is added to `frontend/src/lib/date-utils.ts`, computed by formatting `now + n days` in the target IANA timezone using `Intl.DateTimeFormat`. Day-string comparison sidesteps DST cleanly because we never subtract timestamps across boundaries.

## 6. Seeded Kiosk Dashboard

### 6.1 Seed Layout File

`frontend/src/lib/dashboard/kiosk-seed-layout.ts`:

```ts
import type { Layout } from "@/lib/dashboard/schema";

function uuid(): string { return (crypto as Crypto).randomUUID(); }

export function kioskSeedLayout(): Layout {
  return {
    version: 1,
    widgets: [
      // Column 1 (x:0, w:3) — h totals 12
      { id: uuid(), type: "weather",          x: 0, y: 0,  w: 3, h: 3,  config: { units: "imperial", location: null } },
      { id: uuid(), type: "meal-plan-today",  x: 0, y: 3,  w: 3, h: 5,  config: { horizonDays: 3, view: "today-detail" } },
      { id: uuid(), type: "tasks-today",      x: 0, y: 8,  w: 3, h: 4,  config: {} },
      // Column 2 (x:3, w:3) — h totals 12
      { id: uuid(), type: "calendar-today",   x: 3, y: 0,  w: 3, h: 12, config: { horizonDays: 1, includeAllDay: true } },
      // Column 3 (x:6, w:3) — h totals 12
      { id: uuid(), type: "weather-tomorrow", x: 6, y: 0,  w: 3, h: 3,  config: {} },
      { id: uuid(), type: "calendar-tomorrow",x: 6, y: 3,  w: 3, h: 5,  config: {} },
      { id: uuid(), type: "tasks-tomorrow",   x: 6, y: 8,  w: 3, h: 4,  config: {} },
      // Column 4 (x:9, w:3) — h totals 12
      { id: uuid(), type: "reminders-today",  x: 9, y: 0,  w: 3, h: 12, config: {} },
    ],
  };
}
```

### 6.2 Column Heights Summary

| Column | x range | Widgets (top → bottom) | h breakdown | Total |
|---|---|---|---|---|
| 1 | 0–2 | weather, meal-plan-today (today-detail), tasks-today | 3 / 5 / 4 | 12 |
| 2 | 3–5 | calendar-today | 12 | 12 |
| 3 | 6–8 | weather-tomorrow, calendar-tomorrow, tasks-tomorrow | 3 / 5 / 4 | 12 |
| 4 | 9–11 | reminders-today | 12 | 12 |

Every widget's `h` is within its declared min/max from PRD §4. Every widget satisfies `x + w ≤ 12`.

### 6.3 Why Two Gating Signals Are Needed

Per-key seeding alone makes seeding **race-safe and idempotent within a single load**, but it does not prevent **resurrection on subsequent loads** after a user deletes the seeded row. The PRD acceptance criterion "Deleting the seeded Kiosk dashboard is permanent" requires a separate "stop trying" signal that survives row deletion.

For the **Home** dashboard the existing frontend already has that signal: "no household-scoped dashboard exists in the visible list → run the home seed". Once Home exists (or has existed and was renamed), the home seed is skipped; once Home is deleted by a user, the frontend's first-run path re-seeds it — which is the existing pre-task-046 behavior and not changed here.

For the **Kiosk** dashboard, "no household-scoped dashboards visible" no longer means "kiosk hasn't been seeded" (Home is also household-scoped). Two reasonable signals:

- A row with `seed_key='kiosk'` exists in the visible list → already seeded.
- A persistent preference flag that survives row deletion → "we've finished kiosk seeding for this user".

We need both. The row-existence check handles members who joined the household after kiosk was seeded (their preference is unset, but the row exists, so they don't re-seed). The preference flag handles the user who deleted Kiosk (the row is gone, but the flag stays, so we don't resurrect).

The flag lives in `account-service` as a new preference key `kiosk_dashboard_seeded` (boolean, default `false`, scoped per-(tenant, user, household), like other prefs). No schema migration — prefs are key/value.

### 6.4 Seeding Algorithm

In `DashboardRedirect.tsx` (or an extracted helper), on every load, after `useDashboards()` and `useHouseholdPreferences()` resolve:

1. **Compute `homeNeeded`**: no household-scoped dashboard exists in the visible list (existing check, unchanged).
2. **Compute `kioskNeeded`**: `preferences.kiosk_dashboard_seeded !== true` AND no row with `seed_key === 'kiosk'` exists in the visible list.
3. Issue the relevant seed calls in parallel:
   - if `homeNeeded`: `seed({ name: "Home", key: "home", layout: seedLayout() })`
   - if `kioskNeeded`: `seed({ name: "Kiosk", key: "kiosk", layout: kioskSeedLayout() })`
4. `await Promise.all(...)` and refetch the dashboard list.
5. After the refetch resolves, if **any** row with `seed_key === 'kiosk'` is now visible AND `preferences.kiosk_dashboard_seeded !== true`, `PATCH /api/v1/preferences` to set it `true`. (This fires whether or not *this* load issued the seed call, so members who joined after kiosk was already seeded still get their flag set on first observation.)
6. Resolve to the default dashboard following the existing redirect rules.

`useSeedDashboard()` mutation hook gains an optional `key: string` param; `dashboardService.seedDashboard` adds it to `attributes`. Existing test callers that pass no key continue to hit the legacy `seedKey == nil` branch in the processor.

**Brownfield walkthrough** (existing household with Home only, kiosk preference unset):
- Migration backfills `seed_key='home'` on the existing Home row.
- Next load: `homeNeeded = false` (Home is in the list); `kioskNeeded = true` (no `seed_key='kiosk'` row, preference unset).
- Kiosk seed fires, creates the row, returns `Created: true`.
- Frontend sets `kiosk_dashboard_seeded = true`.
- User deletes Kiosk later: row is gone, but preference is set, so `kioskNeeded = false` on subsequent loads. No resurrection.

**Member-joins-later walkthrough** (kiosk row already exists, new member's preference is unset):
- `kioskNeeded = false` because the row is present.
- Step 5 still fires: row is visible, preference is `false`, so the PATCH sets it `true`. Now if this member is later the one to delete Kiosk, no resurrection — their preference is set.

**Failure modes**:
- Home seed succeeds, kiosk seed fails: household has Home only; preference unset; next load retries kiosk. No infinite loop because exponential refetch isn't a thing here — it's just one attempt per page-load.
- Kiosk create succeeds, preference PATCH fails: row exists, but `kiosk_dashboard_seeded` is unset. Next load: row is visible → `kioskNeeded = false` → step 5's row-observation PATCH retries, setting the preference. The PATCH retries on every load until it sticks; once it does, deletion is permanent.

## 7. Allowlist & Parity Tests

`shared/go/dashboard/types_test.go` already validates the Go map matches the JSON fixture. Adding entries to both keeps that test green.

`frontend/src/pages/__tests__/dashboard-renderer-parity.test.tsx` currently asserts every type in the allowlist appears in `seedLayout()`. With two seed layouts, the assertion needs to be:

> Every type in `WIDGET_TYPES` appears in **at least one of** `seedLayout()` or `kioskSeedLayout()`.

This keeps the contract that registry adoption forces seed-fixture coverage.

## 8. meal-plan-today Extension

`frontend/src/lib/dashboard/widgets/meal-plan-today.ts` Zod schema gains `view` per §4.1. Default `"list"`.

`meal-plan-adapter.tsx` branches:

```ts
if (config.view === "today-detail") {
  return <MealPlanTodayDetail horizonDays={config.horizonDays} />;
}
return <MealPlanWidget horizonDays={config.horizonDays} />;
```

`MealPlanTodayDetail` is a new component (under `frontend/src/components/features/meals/meal-plan-today-detail.tsx`) that renders two sections:

- "Today" — every populated meal slot (breakfast / lunch / dinner / snack / side, in canonical order).
- "Next N days" — one row per follow-up day showing only the dinner slot. `N = horizonDays` (so `horizonDays: 1` collapses to today only, `3` shows 3 follow-up days, `7` shows 7).

Data-fetch: reuses the same meal-plan query plumbing the existing `MealPlanWidget` uses, just rendered differently. No new API.

Backwards compat: layouts persisted before this change have no `view` field; Zod's default makes them parse as `"list"`. Existing `seedLayout()` (Home) is unchanged. The validator's per-widget config size cap (4 KB) is unaffected.

## 9. Error and Empty States

Each adapter uses the existing skeleton, empty-state, and error-banner patterns from the nine current adapters:

- **Loading**: `<Skeleton>` placeholders sized to the widget's expected content.
- **Error**: an inline banner (text + retry affordance) that does not crash siblings. The dashboard renderer already isolates per-widget errors via boundary patterns; we follow that.
- **Empty**: friendly copy as specified in PRD §4 (e.g., "No active reminders", "No tasks for tomorrow", "Tomorrow's forecast not available").
- **All-completed-today fallback** (`tasks-today` only, when `includeCompleted: true`): "All tasks completed!" headline above a de-emphasized list of completed-today tasks.

One widget failing — e.g., the weather endpoint times out — must not affect siblings on the same dashboard. This is already true of the existing renderer; the new adapters just need to throw their errors locally rather than letting them propagate.

## 10. Testing Strategy

### 10.1 Unit Tests

Per new widget definition (5 files):

- Zod schema accepts valid configs, rejects invalid ones (per `data-model.md`).
- Default config matches the PRD's defaults.
- Min/max sizes match the PRD.

Per adapter (5 files): Vitest + React Testing Library, mocked query hooks:

- Loading state renders skeleton.
- Empty state renders the documented copy.
- Loaded state renders the expected items.
- Error state renders the banner without crashing.
- For `tasks-today`: overdue + today + completed-today branches all render correctly.
- For `weather-tomorrow`: the "tomorrow not in the forecast" branch renders the fallback.

For the `meal-plan-today` extension:
- `view: "list"` (and absence of `view`) renders the existing `MealPlanWidget` byte-for-byte.
- `view: "today-detail"` renders the two-section layout.
- The Zod schema's default for `view` is `"list"`.

### 10.2 Backend Tests

`processor_test.go`:

- `Seed` with `seedKey = nil` preserves the existing "any-row counts" semantics (existing tests, unchanged).
- `Seed` with `seedKey = &"home"` on a fresh household: creates, returns `Created: true`.
- `Seed` with `seedKey = &"home"` after a previous home seed: returns `Created: false` with the existing row.
- `Seed` with `seedKey = &"kiosk"` after a prior `home` seed: creates a second row.
- Concurrent `Seed` calls with the same key: both serialize on the advisory lock; only one creates.
- Concurrent `Seed` calls with **different** keys for the same household: both proceed independently.
- Brownfield backfill: the migration UPDATEs existing rows matching `(name='Home', sort_order=0, user_id IS NULL, seed_key IS NULL)`; rows that don't match are untouched.

`rest_test.go`:

- Body with malformed `key` returns `422` with pointer `/data/attributes/key`.
- Body with no `key` continues to work (legacy clients).

### 10.3 Parity & Integration

`dashboard-renderer-parity.test.tsx` assertion is broadened (§7) to consider both seed layouts.

`seed-layout.test.ts` for `kiosk-seed-layout.ts`:

- Output passes the layout schema.
- Every widget type referenced is in the allowlist.
- No widget violates `x + w ≤ 12`.
- Each invocation produces fresh widget UUIDs.

DashboardRedirect tests are extended to cover:

- Brand-new household: both seed calls fire, list refetches, redirect to Home.
- Brownfield household with Home only: kiosk seed fires, home seed is a no-op, list refetches, redirect to Home.
- Returning household with both already seeded: no seed calls fire (home guard via list, kiosk guard via preference flag).
- Household where user deleted Kiosk: preference flag is set, no kiosk seed call.

## 11. Service Impact Summary

| Service | Change |
|---|---|
| **frontend** | Five new widget definitions + adapters; updated `widget-registry.ts`; new `kiosk-seed-layout.ts`; updated `meal-plan-today` schema and adapter (new `MealPlanTodayDetail`); new `useLocalDateOffset` hook; updated `DashboardRedirect.tsx` two-seed orchestration with preference-flag gate; updated `useSeedDashboard` and `dashboardService.seedDashboard` to pass `key`. |
| **dashboard-service** | `Entity.SeedKey *string`; migration adds column + partial unique index + brownfield backfill; `Processor.Seed` reworked with optional seed key and per-key advisory lock; REST handler reads `attributes.key`; allowlist additions in `shared/go/dashboard/types.go` + parity fixture. |
| **account-service** | New preference key `kiosk_dashboard_seeded` (boolean, per-user-per-household). No schema migration — key/value bag. |
| **All other services** | None. |

## 12. Open Items for Implementation Plan

These are not unresolved questions — they're concrete implementation choices the plan should pin down:

1. **Brownfield backfill heuristic exactness.** Whether to match `name = 'Home'` AND `sort_order = 0` AND `user_id IS NULL`, or relax/tighten. Recommended as listed; rename-edge case is acknowledged trade-off.
2. **`useCalendarEventsForDate` extraction.** Whether to refactor `CalendarWidget` to consume the extracted hook in this task or leave its current internals and only build the extracted hook for `calendar-tomorrow`. Recommended: extract, so today and tomorrow share a single code path. Adds slight scope; pays back in one fewer code path to maintain.
3. **Reminders relative-time formatter.** Confirm whether a shared helper exists; if not, add one in `lib/date-utils.ts`. Inline computation is acceptable if a single call site needs it.
4. **`MealPlanTodayDetail` data shape.** Whether the existing meal-plan query already returns the multi-day window when `horizonDays > 1`; if not, a small server-side parameter or client-side window expansion is needed. Should be determined while reading `MealPlanWidget` during planning.

## 13. Acceptance Criteria Mapping

Every PRD §10 acceptance criterion maps to a concrete code/test artifact in this design:

- Five new widgets in palette + render + read-only + header link → §4 widget files + §10.1 tests.
- `meal-plan-today` `view: "today-detail"` → §8 + §10.1 tests.
- Allowlist additions in Go and TS, parity test passes → §3.4 + §7.
- Brand-new and brownfield households both get Home and Kiosk → §6.4.1 + §10.3.
- Deletion of Kiosk is permanent → §6.4 preference flag + §10.3 test.
- Independent loading/empty/error states → §9.
- Unit tests for schemas and rendering → §10.1.
- Parity test against updated seed layouts → §7 + §10.3.
- No regressions on existing Home dashboard → enforced by Zod default `view: "list"` + the unchanged `seedLayout()` + existing snapshot tests.
