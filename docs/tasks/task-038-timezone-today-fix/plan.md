# Task 038: Timezone-Aware "Today" Across Frontend & Calendar Service

Last Updated: 2026-04-12

---

## Executive Summary

Multiple frontend components compute "today" using naive `new Date()` with `.toISOString()` (UTC conversion), causing date-boundary mismatches for users west of UTC. The calendar-service backend also lacks the timezone resolution pattern already present in workout-service and tracker-service. This plan creates a shared frontend date utility, replaces all ad-hoc date computations with it, wires up the calendar-utils timezone parameter, and adds timezone resolution to the calendar-service backend.

---

## Current State Analysis

### Frontend

- **6+ components** independently compute "today", "this week", or "this month" using naive `new Date()`.
- `calendar-widget.tsx` has `getTodayRange()` that constructs local midnight then calls `.toISOString()`, shifting the window by the UTC offset (e.g., 4 hours forward for US Eastern).
- `calendar-utils.ts` already has timezone-aware helpers (`getTimeInZone`, `getDateInZone`, `isToday`) but `getEventsForDay()` ignores its `_timezone` parameter.
- `meal-plan-widget.tsx` and `MealsPage.tsx` have duplicate `getMonday()` functions.
- `TrackerPage.tsx` and `calendar-grid.tsx` compute current month/today naively.
- `task.ts` `isTaskOverdue()` uses naive `new Date()` for today comparison.
- The `X-Timezone` header is already sent on every API request via `frontend/src/lib/api/client.ts` line 142-143.
- No centralized date utility (`date-utils.ts`) exists.

### Backend

- **workout-service** and **tracker-service** already have `internal/tz/resolver.go` implementing the timezone resolution pattern (X-Timezone header -> household lookup -> UTC fallback).
- **calendar-service** uses `time.Now().UTC()` for default date range when `start`/`end` params are absent. It does not read `X-Timezone` or resolve household timezone.

---

## Proposed Future State

### Frontend

- A new `frontend/src/lib/date-utils.ts` provides `getLocalToday`, `getLocalTodayStr`, `getLocalTodayRange`, `getLocalWeekStart`, and `getLocalMonth` using `Intl.DateTimeFormat` APIs (zero dependencies).
- All 6+ components use the shared utility instead of inline date computation.
- `calendar-utils.ts` `getEventsForDay()` uses its timezone parameter for day-boundary and all-day event bucketing.
- Duplicate helper functions (`getMonday`, `getTodayStr`, `getCurrentMonth`) are removed from individual components.

### Backend

- `calendar-service` gains `internal/tz/resolver.go` following the same pattern as workout-service and tracker-service.
- `listEventsHandler` resolves timezone from the request and uses it for default start/end computation.
- Invalid `X-Timezone` values never cause 5xx.

---

## Implementation Phases

### Phase 1: Shared Frontend Date Utility (Foundation)

Create the reusable date utility that all subsequent frontend changes depend on.

**Task 1.1: Create `frontend/src/lib/date-utils.ts`** [Effort: M]
- Implement `getLocalToday(tz?: string): Date`
- Implement `getLocalTodayStr(tz?: string): string`
- Implement `getLocalTodayRange(tz?: string): { start: string; end: string }`
- Implement `getLocalWeekStart(tz?: string): Date`
- Implement `getLocalMonth(tz?: string): string`
- Use `Intl.DateTimeFormat` with `timeZone` option to extract correct date components.
- No external date library dependencies.
- Acceptance: All functions return correct values for UTC+N, UTC-N, and DST boundary cases.

**Task 1.2: Unit tests for date utility** [Effort: M]
- Test each function with explicit IANA timezones (e.g., `America/New_York`, `Asia/Tokyo`, `Pacific/Auckland`).
- Test DST boundary (e.g., US spring-forward date at 11:30 PM).
- Test midnight edge case (exactly midnight in the target timezone).
- Test fallback behavior when no timezone is provided.
- Acceptance: Tests pass and cover all specified scenarios.

### Phase 2: Frontend Component Migration

Replace all ad-hoc "today" computations with the shared utility. Each task is independent.

**Task 2.1: Fix calendar widget** [Effort: S]
- File: `frontend/src/components/features/calendar/calendar-widget.tsx`
- Remove inline `getTodayRange()` function (lines 9-14).
- Import and use `getLocalTodayRange()` from `date-utils.ts`.
- Acceptance: Calendar widget sends correct timezone-aware start/end to the API.

**Task 2.2: Fix calendar-utils `getEventsForDay()`** [Effort: S]
- File: `frontend/src/components/features/calendar/calendar-utils.ts`
- Un-prefix `_timezone` parameter in `getEventsForDay()`.
- Use the timezone parameter (with browser fallback) to compute day boundaries and the date string for all-day event comparison.
- Leverage existing `getDateInZone()` helper already in the same file.
- Acceptance: All-day events are bucketed to the correct day in the household timezone.

**Task 2.3: Fix meal plan widget** [Effort: S]
- File: `frontend/src/components/features/meals/meal-plan-widget.tsx`
- Remove inline `getMonday()` and `getTodayStr()` (lines 16-34).
- Import and use `getLocalWeekStart()` and `getLocalTodayStr()` from `date-utils.ts`.
- Acceptance: Meal plan widget shows correct week and highlights today correctly.

