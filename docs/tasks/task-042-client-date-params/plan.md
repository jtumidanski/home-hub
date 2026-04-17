# Task 042: Client-Supplied Date Parameters for "Today" Endpoints

Last Updated: 2026-04-16

---

## Executive Summary

Four backend services (tracker, workout, productivity, calendar) currently compute "today" server-side by resolving the user's timezone from an `X-Timezone` HTTP header with a fallback to an account-service household lookup. The fallback is silently broken — it authenticates via `Authorization` while the platform uses cookie auth — so when `X-Timezone` is missing or invalid the chain drops to UTC, flipping "today" to "tomorrow" at UTC midnight (8 PM US Eastern).

This task removes server-side timezone resolution entirely. Every affected endpoint takes the date as an explicit query parameter (`?date=YYYY-MM-DD` or `?start=`/`?end=` for calendar ranges) computed by the frontend from the household timezone. The three `internal/tz/` packages, productivity's inline `resolveTimezone` helper, and the frontend's `X-Timezone` header injection are all deleted.

Rollout follows a three-PR sequence to avoid any mid-deploy 400s: backend accept-both → frontend switch → backend require + cleanup. See `migration-plan.md` for details.

---

## Current State Analysis

### Backend

- `tracker-service`, `workout-service`, `calendar-service` each own an `internal/tz/` package with `resolver.go` + `accountclient.go` that runs the X-Timezone → household-tz → UTC chain.
- The household-tz fallback uses `r.Header.Get("Authorization")`, but `shared/go/auth/auth.go` only reads the `access_token` cookie — so the outbound request is unauthenticated and always 401s. Fallback is effectively dead code; all successful resolutions rely entirely on the header.
- `productivity-service/internal/summary/resource.go` has an inline `resolveTimezone` that skips the household fallback and goes straight to UTC on any header failure.
- Server-side `time.Now()` leaks into downstream processors: tracker's entry/month gates, workout's planned-item completion check, productivity's task summary counts.
- `time.Now().In(loc)` is converted to `time.Date(Y, M, D, 0, 0, 0, 0, time.UTC)` before DB queries, which works correctly *if* `loc` was right — and silently yields the wrong day when `loc` fell back to UTC.

### Frontend

- `src/lib/api/client.ts:141-146` injects `X-Timezone` from `Intl.DateTimeFormat().resolvedOptions().timeZone` on every request.
- Dashboard widgets (`habits-widget.tsx`, `workout-widget.tsx`) and today pages (`today-view.tsx`, `WorkoutTodayPage.tsx`) call `useTrackerToday()` / `useWorkoutToday()` with no date argument — the backend decides.
- Task summary widgets call `useTaskSummary()` similarly.
- React Query keys for these "today" hooks do NOT include the date — so a cache populated before midnight survives past midnight without invalidation.
- Meal plan and calendar widgets already compute dates client-side and pass them as params (`getLocalTodayStr(tz)`, `getLocalWeekStart(tz)`, `getLocalTodayRange(tz)`). These are the template to follow.

### Observed Bug

At 8 PM US Eastern (00:00 UTC), the dashboard's Habits and Workout widgets, plus the Today views on mobile, flip to tomorrow's content. Root cause: X-Timezone header occasionally fails to arrive on mobile, the fallback silently 401s, and the chain drops to UTC.

---

## Proposed Future State

### Backend

- Every affected endpoint takes the date as a required query parameter. Missing/malformed → 400.
- A shared `shared/go/http/params.go` provides `ParseDateParam(r, name) (time.Time, error)`.
- The date flows through processors as a `time.Time` anchored to midnight UTC (a calendar day, not an instant). Downstream DB queries use it directly.
- `internal/tz/` packages deleted from tracker, workout, calendar services. `resolveTimezone` deleted from productivity. `cmd/main.go` in those services no longer wires `accountBaseURL` for tz lookups.
- Services run identically regardless of the host's local timezone (verified by a test that sets `time.Local` to a non-UTC zone).

