# Timezone-Aware "Today" Across Frontend & Calendar Service — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-04-12
---

## 1. Overview

Multiple frontend components compute "today" by constructing a JavaScript `Date` in browser-local time and converting it to an ISO string via `.toISOString()`, which silently shifts the value to UTC. For users west of UTC (all US timezones), this shifts the query window hours into the future — causing tomorrow's all-day calendar events to appear in today's widget, and potentially miscounting which day/week/month the user is in across meals, tracker, and task views.

Task-035 fixed the backend `/today` endpoints in workout-service and tracker-service by adding household-timezone resolution via `X-Timezone` header and account-service fallback. However, the frontend was never updated to match — every component that computes "today" locally still uses raw `new Date()` with UTC conversion. The calendar-service backend also lacks the timezone resolution pattern that the other services now have.

This task creates a shared frontend date utility that computes today/ranges in the household's timezone, replaces all ad-hoc "today" computations with it, and brings the calendar-service backend into alignment with the timezone-aware pattern.

## 2. Goals

Primary goals:
- Eliminate the class of bug where frontend "today" boundaries are shifted by the browser's UTC offset
- Create a single, reusable frontend utility for timezone-aware date computations so this bug cannot recur
- Replace all existing ad-hoc "today" / "this week" / "this month" computations across the frontend with the shared utility
- Add timezone resolution to the calendar-service backend for consistency with workout-service and tracker-service

Non-goals:
- Per-user timezone preferences (household-level is sufficient)
- UI for editing the household timezone (already exists at account-service level)
- Historical data correction or backfilling
- Changes to workout-service or tracker-service `/today` endpoints (already fixed in task-035)

## 3. User Stories

- As a user checking my dashboard in the evening, I want the calendar widget to show only today's events so that tomorrow's all-day events don't appear prematurely.
- As a user viewing my habits/tracker "today" view, I want the date to match my actual local day so entries aren't attributed to the wrong date.
- As a user viewing my meal plan, I want the current week and today's highlight to reflect my actual local day, not a UTC-shifted day.
- As a user viewing my task list, I want overdue indicators to be accurate relative to my actual local day.

## 4. Functional Requirements

### 4.1 Shared Frontend Date Utility

Create a new utility module at `frontend/src/lib/date-utils.ts` (or similar) that provides:

