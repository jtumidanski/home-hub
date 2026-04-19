# Task 043: Workout Weekly Review

Last Updated: 2026-04-19

---

## Executive Summary

The Workout section ships a Summary tab today: three header counts (Planned / Performed / Skipped) plus per-theme and per-region totals. The backend summary projection already computes a full per-day breakdown (what was planned, what was actually done, per-item status) and the UI throws it away. There is also no prev/next navigation on the Summary page, so an empty week dead-ends on a generic error.

This task converts that tab into a proper **Weekly Review**:

1. Surface the per-day breakdown with planned-vs-actual rendered side-by-side (including per-set strength detail).
2. Add prev / next week navigation mirroring `WorkoutWeekPage`.
3. Add "jump to previous/next populated week" controls so users can skip past empty weeks.
4. Replace the empty-week error with a friendly card that keeps navigation functional.
5. Rename `Summary` → `Review` in routes, tab label, and page component (old URLs redirect).

Scope is presentation plus one additive navigation helper endpoint and one projection extension. No schema changes, no new domain logic, no analytics. Cross-week charts remain deferred (task-027 §2 non-goals).

Rollout: a single PR is appropriate — the new summary fields are additive, the `/weeks/nearest` endpoint is new, and `/summary` redirects keep existing bookmarks working.

---

## Current State Analysis

### Backend

- `services/workout-service/internal/summary/processor.go` builds the `RestModel` with `weekStartDate`, `restDayFlags`, totals, `byDay`, `byTheme`, `byRegion`. Per-set performances are collapsed to `sets = count`, `reps = max(reps)`, `weight = max(weight)` (PRD cites lines 487–505).
- `services/workout-service/internal/week/provider.go:26` already implements `GetMostRecentPriorWithItems(db, userID, before)` — a `planned_items INNER JOIN weeks` pattern that finds the nearest earlier populated week. No symmetric "next" helper exists.
- `services/workout-service/internal/weekview/` owns week-scoped REST routes that cross package boundaries (kept separate from `week` to avoid import cycles, per `internal/week/resource.go:10`).
- No endpoint exposes "nearest populated week."

### Frontend

- `frontend/src/pages/WorkoutSummaryPage.tsx` renders totals + per-theme + per-region. It does not render `byDay` or any per-item detail.
- `frontend/src/components/features/workout/workout-shell.tsx` wires the `Summary` sidebar tab to `/app/workouts/summary`.
- `frontend/src/lib/hooks/api/use-workouts.ts` exports `useWorkoutWeekSummary`; the response type in `frontend/src/types/models/workout.ts` does not yet expose per-day detail, per-set rows, or populated-week pointers.
- `WorkoutWeekPage` already implements a prev/next header with `« Prev` / `Next »` buttons — it is the reference pattern.
- Empty-week state currently falls through to a generic error; there is no navigation affordance.

### Routes

- `/app/workouts/summary` and `/app/workouts/summary/:weekStart` are wired to `WorkoutSummaryPage`.
- `/app/workouts/week` and `/app/workouts/week/:weekStart` host `WorkoutWeekPage`.

---

## Proposed Future State

### Backend

- `GET /api/v1/workouts/weeks/{weekStart}/summary` response includes two new top-level attributes: `previousPopulatedWeek` and `nextPopulatedWeek` (each `string | null`).
- Per-item `actualSummary` for per-set performances additionally includes a `setRows` array (renamed from `sets` in the PRD example to avoid collision with the existing scalar `sets` count).
- New endpoint `GET /api/v1/workouts/weeks/nearest?reference=YYYY-MM-DD&direction=prev|next` returns `{weekStartDate}` or `404`. Authenticated; filtered by user.
- New provider helper `week.GetSoonestNextWithItems(db, userID, after)` mirrors the existing prior-direction helper.

### Frontend

- `WorkoutSummaryPage.tsx` renamed to `WorkoutReviewPage.tsx`; sidebar tab reads `Review`; canonical routes are `/app/workouts/review[/:weekStart]`; `/summary[/:weekStart]` paths redirect.
- Page renders four-stat header (Planned / Performed / Pending / Skipped), prev / next / jump-to-populated navigation, and a seven-column per-day grid (one column on mobile).
- Per-item cards show status badge + planned-vs-actual body, kind-aware (strength summary, strength per-set, isometric, cardio), with a target-met `✓` when actual ≥ planned on the primary metric.
- Empty-week state renders a card and calls `/weeks/nearest` in both directions to hydrate jump buttons.

---

## Implementation Phases

