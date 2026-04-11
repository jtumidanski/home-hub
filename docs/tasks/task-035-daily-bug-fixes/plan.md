# Task 035 — Daily Use Bug Fixes — Implementation Plan

Last Updated: 2026-04-10
Status: Ready for implementation
Source PRD: `prd.md`

## Executive Summary

Four small but workflow-breaking bugs surfaced during a single dogfooding session. They span the frontend (3) and the Go backend (1, touching workout-service + tracker-service). All fixes are surgical; the largest piece of work is plumbing a per-request timezone resolution into the `/today` processors. Bundling the four into a single task is intentional — they share a deploy and a verification pass.

The four bugs:

1. **Bug 1 — Strength weight input missing.** Weight + Unit fields don't appear (or aren't reachable) for non-bodyweight strength exercises on `WorkoutTodayPage` and `WorkoutWeekPage`.
2. **Bug 2 — `/today` is UTC.** Workouts/trackers compute "today" using `time.Now().UTC()`, so late-evening sessions in non-UTC zones are attributed to the wrong day.
3. **Bug 3 — Tracker range slider sticky state.** `RangeEditor` retains the previous day's slider value when navigating between days.
4. **Bug 4 — UUIDs in exercise dialog Select triggers.** A regression of the task-032 §4.2 fix; one or more `Select`s still display raw UUIDs.

## Current State Analysis

### Bug 1 — Weight input
- `frontend/src/pages/WorkoutTodayPage.tsx:167–196` renders the strength row using `grid-cols-4` when `!isBw`. The `!isBw` branch does include Weight + Unit inputs. The likely culprits per PRD §4.1 are: (a) stale `weightType` data, (b) the 4-column grid clipping/squeezing on mobile, or (c) projection bug. Reproduction is required before patching.
- `frontend/src/pages/WorkoutWeekPage.tsx:384–401` uses `grid-cols-2` and may have the gating identical issue.

### Bug 2 — Timezone
- `services/workout-service/internal/today/processor.go:45` and `services/tracker-service/internal/today/processor.go:40`, `resource.go:26` use `time.Now().UTC().Truncate(24 * time.Hour)`.
- Grep across `services/**/today/*.go` confirms only **workout-service** and **tracker-service** expose `/today`. No productivity-service or meals `/today` endpoint exists today.
- `services/account-service/internal/household` already carries a `Timezone` field; no schema work.
- Frontend axios instance lives in `frontend/src/services/api/`; an interceptor exists for tenant headers.

### Bug 3 — Range slider
- `frontend/src/components/features/tracker/calendar-grid.tsx:356` (`RangeEditor`) uses `useEffect([initial])` that early-returns when `initial === undefined`, so previous-day state persists.
- `frontend/src/components/features/tracker/today-view.tsx:165` (`RangeInput`) shows the same pattern at line 169.

### Bug 4 — Select UUIDs
- `frontend/src/pages/WorkoutExercisesPage.tsx:189–191, 206–208` already has the `find().attributes.name` fix. Bug must live elsewhere — most likely the WeekPage filter selects at `WorkoutWeekPage.tsx:523, 536`, which use `<SelectValue placeholder="Theme" />` with no children fallback. Reproduction must pin it before patching.

## Proposed Future State

- Strength exercises with `weightType !== "bodyweight"` show editable Weight + Unit inputs everywhere they're loggable, on all viewports ≥360px.
- Every `/today` endpoint computes "now" in the resolved local zone via the hybrid header→household→UTC strategy. The frontend axios layer attaches `X-Timezone` automatically.
- Tracker range slider always reflects the **active** cell — never the previously viewed one.
- Every `Select` whose value is a UUID renders a name in its trigger using the same `find().attributes.name` pattern, with explicit children fallback.

## Implementation Phases

The four bugs are independent and can be tackled in parallel by separate sub-passes, but they share a single PR/branch. Recommended order optimizes for risk: backend plumbing first (highest blast radius), then the three frontend fixes.

