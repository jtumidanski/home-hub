# Task 038: Task Checklist

Last Updated: 2026-04-12

---

## Phase 1: Shared Frontend Date Utility

- [ ] **1.1** Create `frontend/src/lib/date-utils.ts` with `getLocalToday`, `getLocalTodayStr`, `getLocalTodayRange`, `getLocalWeekStart`, `getLocalMonth` using `Intl.DateTimeFormat` [M]
- [ ] **1.2** Add unit tests for date utility covering UTC+N, UTC-N, DST boundary, midnight edge case [M]

## Phase 2: Frontend Component Migration

- [ ] **2.1** Replace `getTodayRange()` in `calendar-widget.tsx` with `getLocalTodayRange()` [S]
- [ ] **2.2** Wire up `_timezone` param in `calendar-utils.ts` `getEventsForDay()` using `getDateInZone()` [S]
- [ ] **2.3** Replace `getMonday()`/`getTodayStr()` in `meal-plan-widget.tsx` with shared utils [S]
- [ ] **2.4** Replace `getMonday()`/`formatDateStr()` in `MealsPage.tsx` with shared utils [S]
- [ ] **2.5** Replace `getCurrentMonth()` in `TrackerPage.tsx` with `getLocalMonth()` [S]
- [ ] **2.6** Replace naive today/month computations in `calendar-grid.tsx` (lines 91, 95, 209) [S]
- [ ] **2.7** Simplify `isTaskOverdue()` in `task.ts` to use `getLocalTodayStr()` string comparison [S]

## Phase 3: Calendar Service Backend

- [ ] **3.1** Create `services/calendar-service/internal/tz/resolver.go` following workout-service pattern [M]
- [ ] **3.2** Update `listEventsHandler` in `resource.go` to use timezone resolver for default range [S]
- [ ] **3.3** Verify calendar-service Dockerfile builds successfully [S]

## Phase 4: Verification

- [ ] **4.1** All frontend tests pass (`npm test` / `npx vitest`) [S]
- [ ] **4.2** All calendar-service tests pass (`go test ./...`) [S]
- [ ] **4.3** Docker builds succeed for all affected services [S]
- [ ] **4.4** Manual verification via `scripts/local-up.sh` — calendar widget, meal plan, tracker, task overdue all correct [M]