A single PR. Phases correspond to work streams that can be built in parallel once the backend changes land on a shared branch.

### Phase 1 — Backend: Summary Projection Extensions

**Task 1.1: `week.GetSoonestNextWithItems` helper** [S]
- Add to `services/workout-service/internal/week/provider.go` mirroring `GetMostRecentPriorWithItems`.
- `planned_items INNER JOIN weeks WHERE user_id = ? AND week_start_date > ?` ordered ascending, LIMIT 1.
- Unit tests in `provider_test.go` (or adjacent) covering: hit, miss, cross-user isolation, boundary (strictly greater than `after`).
- Acceptance: helper returns `(time.Time, error)` or sentinel equivalent matching the existing helper; tests pass.

**Task 1.2: Populate `previousPopulatedWeek` / `nextPopulatedWeek` in summary** [S]
- Extend `summary.Processor.Build` in `services/workout-service/internal/summary/processor.go` to call both helpers and thread results into `RestModel`.
- Emit as `string | null` attributes. Format `YYYY-MM-DD`.
- Processor test: fixtures with prior populated week only, next only, both, neither.
- Acceptance: summary `200` response includes both fields; values correct across four scenarios.

**Task 1.3: Include per-set rows in `actualSummary` when `mode == per_set`** [M]
- Extend `buildActualSummary` in `services/workout-service/internal/summary/processor.go` to query `workout.performance_sets` for per-set performances and append a `setRows` array to the actual summary.
- Preserve existing collapsed scalar fields (`sets`, `reps`, `weight`, `weightUnit`).
- Shape per row: `{ setNumber: int, reps: int, weight: number }`.
- Omit the array (do not emit `null`) when mode is not `per_set` or the item is non-strength.
- Query batching: fetch all per-set rows for the week in a single query keyed by performance IDs; do not N+1.
- Projection test asserting presence only in per-set case and correct per-row values.
- Acceptance: summary response for a per-set performance contains `setRows`; summary-mode and non-strength items do not.

### Phase 2 — Backend: `/weeks/nearest` Endpoint

**Task 2.1: Handler + route wiring** [M]
- Add handler in `services/workout-service/internal/weekview/resource.go` (or a new `resource_nearest.go` in the same package): parses `reference` and `direction`, normalizes `reference` to Monday-of-ISO-week, dispatches to the appropriate provider helper, returns `{weekStartDate}` or `404`.
- Wire route in `weekview` REST registration (whichever file registers routes today — `rest.go`).
- JSON:API envelope: `data.type = "workoutWeekPointer"`, `data.id = weekStartDate`, `attributes.weekStartDate`.
- Acceptance: `GET /api/v1/workouts/weeks/nearest?reference=2026-04-13&direction=prev` returns `200` with payload or `404`.

**Task 2.2: Validation and errors** [S]
- `400` on missing/invalid `reference` (bad date, bad format) or missing/invalid `direction` (anything other than `prev` / `next`).
- `404` on no populated week in the requested direction for the current user.
- Reuses shared HTTP helpers used elsewhere in this service for error bodies.
- Acceptance: unit/REST tests cover each error path.

**Task 2.3: Cross-user isolation test** [S]
- Integration test: seed two users, each with populated weeks, confirm user A's request never returns user B's weeks.
- Acceptance: test passes.

### Phase 3 — Frontend: Types, Hooks, and Routing

**Task 3.1: Extend `WeekSummary` type** [S]
- Update `frontend/src/types/models/workout.ts` to add `previousPopulatedWeek: string | null`, `nextPopulatedWeek: string | null`.
- Add `setRows?: Array<{ setNumber: number; reps: number; weight: number }>` to the per-item `actualSummary` shape.
- Acceptance: `npm run build` succeeds; no untyped access at call sites.

**Task 3.2: `useWorkoutNearestPopulatedWeek` hook** [S]
- New hook in `frontend/src/lib/hooks/api/use-workouts.ts`: `useWorkoutNearestPopulatedWeek(reference: string, direction: "prev" | "next", enabled?: boolean)`.
- Calls `/weeks/nearest`, returns `{ weekStartDate } | null`.
- React Query key includes `reference` and `direction`.
- Acceptance: hook compiles; usable from the empty-week flow.