- **`getLocalToday(tz?: string): Date`** — Returns a Date object representing the start of today in the given IANA timezone. If no timezone is provided, falls back to `Intl.DateTimeFormat().resolvedOptions().timeZone` (the browser's timezone).
- **`getLocalTodayStr(tz?: string): string`** — Returns today's date as `YYYY-MM-DD` in the given timezone.
- **`getLocalTodayRange(tz?: string): { start: string; end: string }`** — Returns ISO 8601 start/end timestamps representing the full day of "today" in the given timezone, suitable for passing as query parameters to the backend.
- **`getLocalWeekStart(tz?: string): Date`** — Returns the Monday of the current week in the given timezone.
- **`getLocalMonth(tz?: string): string`** — Returns the current month as `YYYY-MM` in the given timezone.

Implementation notes:
- Use `Intl.DateTimeFormat` with the `timeZone` option to determine the correct local date components, then construct boundaries accordingly.
- The timezone parameter should come from the household context where available. The browser timezone is an acceptable fallback since users are expected to be in the same timezone as their household.
- Do NOT add a dependency on a date library (date-fns, luxon, etc.) — `Intl` APIs are sufficient for this use case.

### 4.2 Calendar Widget Fix (Primary Reported Bug)

**File**: `frontend/src/components/features/calendar/calendar-widget.tsx` (lines 9–14)

**Current behavior**: `getTodayRange()` creates local midnight/end-of-day via `new Date(year, month, day, ...)` then calls `.toISOString()`, which converts to UTC. In US Eastern (UTC-4), this sends `04:00 UTC` to `03:59 UTC+1d` — a window shifted 4 hours into the future.

**Required behavior**: Replace `getTodayRange()` with a call to the shared `getLocalTodayRange()` utility. The resulting ISO strings must represent the actual local day boundaries so that:
- All-day events for tomorrow do NOT appear in today's widget
- All-day events for today DO appear
- Timed events are correctly bounded to today

### 4.3 Calendar Utils All-Day Event Filtering

**File**: `frontend/src/components/features/calendar/calendar-utils.ts` (lines 187–217)

**Current behavior**: `getEventsForDay()` accepts a `_timezone` parameter but ignores it. Day boundaries are computed via `setHours(0,0,0,0)` and `setHours(23,59,59,999)` on the browser-local Date. The date-string comparison for all-day events uses the browser-local `day.getFullYear()/getMonth()/getDate()`.

**Required behavior**: Use the timezone parameter (un-prefix the `_`) to compute day boundaries and the date string in the household timezone. If the timezone parameter is absent, fall back to browser timezone. This ensures all-day events are bucketed to the correct day regardless of browser/server timezone mismatches.

### 4.4 Meal Plan Widget

**File**: `frontend/src/components/features/meals/meal-plan-widget.tsx` (lines 16–34)

**Current behavior**: `getMonday()` and `getTodayStr()` use `new Date()` in browser-local time.

**Required behavior**: Replace with calls to the shared `getLocalWeekStart()` and `getLocalTodayStr()` utilities.

### 4.5 Meals Page

**File**: `frontend/src/pages/MealsPage.tsx` (lines 32–40)

**Current behavior**: `getMonday()` uses `new Date()` in browser-local time.

**Required behavior**: Replace with a call to the shared `getLocalWeekStart()` utility.

### 4.6 Tracker Page

**File**: `frontend/src/pages/TrackerPage.tsx` (lines 11–14)

**Current behavior**: `getCurrentMonth()` uses `new Date()` in browser-local time.

**Required behavior**: Replace with a call to the shared `getLocalMonth()` utility.

### 4.7 Tracker Calendar Grid

**File**: `frontend/src/components/features/tracker/calendar-grid.tsx` (lines 91, 95)

**Current behavior**: `today` is computed via `new Date().toISOString().slice(0, 10)` (UTC date, not local) and `currentMonth` via `new Date().getFullYear()/getMonth()`.

**Required behavior**: Replace with calls to the shared `getLocalTodayStr()` and `getLocalMonth()` utilities.

### 4.8 Task Overdue Logic

**File**: `frontend/src/types/models/task.ts` (lines 45–59)

**Current behavior**: `isTaskOverdue()` constructs today's date using `new Date()` with `setHours(0,0,0,0)` in browser-local time.

**Required behavior**: Replace the today computation with `getLocalTodayStr()` and compare as date strings (`dueDate < todayStr`). This is simpler and avoids the Date object timezone ambiguity entirely.

### 4.9 Calendar Service Backend Timezone Resolution

**File**: `services/calendar-service/internal/event/resource.go` (lines 43, 57)

**Current behavior**: When no `start` query parameter is provided, the handler uses `time.Now().UTC()` to compute the default start of today. This is inconsistent with workout-service and tracker-service, which resolve the household timezone.

**Required behavior**:
- Add a timezone resolution helper to the calendar-service following the same pattern as workout-service and tracker-service (`X-Timezone` header → household timezone via account-service → UTC fallback with warning log).
- When no `start` parameter is provided, compute start-of-today in the resolved timezone instead of UTC.
- Invalid `X-Timezone` values must never cause a 5xx — fall back gracefully with a warn-level log.

## 5. API Surface

No new endpoints. One behavioral change:

- **`GET /calendar/events`**: When called without a `start` parameter, the default start is now computed in the household's timezone (resolved via `X-Timezone` header or account-service lookup) instead of UTC. The `X-Timezone` header is already sent by the frontend (added in task-035). Response shape is unchanged.

## 6. Data Model

No schema changes. No migrations.

## 7. Service Impact

| Service | Change |
|---------|--------|
| **frontend** | New shared date utility; 6+ component updates to use it; calendar-utils timezone parameter wired up |
| **calendar-service** | Add timezone resolution helper; update `listEventsHandler` default range computation |
| **account-service** | No changes — household timezone field already exists |
| **workout-service** | No changes — already fixed in task-035 |
| **tracker-service** | No changes — already fixed in task-035 |

## 8. Non-Functional Requirements

- The shared date utility must have no external dependencies beyond browser `Intl` APIs.
- The calendar-service timezone resolution must not add a synchronous account-service round-trip on every request when `X-Timezone` header is present (header takes priority, no lookup needed). When the header is absent and account-service is consulted, cache per-request.
- Invalid `X-Timezone` values on the calendar-service must never cause a 5xx.
- All existing frontend and backend tests must continue to pass.
- Add unit tests for the shared date utility covering: UTC+N timezone, UTC-N timezone, DST boundary, midnight edge case.

## 9. Open Questions

None — all questions were resolved during scoping.

## 10. Acceptance Criteria

- [ ] A shared frontend date utility exists and is used by all components that compute "today", "this week", or "this month".
- [ ] The calendar widget on the dashboard shows only today's events — an all-day event for tomorrow does NOT appear in the widget, even in the evening hours.
- [ ] The calendar widget and calendar page agree on which events belong to today.
- [ ] The meal plan widget and meals page show the correct current week and highlight today correctly, matching the user's local date.
- [ ] The tracker page defaults to the correct current month in the user's local timezone.
- [ ] The tracker calendar grid highlights the correct "today" cell.
- [ ] `isTaskOverdue()` correctly identifies overdue tasks relative to the user's local date.
- [ ] The calendar-service `GET /calendar/events` endpoint, when called without `start`/`end` parameters, defaults to a range based on the household timezone (via `X-Timezone` header or account-service lookup), not UTC.
- [ ] Invalid `X-Timezone` header values on the calendar-service do not cause 5xx errors.
- [ ] All existing tests pass; new unit tests cover the shared date utility.
- [ ] No new third-party date library dependencies are introduced.
