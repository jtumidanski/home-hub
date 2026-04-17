
# Client-Supplied Date Parameters for "Today" Endpoints — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-04-16
---

## 1. Overview

Multiple backend endpoints currently compute "today" server-side by resolving the user's timezone from an `X-Timezone` HTTP header, with a fallback to looking up the household's timezone from account-service. This chain has a silent failure mode: the account-service lookup is authenticated via the `Authorization` header, but the rest of the platform authenticates via an `access_token` cookie. When `X-Timezone` is missing or invalid (observed intermittently on mobile browsers), the fallback lookup always returns 401 and `tz.Resolve()` drops to UTC. For users in UTC-offset zones this causes "today" to flip to "tomorrow" at UTC midnight — e.g., 8 PM Eastern on April 16 renders April 17 content.

Task-035 introduced the `X-Timezone` header mechanism and task-038 made the frontend consistent with it. Both tasks preserved server-side "today" resolution as the source of truth. This task reverses that choice: the client owns the date. Every endpoint that needs a calendar day takes `?date=YYYY-MM-DD` (or `?start`/`?end` for ranges) as a required query parameter computed by the frontend from the household timezone. The server never resolves a timezone from a request again.

This eliminates an entire class of silent-UTC-fallback bugs, simplifies backend code (three `internal/tz/` packages plus one inline helper deleted), and makes the date behavior of each endpoint explicit and testable.

## 2. Goals

Primary goals:
- Make "today" an explicit, client-supplied parameter on every affected endpoint.
- Remove all server-side timezone resolution from HTTP request paths.
- Ensure React Query caches for date-scoped data invalidate correctly when the local day changes while the app is open.
- Keep service docs (`rest.md`) in lockstep with the new endpoint contracts.

Non-goals:
- Per-user timezone preferences.
- Changes to the household-timezone field (retained for internal/cron use).
- Changes to reminder SQL `CURRENT_TIMESTAMP` filters — these are timestamp comparisons and timezone-invariant.
- Any change to auth-service or account-service.
- New features or UI.

## 3. User Stories

- As a mobile user at 8 PM local time, I want the dashboard's habits and workout widgets to show today's items so I can still log them before bed.
- As a user scrolling past midnight with the app open, I want today's data to refresh automatically instead of showing yesterday.
- As a backend engineer reviewing a 5xx, I want a single obvious reason the date filter is wrong (client sent the wrong `?date=`) instead of a three-layer timezone resolution chain.

## 4. Functional Requirements

### 4.1 Tracker Service — `GET /trackers/today`

**New contract:** `GET /trackers/today?date=YYYY-MM-DD`
- `date` is required. Missing, empty, or malformed values return `400 Bad Request` with a JSON:API error body.
- The date is parsed as a calendar day; no timezone component. The service uses it directly for schedule matching and entry lookups.
- The response document's `attributes.date` field echoes the input date.

**Ripple:**
- `internal/today/processor.go`: drop `time.Now().In(loc)` argument; accept a parsed `civil.Date`-equivalent (see §5.3).
- `internal/today/resource.go`: delete `tz.Resolve`, `tz.WithLocation`, and the `tz.NewAccountHouseholdLookup` call. Delete the `accountBaseURL` constructor argument; update `cmd/main.go` accordingly.
- `internal/entry/processor.go:53,123`: the "can't log a future-dated entry" gate currently uses `time.Now().UTC().Truncate(24*time.Hour)`. Replace with a parameter threaded from the REST handler via a new `?today=YYYY-MM-DD` query parameter on `PUT /trackers/:item/entries/:date` — **or** remove the check entirely if the frontend already prevents future entries. See §9 Open Question 1.
- `internal/month/processor.go:71`: same treatment via `?today=YYYY-MM-DD` on `GET /trackers/months/:month`.

### 4.2 Workout Service — `GET /workouts/today`

**New contract:** `GET /workouts/today?date=YYYY-MM-DD`
- `date` required; same error semantics as §4.1.
- The service derives `dayOfWeek` from the parsed date using ISO week rules (Monday = 0). It also derives `weekStart` as the Monday of the ISO week containing `date`.

**Ripple:**
- `internal/today/resource.go`: delete `tz.Resolve` + `tz.WithLocation` + lookup; drop `accountBaseURL` constructor argument.
- `internal/today/processor.go`: signature becomes `Today(userID uuid.UUID, date civil.Date)`. `dayOfWeek` and `weekStart` computed from `date` rather than from `time.Now()`.
- `internal/planneditem/processor.go:251`: completion-state helper uses `time.Now().UTC()`. Thread the client's date in instead, via the same `Today(..., date)` path or a dedicated `?today=` param on the consuming endpoint — see §9 Open Question 1.

### 4.3 Productivity Service — Task & Dashboard Summaries

**New contracts:**
- `GET /summary/tasks?date=YYYY-MM-DD` — `date` required.
- `GET /summary/dashboard?date=YYYY-MM-DD` — `date` required.
- `GET /summary/reminders` — unchanged. Reminder counts are timestamp-based (`CURRENT_TIMESTAMP`) and not affected.