**Task 3.3: Rename page + route changes** [S]
- Rename `frontend/src/pages/WorkoutSummaryPage.tsx` → `WorkoutReviewPage.tsx` (git-mv friendly; update all imports).
- In `frontend/src/App.tsx`, add routes `/app/workouts/review` and `/app/workouts/review/:weekStart` pointing at the renamed page.
- Convert `/app/workouts/summary[/:weekStart]` to `<Navigate replace>` redirects to the `review` equivalents, preserving `:weekStart`.
- Update `frontend/src/components/features/workout/workout-shell.tsx`: tab label `Summary` → `Review`; link target `/app/workouts/review`.
- Acceptance: navigating to `/app/workouts/summary` lands on `/app/workouts/review`; deep bookmarks preserve the week param.

### Phase 4 — Frontend: Review Page UI

**Task 4.1: Four-stat totals card** [S]
- Add `Pending = totalPlannedItems − totalPerformedItems − totalSkippedItems` computed client-side.
- Render four stats in the existing totals card: `Planned`, `Performed`, `Pending`, `Skipped` (order per PRD §4.3 and ux-flow.md).
- Acceptance: totals card renders four stats.

**Task 4.2: Navigation header** [M]
- Top row mirrors `WorkoutWeekPage`: `« Prev` / `Week of YYYY-MM-DD` / `Next »`. Prev/Next compute `addDays(weekStart, ±7)` and `navigate(path, { replace: false })`.
- Second row: `↞ Previous populated (YYYY-MM-DD)` / `Next populated ↠ (YYYY-MM-DD)`.
- On populated weeks: jump buttons read `previousPopulatedWeek` / `nextPopulatedWeek` from the summary response.
- On empty weeks (see Task 4.5): two parallel `useWorkoutNearestPopulatedWeek` calls hydrate jump targets.
- Disable buttons when no target exists; disable when target equals current week.
- Acceptance: all four buttons function; URL pushes history entries; browser back returns to prior week.

**Task 4.3: Per-day grid** [M]
- Under the totals card, a `Per Day` section renders seven day blocks Mon → Sun from the `byDay` array.
- Each block: day name, `Rest day` pill when `isRestDay`, item count (`3 exercises`) or `Nothing scheduled` for empty non-rest days.
- Layout: `grid-cols-1 md:grid-cols-7` matching `WorkoutWeekPage`'s breakpoint.
- Items rendered in backend-provided order (`position` ascending).
- Acceptance: all seven days present, including rest-day pill and empty-day placeholder.

**Task 4.4: Per-item card** [L]
- Component renders exercise name, status badge (`Done` / `Partial` / `Skipped` / `Pending`), planned line, actual line.
- Status badge: both color and text label; skipped name gets strikethrough; pending name gets italic + muted.
- Planned-vs-actual body dispatches on `kind`:
  - Strength summary: `Planned: 3×10 @ 135 lb` / `Actual: 3×10 @ 140 lb`.
  - Strength per-set (when `setRows` present): `Actual: set 1: 10 @ 135 · set 2: 8 @ 145 · …`; wraps when ≥ 6 sets.
  - Isometric: `Planned: 3×60s` / `Actual: 3×55s`.
  - Cardio: `Planned: 30:00 · 3.0 mi` / `Actual: 28:45 · 3.1 mi` (omit distance when absent).
- `✓` rendered on the actual line when the target-met predicate holds AND both sides are non-null:
  - Strength: `actualWeight * actualReps * actualSets ≥ plannedWeight * plannedReps * plannedSets`.
  - Isometric: `actualSets * actualDuration ≥ plannedSets * plannedDuration`.
  - Cardio: `actualDistance ≥ plannedDistance` when both have distance, else `actualDuration ≥ plannedDuration`.
- `✓` carries `aria-label="Target met"`.
- `Skipped` → planned muted, actual reads `Actual: Skipped`.
- `Pending` → actual reads `Actual: —`.
- Acceptance: all four kinds × four statuses render correctly in unit tests.

**Task 4.5: Empty-week card** [S]
- When `useWorkoutWeekSummary` returns 404, render a card with the week date in the title and copy: `No workouts logged for this week.`
- Keep the navigation header rendered above the card and functional.
- Fetch `nearest?prev` and `nearest?next` in parallel to hydrate jump buttons.
- Acceptance: empty-week renders no error; jump buttons reflect actual availability; prev/next week buttons still navigate.

**Task 4.6: Accessibility pass** [S]
- `aria-label` on `✓`. Status badges carry both color + text. Strikethrough skipped items paired with `Skipped` badge (covered by badge logic). Buttons have explicit `aria-label` or visible text.
- Day sections use `<section>` with an `<h2>` containing the day name.
- Acceptance: keyboard-only traversal reaches every button; axe-style smoke check passes on the rendered page.

### Phase 5 — Tests, Build, Deploy

