# Task 042: Task Checklist

Last Updated: 2026-04-16

---

## PR 1 — Backend Accept-Both (Non-Breaking)

- [ ] **1.1** Create `shared/go/http/params.go` with `ParseDateParam(r, name) (time.Time, error)` + unit tests covering valid, missing, empty, malformed, invalid-calendar-date, whitespace-padded inputs [S]
- [ ] **1.2** tracker-service: `/trackers/today` accepts `?date=` when present; fall back to current tz-resolution behavior when absent [S]
- [ ] **1.3** workout-service: `/workouts/today` accepts `?date=` when present; fallback behavior preserved [S]
- [ ] **1.4** productivity-service: `/summary/tasks` and `/summary/dashboard` accept `?date=` when present; fallback behavior preserved [M]
- [ ] **1.5** Add processor/handler tests covering transitional accept-both behavior for the three services [S]
- [ ] **1.6** `./scripts/test-all.sh` passes [S]
- [ ] **1.7** `./scripts/lint-all.sh` and `./scripts/build-all.sh` pass; Docker builds succeed [S]
- [ ] **1.8** Merge and deploy PR 1 to prod

---

## PR 2 — Frontend Switch (Non-Breaking Against PR 1 Backend)

Depends on PR 1 deployed.

- [ ] **2.1** Add `frontend/src/lib/hooks/use-local-date.ts` + test with fake timers crossing midnight [S]
- [ ] **2.2** `useTrackerToday(date: string)` — update hook, service, and query key [S]
- [ ] **2.3** `useWorkoutToday(date: string)` — update hook, service, and query key [S]
- [ ] **2.4** `useTaskSummary(date: string)` — update hook, service, and query key [S]
- [ ] **2.5** `useMonthSummary(month, today?)` — update hook, query key, service [S]
- [ ] **2.6** `HabitsWidget` and `WorkoutWidget`: compute date via `useLocalDate`; pass to hook [S]
- [ ] **2.7** `TodayView` and `WorkoutTodayPage`: same pattern [S]
- [ ] **2.8** `CalendarGrid` (tracker): thread `today` into `useMonthSummary` [S]
- [ ] **2.9** Dashboard task summary callers: thread `useLocalDate` value through each [S]
- [ ] **2.10** Delete `X-Timezone` header injection from `frontend/src/lib/api/client.ts:141-146` [S]
- [ ] **2.11** Update frontend hook tests for new date arg signatures [M]
- [ ] **2.12** Manual verification: `./scripts/local-up.sh`; simulate 8 PM Eastern; midnight rollover; confirm Network tab has `?date=` and no `X-Timezone` [M]
- [ ] **2.13** `npm run build`, `./scripts/lint-all.sh`; Docker frontend build succeeds [S]
- [ ] **2.14** Merge and deploy PR 2 to prod; wait ≥ 1 hour; verify `X-Timezone` traffic is zero in access logs

---

## PR 3 — Backend Require + Cleanup (Breaking)

Depends on PR 2 soaked in prod for ≥ 1 hour with no `X-Timezone` traffic remaining.

### tracker-service

- [ ] **3.1** `/trackers/today`: require `?date=`; use `ParseDateParam`; return 400 on error [M]
- [ ] **3.2** Delete `services/tracker-service/internal/tz/` directory [S]
- [ ] **3.3** Remove `accountBaseURL` from `InitializeRoutes` call for today route; update `cmd/main.go` [S]
- [ ] **3.4** `internal/entry/processor.go:53,123`: accept `today time.Time` param; plumb via `?today=` on PUT entry [S]
- [ ] **3.5** `internal/month/processor.go:71`: accept `today time.Time` param; plumb via `?today=` on month handler [S]
- [ ] **3.6** Add processor test with `time.Local` set to `America/Los_Angeles` proving tz independence [S]

### workout-service

- [ ] **3.7** `/workouts/today`: require `?date=`; use `ParseDateParam` [M]
- [ ] **3.8** Delete `services/workout-service/internal/tz/` directory [S]
- [ ] **3.9** Remove `accountBaseURL` wiring in `cmd/main.go` [S]
- [ ] **3.10** `internal/planneditem/processor.go:251`: thread client date through if on request path, else confirm UTC is acceptable [S]
- [ ] **3.11** Add tz-independence test [S]

### productivity-service

- [ ] **3.12** Delete `resolveTimezone` function in `internal/summary/resource.go` [S]
- [ ] **3.13** `taskSummaryHandler` and `dashboardSummaryHandler`: use `ParseDateParam`; 400 on error [S]
- [ ] **3.14** `TaskSummary(date time.Time)`; update processor signatures [S]
- [ ] **3.15** `countOverdue(db, date)` and `countCompletedToday(db, date)` in `internal/task/provider.go` [S]
- [ ] **3.16** Add tz-independence test [S]

### calendar-service

- [ ] **3.17** `/calendar/events`: require both `?start` and `?end`; 400 on missing, malformed, or `end <= start` [M]
- [ ] **3.18** Delete default-range fallback block in `internal/event/resource.go` [S]
- [ ] **3.19** Delete `services/calendar-service/internal/tz/` directory [S]
- [ ] **3.20** Remove `accountBaseURL` wiring in `cmd/main.go` [S]
- [ ] **3.21** Add test confirming 400 on missing/malformed params [S]

### Docs & collections

- [ ] **3.22** Update `services/tracker-service/docs/rest.md` for `?date=` and `?today=`; remove X-Timezone mentions [S]
- [ ] **3.23** Update `services/workout-service/docs/rest.md`; remove X-Timezone mentions [S]
- [ ] **3.24** Update `services/productivity-service/docs/rest.md`; remove X-Timezone mentions [S]
- [ ] **3.25** Update `services/calendar-service/docs/rest.md`; remove X-Timezone mentions [S]
- [ ] **3.26** Update Bruno collections under `bruno/` to include new required params [S]

### Final verification

- [ ] **3.27** `./scripts/test-all.sh` passes [S]
- [ ] **3.28** `./scripts/lint-all.sh` passes [S]
- [ ] **3.29** `./scripts/build-all.sh` + Docker builds pass for tracker, workout, productivity, calendar [S]
- [ ] **3.30** `grep -r "tz\.Resolve\|X-Timezone" services/ shared/ frontend/src/` returns nothing [S]
- [ ] **3.31** Merge and deploy PR 3 to prod; smoke-test each affected endpoint with and without params [S]