**Ripple:**
- `internal/summary/resource.go`: delete the inline `resolveTimezone` helper and its call sites. Replace with `parseDateParam(r)` that returns a `civil.Date` or a 400.
- `internal/summary/processor.go`: `TaskSummary(now time.Time)` becomes `TaskSummary(date civil.Date)`. Downstream `task.CompletedTodayCount(now)` and `task.OverdueCount(now)` similarly accept `civil.Date`.
- `internal/task/provider.go:38-54`: rewrite `countCompletedToday` and `countOverdue` to accept a date instead of a `time.Time`. `countCompletedToday` builds `startOfDay` as midnight UTC on that date; `countOverdue` compares `due_on < date`.

### 4.4 Calendar Service — `GET /calendar/events`

**New contract:** `GET /calendar/events?start=<RFC3339>&end=<RFC3339>`
- Both `start` and `end` are required. Missing values return 400.
- Invalid RFC 3339 values return 400 (previously: silent UTC default when malformed). 

**Ripple:**
- `internal/event/resource.go:43-73`: delete the `tz.Resolve` call and the default-range computation. Parse `start`/`end` as `time.Time`; 400 on failure or if either is absent.
- Delete `internal/tz/` package.

### 4.5 Frontend — Hook & Widget Updates

Every hook that fetches a date-scoped endpoint takes the date as an argument and includes it in the React Query key:

| Hook | New signature | Key change |
|---|---|---|
| `useTrackerToday` | `useTrackerToday(date: string)` | `trackerKeys.today(tenant, household, date)` |
| `useWorkoutToday` | `useWorkoutToday(date: string)` | `workoutKeys.today(tenant, household, date)` |
| `useTaskSummary` | `useTaskSummary(date: string)` | `taskKeys.summary(tenant, household, date)` |
| `useReminderSummary` | unchanged | unchanged |
| `useCalendarEvents` | already takes `start`/`end` | no signature change; ensure widget always supplies both |
| `useMonthSummary` | add optional `today?: string` argument | `trackerKeys.month(tenant, household, month, today)` |

**Widget/page updates** — each caller reads household timezone and computes the date:

```ts
const { household } = useTenant();
const tz = household?.attributes.timezone; // optional — falls back to browser tz
const today = getLocalTodayStr(tz);
const { data } = useTrackerToday(today);
```

Sites to update:
- `src/components/features/trackers/habits-widget.tsx`
- `src/components/features/tracker/today-view.tsx`
- `src/components/features/workouts/workout-widget.tsx`
- `src/pages/WorkoutTodayPage.tsx`
- Dashboard task summary usages (grep `useTaskSummary`)
- `src/components/features/tracker/calendar-grid.tsx` (pass `today` to `useMonthSummary`)

### 4.6 Frontend — API Client Cleanup

**File:** `src/lib/api/client.ts:141-146`

Delete the X-Timezone header injection block. The backend ignores it after this task; sending it is wasted bytes and dead code.

### 4.7 Midnight-Crossing Freshness

Because the date is now part of every query key, React Query will treat `today` at a new local date as a distinct cache entry. The first render after midnight triggers a refetch naturally. No additional timer or invalidation logic is required.

However, the date is computed at render time. If the component is long-mounted (e.g., dashboard left open overnight), the rendered date won't update until something triggers a re-render. Add a lightweight wall-clock poller at a single app-level location (e.g., a `useLocalDate()` hook used by the top-level dashboard or tenant context) that emits a re-render when the local date string changes. Polling every 60 seconds is sufficient.

### 4.8 Documentation Updates

Update per-service `rest.md` to reflect the new contracts:
- `services/tracker-service/docs/rest.md` — `GET /trackers/today` requires `?date=`
- `services/workout-service/docs/rest.md` — `GET /workouts/today` requires `?date=`
- `services/productivity-service/docs/rest.md` — `/summary/tasks` and `/summary/dashboard` require `?date=`
- `services/calendar-service/docs/rest.md` — `GET /calendar/events` requires `?start=` and `?end=`

Remove any mention of `X-Timezone` from these docs.

## 5. API Surface

### 5.1 Request parameter format

All dates use ISO 8601 calendar-date format: `YYYY-MM-DD`. No time component, no timezone suffix. The server parses with Go's `time.Parse("2006-01-02", v)`.

All timestamp ranges use RFC 3339 (`time.RFC3339`). Both bounds required.

### 5.2 Error shape

Missing or malformed date parameters return JSON:API 400:

```json
{
  "errors": [{
    "status": "400",
    "title": "Invalid request",
    "detail": "query parameter 'date' is required and must be YYYY-MM-DD"
  }]
}
```

### 5.3 Parse helper

Add a shared parser — suggest `shared/go/http/params.go` — with:

```go
func ParseDateParam(r *http.Request, name string) (time.Time, error)
```