### Frontend

- Hooks take the date as an argument and include it in the React Query key:
  - `useTrackerToday(date: string)`
  - `useWorkoutToday(date: string)`
  - `useTaskSummary(date: string)`
  - `useMonthSummary(month, today?)`
- Widgets read `household?.attributes.timezone` (with browser-tz fallback) and compute the date via `getLocalTodayStr(tz)`.
- A new `useLocalDate(tz)` hook re-renders when the local date string changes (polls wall clock every 60s). Dashboard and today pages use it so overnight-open tabs refresh correctly.
- `X-Timezone` injection removed from `src/lib/api/client.ts`.

---

## Implementation Phases

Phases mirror the 3-PR migration plan.

### Phase 1 — PR 1: Backend Accept-Both (Non-Breaking)

Goal: every affected endpoint accepts the new parameter(s); when absent, falls back to current tz-resolution behavior. No 400s yet.

**Task 1.1: Shared `ParseDateParam` helper** [S]
- Create `shared/go/http/params.go` with `ParseDateParam(r *http.Request, name string) (time.Time, error)`.
- Returns `time.Date(Y, M, D, 0, 0, 0, 0, time.UTC)` on success.
- Error format names parameter and raw value.
- Unit tests cover: valid, missing, empty, malformed format, invalid calendar date (e.g. `2026-02-30`), whitespace padding.
- Acceptance: tests pass; helper exported and importable from each service module.

**Task 1.2: tracker-service accept `?date=` on `/trackers/today`** [S]
- `internal/today/resource.go`: if `?date=` present, parse it; else fall back to existing `time.Now().In(loc)` path.
- `internal/today/processor.go`: signature can remain `Today(userID, date time.Time)`; the resource decides which `date` to pass.
- No other handlers changed in this PR.
- Acceptance: existing behavior unchanged when `?date=` absent; when present, the returned `attributes.date` echoes the input and items/entries match that day.

**Task 1.3: workout-service accept `?date=` on `/workouts/today`** [S]
- Same pattern as 1.2. Resource parses `?date=` when present; else uses tz-resolved now.
- Acceptance: same as 1.2.

**Task 1.4: productivity-service accept `?date=` on `/summary/tasks` and `/summary/dashboard`** [M]
- `internal/summary/resource.go`: both handlers check for `?date=` first.
- `internal/summary/processor.go`: `TaskSummary(date time.Time)` — existing `now time.Time` parameter can be reused; resource decides what to pass.
- `internal/task/provider.go`: no changes (still accepts `time.Time`).
- Acceptance: `completedTodayCount` and `overdueCount` use the supplied date when present; otherwise existing behavior.

**Task 1.5: calendar-service — no change in PR 1** [note only]
- Calendar already accepts `?start`/`?end`. Making them required is a PR 3 task. No transitional work needed.

**Task 1.6: Processor tests for transitional behavior** [S]
- For each service touched, add a test verifying the handler reads `?date=` correctly when present and behaves as before when absent.
- Run `./scripts/test-all.sh` — all green.

**Task 1.7: Build & lint** [S]
- `./scripts/build-all.sh` and `./scripts/lint-all.sh` pass.
- Docker builds succeed for all touched services.

**PR 1 merges and deploys. No user-visible change.**

---

### Phase 2 — PR 2: Frontend Switch (Non-Breaking Against PR 1 Backend)

Goal: frontend sends the new parameters on every request; remove `X-Timezone` header. React Query keys include the date.

**Task 2.1: Add `useLocalDate` hook** [S]
- New file `frontend/src/lib/hooks/use-local-date.ts`.
- Signature: `useLocalDate(tz?: string): string` — returns `YYYY-MM-DD`, re-renders when the local date string changes.
- Implementation: compute initial via `getLocalTodayStr(tz)`; `setInterval(60_000)` compares and `setState` on change; cleanup on unmount.
- Unit test: fake timers, advance past midnight, assert the hook re-renders with the new date.
- Acceptance: hook value stable within a day, transitions exactly once across midnight.