**Task 5.1: Backend tests green** [S]
- Unit tests for `GetSoonestNextWithItems`, processor changes, and projection per-set array.
- REST tests for `/weeks/nearest` (200, 404, 400, cross-user).
- `./scripts/test-all.sh` for `workout-service` passes.

**Task 5.2: Frontend tests green** [M]
- Component tests exercise the four kinds × four statuses.
- Hook tests for `useWorkoutNearestPopulatedWeek`.
- Navigation tests (Task 4.2) for prev/next URL push + browser-back round trip.

**Task 5.3: Full build / lint / Docker** [S]
- `./scripts/lint-all.sh`, `./scripts/build-all.sh`, `npm run build` (frontend), Docker builds via `./scripts/local-up.sh`.
- Acceptance: all green.

**Task 5.4: Manual QA against local stack** [M]
- `./scripts/local-up.sh`; exercise:
  - Populated week: header totals + per-day grid + mixed statuses.
  - Strength per-set performance shows each set.
  - Empty week: card + prev/next/jump behavior.
  - `/summary` redirect preserves week param.
  - Sidebar tab reads `Review`.

---

## Risk Assessment and Mitigation

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| JSON key collision between scalar `sets` (count) and per-set `sets` (array) | High | Response fails to serialize correctly or overwrites scalar | Rename per-set array to `setRows` (api-contracts.md §1 naming note). Reflect in frontend type. |
| N+1 query when fetching per-set rows inside the summary projection | Medium | Summary p95 blows past task-027 100ms SLO | Batch-fetch per-set rows for all performances in the week in one query keyed by performance ID; map in-memory. |
| Target-met math on missing planned values | Medium | Crashes or misleading `✓` | Gate the `✓` predicate behind `planned != null && actual != null` on all primary metrics. Unit-test each kind. |
| Sidebar bookmark breakage after rename | Medium | User hits a dead URL | Keep `/summary[/:weekStart]` alive as `<Navigate replace>` redirects indefinitely. |
| Browser back behavior breaks with `navigate(..., { replace: true })` | Medium | Users can't back out of a navigation click | Use default push semantics; verify in manual QA (Task 5.4) and an automated nav test. |
| Jump-to-populated queries scan unbounded history | Low | Slow query if a user has years of data | Existing `(user_id, week_start_date)` composite index + `ORDER BY week_start_date DESC/ASC LIMIT 1` keeps this a single-row index lookup. |
| Cross-user data leak via `/weeks/nearest` | Low | Severe | Handler filters on `user_id` derived from JWT, same as every other workout endpoint; integration test in Task 2.3. |
| Per-set performance mode inference drift (mode set on performance, rows in `performance_sets`) | Low | Missing `setRows` for some users | Gate strictly on `performance.mode == "per_set"`; projection test covers the flag. |

---

## Success Metrics

- **Functional**: every item in §10 Acceptance Criteria of the PRD passes, both in automated tests and in manual QA on `local-up.sh`.
- **Performance**: summary endpoint p95 ≤ 100ms (unchanged SLO from task-027); `/weeks/nearest` p95 ≤ 50ms server-side under typical load (single-row indexed lookup).
- **Accessibility**: keyboard users can reach every navigation button; `✓` has an aria label; status badges carry text in addition to color.
- **Backwards compat**: `/app/workouts/summary[/:weekStart]` bookmarks resolve to the renamed page.
- **No regressions**: `WorkoutWeekPage`, workout today view, and all other workout surfaces behave as before.

---

## Required Resources and Dependencies

- **Services touched**: `workout-service` only.
- **Frontend touched**: `frontend` (page rename, router, shell, hook, types).
- **No infra changes**: nginx/k3s ingress already covers `/api/v1/workouts/*`.
- **No schema changes**: all data comes from existing `workout.weeks`, `workout.planned_items`, `workout.performances`, `workout.performance_sets`.
- **No auth / tenant changes**: existing middleware is reused.
- **No cross-service calls**: all data is local to `workout-service`.

---

## Timeline Estimates

Rough effort (single engineer, end-to-end):

| Phase | Effort |
|---|---|
| Phase 1 — Backend projection + pointers | 1.0–1.5 days |
| Phase 2 — `/weeks/nearest` endpoint | 0.5 day |
| Phase 3 — Frontend types / hook / routing | 0.5 day |
| Phase 4 — Review page UI | 1.5–2.0 days |
| Phase 5 — Tests, build, QA | 0.5–1.0 day |
| **Total** | **4.0–5.5 days** |

Phases 1+2 can merge to a short-lived branch first; phases 3+4 pick up in parallel once the backend shape is stable.