Returns `time.Date(Y, M, D, 0, 0, 0, 0, time.UTC)` on success; error on missing/malformed. UTC anchor is deliberate: the value represents a calendar day, not an instant, and downstream queries against `type:date` GORM columns compare by date irrespective of location.

## 6. Data Model

No schema changes. No migrations. The household `timezone` column remains on the `account.households` table for use by internal/cron code paths.

## 7. Service Impact

| Service | Change |
|---|---|
| tracker-service | `/trackers/today` takes `?date=`. Month and entry endpoints take `?today=`. Delete `internal/tz/`. Drop `accountBaseURL` constructor wiring. Update `rest.md`. |
| workout-service | `/workouts/today` takes `?date=`. Delete `internal/tz/`. Drop `accountBaseURL` wiring. Update `rest.md`. |
| productivity-service | `/summary/tasks` and `/summary/dashboard` take `?date=`. Delete inline `resolveTimezone`. Update `rest.md`. |
| calendar-service | `/calendar/events` requires `?start=`/`?end=`. Delete `internal/tz/`. Drop `accountBaseURL` wiring. Update `rest.md`. |
| frontend | Hook signatures take date args. Query keys include date. Widgets read household tz + compute date. Remove `X-Timezone` header injection. Add `useLocalDate` poller. |
| shared/go/http | New `ParseDateParam` helper + unit tests. |
| account-service | No changes. |
| auth-service | No changes. |

## 8. Non-Functional Requirements

- **Correctness:** endpoints must behave identically regardless of the server process's local timezone. The test suite must include a test that runs the today processor under `time.Local = time.FixedZone("UTC+14", 14*3600)` and `time.FixedZone("UTC-12", -12*3600)` and verifies the returned date matches the supplied `?date=` parameter.
- **Backwards compatibility:** this is a breaking API change. See `migration-plan.md` for deploy ordering. The frontend build that drops `X-Timezone` and starts sending `?date=` must deploy *after* all backend services accept the new parameter — otherwise mid-deploy users see 400s.
- **Observability:** `400 invalid date parameter` errors should be logged at `info` (not `warn`) to avoid alerting on client bugs, but with enough structure (`endpoint`, `raw_value`) to debug them.
- **Performance:** removes one synchronous account-service round-trip from any request where `X-Timezone` was missing. Net improvement.
- **Tests:** every new parser/processor path needs unit coverage. Frontend hook tests need the new date argument threaded through — expect churn in test fixtures.

## 9. Open Questions

1. **Future-dated entry gating in tracker-service.** The entry processor currently prevents creating entries for future dates using server-side `time.Now()`. Options:
   - (a) Add `?today=YYYY-MM-DD` to the `PUT /trackers/:item/entries/:date` endpoint and have the handler compare `entryDate > today`.
   - (b) Remove the server-side check entirely and trust the client's UI (which already hides future dates).
   - (c) Keep the server-side `time.Now().UTC()` check, accepting that near-UTC-midnight it may be off by a day in either direction.
   
   Recommendation: (a). It preserves the defensive check without reintroducing a timezone chain.

2. **Shared parser location.** `shared/go/http/params.go` or per-service `internal/common/`? Shared is more DRY; per-service respects module boundaries. Recommendation: shared.

3. **Frontend fallback chain for `tz`.** Today: `household?.attributes.timezone ?? Intl.DateTimeFormat().resolvedOptions().timeZone`. Keep this. Do NOT require the household to be loaded before rendering date-scoped widgets — pre-household browser-tz rendering is an acceptable transient state.

## 10. Acceptance Criteria

- [ ] `GET /trackers/today` without `?date=` returns 400; with valid `?date=` returns the items and entries for that calendar day.
- [ ] `GET /workouts/today` without `?date=` returns 400; with valid `?date=` returns the planned items and performances for that day.
- [ ] `GET /summary/tasks` and `GET /summary/dashboard` without `?date=` return 400.
- [ ] `GET /calendar/events` without both `?start=` and `?end=` returns 400.
- [ ] At 8 PM local time on the user's device, the dashboard's habits and workout widgets show today's items — not tomorrow's. Verified on mobile Safari and Chrome.
- [ ] Leaving the dashboard open past midnight causes the widgets to refetch and display the new day's items within 60 seconds, without a manual refresh.
- [ ] The `services/{tracker,workout,calendar}-service/internal/tz/` directories are deleted from the repo.
- [ ] The `resolveTimezone` helper is removed from productivity-service.
- [ ] The `X-Timezone` header injection is removed from `frontend/src/lib/api/client.ts`.
- [ ] Processor tests for all four services pass when `time.Local` is set to a non-UTC zone.
- [ ] `shared/go/http/params.go` provides `ParseDateParam` with unit tests covering valid, empty, malformed, and whitespace-padded inputs.
- [ ] All service `rest.md` files reflect the new contracts; no stale `X-Timezone` references remain.
- [ ] `./scripts/test-all.sh` passes.
- [ ] `./scripts/lint-all.sh` passes.
- [ ] Frontend build passes; no TypeScript errors.