**Task 2.2: Update `useTrackerToday` signature** [S]
- File: `frontend/src/lib/hooks/api/use-trackers.ts`
- New signature: `useTrackerToday(date: string)`.
- Pass `?date=${date}` to `trackerService.getToday(tenant, date)`.
- Update `trackerKeys.today` to `(tenant, household, date)`.
- Update `trackerService.getToday` in `frontend/src/services/api/tracker.ts` to append the query param.
- Acceptance: hook compiles; tests updated.

**Task 2.3: Update `useWorkoutToday` signature** [S]
- File: `frontend/src/lib/hooks/api/use-workouts.ts`
- Same treatment as 2.2.
- Update `trackerKeys.today` in `workoutKeys.today` path.
- Update invalidations: existing `qc.invalidateQueries({ queryKey: workoutKeys.today(...) })` calls still match by prefix; no change needed because React Query does partial matching.
- Acceptance: hook compiles; tests updated.

**Task 2.4: Update `useTaskSummary` signature** [S]
- Locate the hook (likely `frontend/src/lib/hooks/api/use-tasks.ts`).
- New signature: `useTaskSummary(date: string)`.
- Pass `?date=` in the query; include in the query key.
- Acceptance: hook compiles; tests updated.

**Task 2.5: Update `useMonthSummary` to accept `today` param** [S]
- File: `frontend/src/lib/hooks/api/use-trackers.ts`
- New signature: `useMonthSummary(month: string, today?: string)`.
- Pass `?today=` when supplied.
- Update `trackerKeys.month` to include `today`.
- Acceptance: existing callers compile; tracker calendar grid updated to pass `today` (Task 2.8).

**Task 2.6: Update `HabitsWidget` and `WorkoutWidget`** [S]
- Files: `src/components/features/trackers/habits-widget.tsx`, `src/components/features/workouts/workout-widget.tsx`
- Read `household` from `useTenant()`; compute date via `useLocalDate(household?.attributes.timezone)`; pass to hook.
- Acceptance: widgets render correctly; no visual regressions.

**Task 2.7: Update `TodayView` and `WorkoutTodayPage`** [S]
- Files: `src/components/features/tracker/today-view.tsx`, `src/pages/WorkoutTodayPage.tsx`
- Same pattern as 2.6.
- Acceptance: today views render correctly.

**Task 2.8: Update tracker `CalendarGrid` to pass `today`** [S]
- File: `src/components/features/tracker/calendar-grid.tsx`
- Already computes `today` via `getLocalTodayStr(timezone)` at line 94. Thread it into `useMonthSummary(month, today)`.
- Acceptance: past/future cell gating is correct in the user's local tz.

**Task 2.9: Update dashboard task summary callers** [S]
- Locate every `useTaskSummary()` call site (dashboard page widgets). Grep: `useTaskSummary\b`.
- Thread a `useLocalDate` value through each.
- Acceptance: all call sites compile and the task counts match the user's local day.

**Task 2.10: Remove `X-Timezone` header injection** [S]
- File: `src/lib/api/client.ts:141-146`
- Delete the try/catch block that sets `X-Timezone`.
- Acceptance: network requests no longer include the header; no test regressions.

**Task 2.11: Update frontend hook tests** [M]
- Every test that stubs `useTrackerToday` / `useWorkoutToday` / `useTaskSummary` needs a date arg.
- Grep test files for these hook names; update fixtures.
- Acceptance: `npm test` passes.

**Task 2.12: Manual verification** [M]
- `./scripts/local-up.sh`.
- Open dashboard at night (or spoof device date/time to 8 PM Eastern).
- Verify habits + workout widgets show today's items, not tomorrow's.
- Let the dashboard sit open across a simulated midnight; verify widgets refresh within 60s.
- Check Network tab: requests have `?date=YYYY-MM-DD`; no `X-Timezone` header.
- Acceptance: all scenarios pass.