**Task 2.4: Fix meals page** [Effort: S]
- File: `frontend/src/pages/MealsPage.tsx`
- Remove inline `getMonday()` and `formatDateStr()` (lines 32-47).
- Import and use `getLocalWeekStart()` from `date-utils.ts`. Use a simple `formatDateStr` based on the Date from `getLocalWeekStart()` or add a format helper to date-utils.
- Acceptance: Meals page initializes to the correct week.

**Task 2.5: Fix tracker page** [Effort: S]
- File: `frontend/src/pages/TrackerPage.tsx`
- Remove inline `getCurrentMonth()` (lines 11-14).
- Import and use `getLocalMonth()` from `date-utils.ts`.
- Acceptance: Tracker page defaults to the correct current month.

**Task 2.6: Fix tracker calendar grid** [Effort: S]
- File: `frontend/src/components/features/tracker/calendar-grid.tsx`
- Replace `new Date().toISOString().slice(0, 10)` (line 91) with `getLocalTodayStr()`.
- Replace `new Date().getFullYear()/getMonth()` currentMonth computation (line 95) with `getLocalMonth()`.
- Replace the same pattern in `MobileDayView()` (line 209).
- Acceptance: Calendar grid highlights the correct "today" and shows the correct month.

**Task 2.7: Fix task overdue logic** [Effort: S]
- File: `frontend/src/types/models/task.ts`
- Replace Date-object comparison with string comparison: `dueOn < getLocalTodayStr()`.
- Acceptance: `isTaskOverdue()` returns correct result for the user's local date.

### Phase 3: Calendar Service Backend

**Task 3.1: Add timezone resolver to calendar-service** [Effort: M]
- Create `services/calendar-service/internal/tz/resolver.go` following the exact pattern from `services/workout-service/internal/tz/resolver.go`.
- Resolution order: X-Timezone header -> household timezone via account-service -> UTC fallback with warn log.
- Cache resolved location on request context.
- Acceptance: Resolver correctly parses valid timezones, gracefully handles invalid ones, and falls back as expected.

**Task 3.2: Update `listEventsHandler` to use timezone** [Effort: S]
- File: `services/calendar-service/internal/event/resource.go`
- When `start` param is absent, resolve timezone via the new resolver and compute start-of-today in that timezone instead of `time.Now().UTC()`.
- Same for `end` param default.
- Acceptance: Default range is timezone-aware; invalid X-Timezone never causes 5xx.

**Task 3.3: Update calendar-service Dockerfile** [Effort: S]
- Verify the calendar-service Dockerfile includes any shared dependencies needed for the tz package (e.g., if it references shared/go packages).
- Acceptance: `docker build` succeeds for calendar-service.

### Phase 4: Verification

**Task 4.1: Run all frontend tests** [Effort: S]
- `npm test` / `npx vitest` in frontend directory.
- Acceptance: All existing + new tests pass.

**Task 4.2: Run calendar-service tests** [Effort: S]
- `go test ./...` in calendar-service directory.
- Acceptance: All tests pass.

**Task 4.3: Docker build verification** [Effort: S]
- Build all affected services via Docker.
- Acceptance: All Docker builds succeed.

**Task 4.4: Manual verification via local deployment** [Effort: M]
- Run `scripts/local-up.sh` and verify:
  - Calendar widget shows only today's events (not tomorrow's all-day events in evening)
  - Meal plan highlights correct day
  - Tracker shows correct month
  - Task overdue indicators are correct
- Acceptance: All scenarios behave correctly.

---

## Risk Assessment and Mitigation

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| `Intl.DateTimeFormat` inconsistency across browsers | Low | Medium | Use well-supported IANA timezone names; fallback to browser timezone when no tz param given |
| Breaking existing calendar event display | Medium | High | Keep API response shape unchanged; only change the default range computation |
| Timezone resolver copy diverges from workout/tracker pattern | Low | Low | Copy the established pattern exactly; consider extracting to shared/go in a future task |
| DST transition edge cases | Medium | Low | Unit tests specifically cover DST boundaries |
| Calendar-service account-service dependency | Low | Medium | X-Timezone header is already sent by frontend, so account-service lookup is rarely needed; graceful fallback to UTC |

---

## Success Metrics

- Zero instances of `new Date().toISOString().slice(0, 10)` or similar naive "today" patterns in frontend code
- Calendar widget event count matches actual today's events for US timezone users
- All 11 acceptance criteria from the PRD are met
- No new third-party dependencies

---

## Required Resources and Dependencies

- **Existing timezone infrastructure**: `X-Timezone` header already sent by frontend API client; workout-service and tracker-service tz resolver pattern exists for copying
- **`Intl.DateTimeFormat` browser API**: Widely supported, no polyfill needed
- **account-service household timezone**: Already exists, no changes needed

---

## Timeline Estimates

| Phase | Effort | Dependencies |
|-------|--------|-------------|
| Phase 1: Shared Date Utility | M | None |
| Phase 2: Frontend Migration | M (7 small tasks) | Phase 1 |
| Phase 3: Calendar Backend | M | None (parallel with Phase 2) |
| Phase 4: Verification | S | Phases 1-3 |

Phases 2 and 3 can proceed in parallel after Phase 1 is complete.