### Phase 1 — Bug 2: Timezone-aware `/today` (backend + frontend interceptor)

The largest piece. Touches two services and the frontend axios layer.

#### 1.1 Backend — TZ resolution helper *(M)*
- Add a small helper inside each service (Option A from PRD §4.2) — `internal/tz/resolver.go` — exposing `tz.Resolve(ctx, headers, householdLookup) *time.Location`.
- Resolution order:
  1. `X-Timezone` header → `time.LoadLocation` → use if valid.
  2. Household timezone via existing inter-service account-service client → `time.LoadLocation` → use if valid.
  3. UTC fallback + warn-level structured log.
- Cache the resolved `*time.Location` on the request `context.Context` via a typed key so multiple call sites in the same request don't re-fetch.
- **Acceptance:** unit tests cover all three branches plus an invalid-header case.
- **Dependencies:** none.

#### 1.2 workout-service `/today` wiring *(S)*
- `services/workout-service/internal/today/resource.go`: read `X-Timezone` from the HTTP request, call the resolver, attach `*time.Location` to `ctx`.
- `services/workout-service/internal/today/processor.go:45`: replace `time.Now().UTC()` with `time.Now().In(loc)`. Recompute `weekStart` and `dayOfWeek` in `loc`.
- Update existing processor tests to pass an explicit location and add a new test: "21:30 America/New_York on a Monday is still Monday's day-of-week even though UTC is Tuesday 02:30."
- **Dependencies:** 1.1.

#### 1.3 tracker-service `/today` wiring *(S)*
- Same change as 1.2 for `services/tracker-service/internal/today/{processor.go,resource.go}` and `processor_test.go`.
- **Dependencies:** 1.1.