**Task 2.13: Build & lint** [S]
- `npm run build`, `./scripts/lint-all.sh` pass.
- Docker frontend build succeeds.

**PR 2 merges and deploys. X-Timezone traffic drops to zero across the fleet.**

---

### Phase 3 — PR 3: Backend Require + Cleanup (Breaking)

**Pre-merge gate:** confirm PR 2 has been in prod at least an hour and no requests to affected endpoints arrive without `?date=` (check nginx/service access logs).

**Task 3.1: tracker-service — require `?date=`, delete `internal/tz/`** [M]
- `internal/today/resource.go`: delete `tz.Resolve`, `tz.WithLocation`, `tz.NewAccountHouseholdLookup`. Use `ParseDateParam`; return 400 on error.
- Delete `accountBaseURL` from `InitializeRoutes` signature; update `cmd/main.go` call site.
- Delete `services/tracker-service/internal/tz/` directory (resolver.go, accountclient.go, resolver_test.go).
- Remove account-service URL env var from tracker-service if unused elsewhere.
- Update `internal/today/processor_test.go` to drop X-Timezone test cases; add a test with `time.Local` set to `America/Los_Angeles` proving tz independence.
- **Ripple tasks (Open Question 1a from PRD):**
  - `internal/entry/processor.go:53,123`: accept `today time.Time` parameter; plumb through the REST handler via `?today=` on `PUT /trackers/:item/entries/:date`. Return 400 if missing.
  - `internal/month/processor.go:71`: same treatment via `?today=` on `GET /trackers/months/:month`.
  - Update processor tests.
- Acceptance: `go test ./...` passes; `go build ./...` passes; `grep -r tz.Resolve` returns nothing in tracker-service.

**Task 3.2: workout-service — require `?date=`, delete `internal/tz/`** [M]
- Same structure as 3.1.
- `internal/today/resource.go` + `processor.go`: switch to `ParseDateParam`. Signature `Today(userID, date time.Time)`.
- Delete `internal/tz/` directory.
- Drop `accountBaseURL` wiring in `cmd/main.go`.
- Ripple: `internal/planneditem/processor.go:251` — thread the client's date through if the helper is on a hot request path, else accept that completion check is UTC-anchored (confirm via code review).
- Acceptance: tests pass; tz package gone.

**Task 3.3: productivity-service — require `?date=`, delete `resolveTimezone`** [M]
- `internal/summary/resource.go`: delete `resolveTimezone` helper. `taskSummaryHandler` and `dashboardSummaryHandler` use `ParseDateParam`.
- `internal/summary/processor.go`: `TaskSummary(date time.Time)`.
- `internal/task/provider.go`:
  - `countOverdue(db, date time.Time)` — compare `due_on < date` (date format string).
  - `countCompletedToday(db, date time.Time)` — `startOfDay := date` (already UTC-anchored midnight).
- Leave `reminderSummaryHandler` unchanged; it has no date parameter.
- Acceptance: tests pass; no `resolveTimezone` references remain.

**Task 3.4: calendar-service — require `?start` and `?end`, delete `internal/tz/`** [M]
- `internal/event/resource.go`: parse `?start` and `?end` as `time.Time` via `time.Parse(time.RFC3339, v)`. 400 on missing or malformed. 400 if `end <= start`.
- Delete the default-range computation (lines 49-74 in current file).
- Delete `tz.Resolve` call. Delete `internal/tz/` directory.
- Drop `accountBaseURL` wiring in `cmd/main.go`.
- Acceptance: tests pass; tz package gone.