#### 1.4 Frontend `X-Timezone` interceptor *(S)*
- Add a request interceptor in `frontend/src/services/api/` (next to the existing tenant interceptor) that sets `X-Timezone: Intl.DateTimeFormat().resolvedOptions().timeZone` on every outgoing request.
- Manual verification: open DevTools → Network → confirm header on `/workouts/today` and `/trackers/today` calls.
- **Dependencies:** none (can ship before backend, header is silently ignored by services that don't read it).

#### 1.5 Other `/today` endpoints sweep *(S)*
- Grep `services/**/today/*.go` (already done — only workout + tracker exist today). Re-grep at implementation time to confirm and document the result.
- If a new `/today` endpoint shows up (e.g., productivity-service), apply the same fix; otherwise note "no other endpoints found" in the task notes.

### Phase 2 — Bug 1: Weight input on strength exercises

#### 2.1 Reproduce *(S)*
- Open a free-weight strength exercise on a 375px-wide viewport on both `WorkoutTodayPage` and `WorkoutWeekPage`. Document whether Weight + Unit are missing entirely vs. visually clipped.
- Inspect the user's exercise data via DB or API call to confirm `weightType` value.
- **Acceptance:** root cause pinned and noted in `tasks.md`.

#### 2.2 Apply fix per root cause *(S)*
- **If layout-clipping** (most likely): change `grid-cols-4` → `grid-cols-2 sm:grid-cols-4` on `WorkoutTodayPage.tsx:168` so mobile reflows into a 2×2 grid. Apply the equivalent reflow on `WorkoutWeekPage.tsx:384–401` if it has the same issue.
- **If projection bug**: fix `services/workout-service/internal/weekview/projection.go:58` to populate `WeightType` correctly and rebuild.
- **If stale data**: do NOT lift `WeightTypeImmutable`. Surface the existing `weightType` on the exercises page so the user can verify, and file a follow-up task for a one-shot data correction.
- **Acceptance:** PRD §10 acceptance criteria for Bug 1 pass on a mobile viewport.

### Phase 3 — Bug 3: Range slider reset

#### 3.1 `RangeEditor` reset *(S)*
- `frontend/src/components/features/tracker/calendar-grid.tsx:359–364`: change the `useEffect` so that when `initial === undefined`, `local` resets to `Math.round((min + max) / 2)` and `touched` resets to `false`.
- Key the effect dependency on `(itemId, date, initial)` so the popover reinitializes on every cell change.

#### 3.2 `RangeInput` defensive fix *(S)*
- Apply the equivalent change to `frontend/src/components/features/tracker/today-view.tsx:165–169`.

#### 3.3 Verify *(S)*
- Open the month grid, edit a value on day N, navigate to day N+1 (no entry), open popover → expect midpoint + "Not set."

### Phase 4 — Bug 4: UUID-in-trigger regression

#### 4.1 Reproduce + locate *(S)*
- Click through every dialog/filter that contains a Theme or Region `Select`. Note which one renders a UUID.
- Suspected: `WorkoutWeekPage.tsx:523, 536` (Add Exercise filter selects with no children fallback).

#### 4.2 Patch all UUID-valued Selects *(S)*
- Apply the `find().attributes.name` children pattern inside `<SelectValue>` for every offending site.
- Search for any other `<SelectValue placeholder="Theme" />` / `placeholder="Region"` patterns and pre-emptively fix them.

#### 4.3 Verify *(S)*
- Select a value in each dialog/filter and confirm the trigger displays the name within one render.

### Phase 5 — Verification & Build

- Run `go test ./...` for `workout-service` and `tracker-service`.
- Run frontend type-check + lint + test (`pnpm test` or equivalent).
- Run `scripts/local-up.sh` and smoke-test all four bug repros end-to-end on a mobile viewport.
- Confirm the `X-Timezone` header is present in DevTools network panel.

## Risk Assessment & Mitigation

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| TZ helper introduces a synchronous account-service call on every `/today` request, adding latency | M | M | Cache the resolved `*time.Location` on the request context; use the existing inter-service client which already has connection pooling. If the household lookup turns out to be hot, escalate to a per-tenant in-memory cache (existing pattern in tenant middleware). |
| Frontend `Intl.DateTimeFormat().resolvedOptions().timeZone` returns an unexpected/missing value in some browsers | L | L | Backend silently falls back to household → UTC. Bad header never 5xxes. |
| Bug 1 root cause is stale data, not layout, requiring a follow-up | M | L | PRD already authorizes a follow-up task; do not expand scope. |
| Existing `today` processor tests break when the function signature gains a `*time.Location` | H | L | Update tests in the same commit; this is mechanical. |
| The `RangeEditor` reset breaks an in-progress edit if `initial` arrives async after the popover opens | L | M | Key the effect on `(itemId, date)` so the same-cell case still respects late-arriving `initial`. |
| Mobile reflow for the weight grid changes existing layout for bodyweight rows | L | L | Only change the `!isBw` branch's column count. |

## Success Metrics

- All eight PRD §10 acceptance checkboxes pass during manual verification.
- `go test ./...` green for both affected services.
- Frontend type-check + lint + tests green.
- Zero regressions in the `WorkoutTodayPage`, `WorkoutWeekPage`, and tracker month-grid flows.

## Required Resources & Dependencies

- Local Docker stack via `scripts/local-up.sh`.
- A user account with: at least one free-weight strength exercise, a tracking item with a range scale, and a household whose timezone is something other than UTC (or use a `X-Timezone` header in manual testing).
- No external dependencies. No schema changes. No data migrations.

## Timeline Estimate

- Phase 1 (TZ): ~½ day — backend helper + two services + frontend interceptor + tests.
- Phase 2 (Weight): ~1 hour reproduce + 1 hour fix.
- Phase 3 (Slider): ~½ hour.
- Phase 4 (UUID): ~½ hour reproduce + ½ hour fix.
- Phase 5 (Verify + build): ~1 hour.

**Total:** ~1 day of focused work, single PR.