**Task 3.5: Update service `rest.md` files** [S]
- `services/tracker-service/docs/rest.md`: document `?date=` (and `?today=` on month endpoint + entry PUT). Remove `X-Timezone` references.
- `services/workout-service/docs/rest.md`: document `?date=`. Remove `X-Timezone` references.
- `services/productivity-service/docs/rest.md`: document `?date=` on summary endpoints. Remove `X-Timezone` references.
- `services/calendar-service/docs/rest.md`: document required `?start` + `?end`. Remove any stale tz-resolution notes.
- Acceptance: docs match current behavior; grep for `X-Timezone` returns nothing in `services/*/docs/`.

**Task 3.6: Bruno collection updates** [S]
- Grep `bruno/` for the affected endpoints.
- Add `?date=` (etc.) to every request that hits them.
- Acceptance: Bruno runs against local compose return 2xx; any that fail 400 have been updated.

**Task 3.7: Build, lint, test** [S]
- `./scripts/build-all.sh`, `./scripts/lint-all.sh`, `./scripts/test-all.sh` all green.
- Docker builds succeed for tracker, workout, productivity, calendar.

**Task 3.8: Post-deploy verification** [S]
- After PR 3 deploys, hit each endpoint without the param; confirm 400.
- Hit each endpoint with a valid param; confirm 200 and expected shape.
- Monitor logs for any unexpected 400 spike (would indicate stale clients).

---

## Risk Assessment and Mitigation

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Stale browser cache hits new backend with old bundle (PR 3 window) | Medium | Medium | Three-PR rollout. PR 3 waits until PR 2 traffic is 100% and `X-Timezone` has disappeared from logs. |
| Frontend computes wrong date when `household.attributes.timezone` is missing | Low | Low | Fallback to `Intl.DateTimeFormat().resolvedOptions().timeZone` (browser). User is almost always in the household's tz. |
| `useLocalDate` polling drifts if tab is throttled | Low | Low | 60s interval tolerates throttling; `refetchOnWindowFocus` already on by default forces a fetch when the user returns. |
| Breaking a non-affected service by accidentally modifying `shared/go/http` | Low | Medium | `ParseDateParam` is additive; unit-tested in isolation. |
| Dockerfile misses shared/go/http module | Low | Low | `go.work` picks it up automatically; CI catches any miss. |
| Ripple endpoints (tracker entry/month) break existing frontend flows | Medium | Medium | Part of PR 2 (frontend) + PR 3 (backend require); Phase 2 tasks explicitly thread `today` through `CalendarGrid`. |
| Test suite regresses for time-sensitive paths | Medium | Low | New tz-independent test (`time.Local = ...`) in processor tests guards against it. |

---

## Success Metrics

- Zero `tz.Resolve` calls in the codebase after PR 3.
- Zero occurrences of the `X-Timezone` string in `frontend/src/lib/api/client.ts` or any `services/*/internal/` directory.
- Zero user reports of "flipped to tomorrow at 8 PM" after PR 2 deploy.
- Test coverage: every processor that consumes the client date has at least one test that runs under a non-UTC `time.Local`.
- `./scripts/test-all.sh` passes at every merge point.
- Dashboard widgets refresh within 60s of local midnight without a manual reload.

---

## Required Resources and Dependencies

- No new third-party libraries.
- No schema changes; no migrations.
- Depends on task-038's `frontend/src/lib/date-utils.ts` being present (it is; verified in earlier exploration).
- Depends on `shared/go/http` module being loadable from each service's `go.mod` (uses `go.work`; no action needed).

---

## Timeline Estimates

| Phase | Effort | Depends On |
|---|---|---|
| Phase 1 (PR 1) | M | None |
| Phase 2 (PR 2) | L | PR 1 deployed to prod |
| Phase 3 (PR 3) | M | PR 2 deployed ≥ 1 hour |

Tasks within a phase are mostly independent and can be tackled in any order — except the shared helper (1.1) must precede 1.2–1.4, and `useLocalDate` (2.1) must precede 2.6–2.9.

Parallelism: within PR 1, the four service updates (1.2, 1.3, 1.4) are independent. Within PR 3, the four service cleanups (3.1, 3.2, 3.3, 3.4) are independent.
