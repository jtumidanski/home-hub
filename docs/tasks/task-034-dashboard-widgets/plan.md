# Task 034: Dashboard "Today" Widgets — Implementation Plan

Last Updated: 2026-04-10

---

## Executive Summary

Add four new read-only "today at a glance" widgets to the dashboard (Meal Plan, Calendar, Habits, Workouts) and update the Package Summary widget to show an empty state instead of hiding. This is a frontend-only change — all required API endpoints already exist with corresponding React Query hooks.

## Current State Analysis

### Dashboard Layout (`frontend/src/pages/DashboardPage.tsx`)
- PullToRefresh wrapper with vertical stack (`space-y-6`)
- WeatherWidget (full width)
- PackageSummaryWidget (conditionally hidden when no packages)
- 3-column stat card grid: Pending Tasks, Active Reminders, Overdue Tasks
- Loading: DashboardSkeleton with skeleton boxes
- Refresh: awaits `refetchTasks()` + `refetchReminders()`

### Existing Hooks (all exist, no new hooks needed for core data)
| Hook | File | staleTime | Notes |
|------|------|-----------|-------|
| `usePackageSummary()` | `use-packages.ts` | 60s | Returns arrivingTodayCount, inTransitCount, exceptionCount |
| `usePlans({ starts_on })` | `use-meals.ts` | 5m | Fetches plans by week start date |
| `useCalendarEvents(start, end)` | `use-calendar.ts` | 60s | Events in date range |
| `useTrackerToday()` | `use-trackers.ts` | 30s | Today's habits + entries |
| `useWorkoutToday()` | `use-workouts.ts` | 30s | Today's workout (isRestDay, items[]) |

### Component Patterns
- Card composition: `Card > CardHeader > CardTitle + CardAction > CardContent`
- Error: `ErrorCard` with destructive border
- Skeleton: `animate-pulse rounded-md bg-muted` divs
- Icons: lucide-react
- Navigation: `useNavigate()` or `Link`

## Proposed Future State

### Dashboard Layout (Desktop)
```
[ Weather Widget (full width)                                    ]
[ Package Summary    |  Calendar Events                          ]
[ Meal Plan          |  Habits                                   ]
[ Workout            |  Tasks/Reminders Stats                    ]
```

Desktop uses a 2-column responsive grid (`md:grid-cols-2`). Mobile stacks vertically.

### Widget Behavior
- Each widget fetches data independently via its own React Query hook
- Each widget has: loading skeleton, error state, empty state, populated state
- Each widget header links to its full page
- `refetchOnMount: 'always'` ensures fresh data when navigating back to dashboard
- PullToRefresh refreshes all widget queries

---

## Implementation Phases

### Phase 1: Package Summary Widget Empty State

Update the existing `PackageSummaryWidget` to show an empty state card instead of returning `null`.

**Rationale**: Smallest change, establishes the empty-state pattern for other widgets.

### Phase 2: New Widget Components

Create four new widget components, each following the same pattern:
1. Meal Plan Widget
2. Calendar Widget
3. Habits Widget
4. Workout Widget

Each widget is a self-contained component with its own data fetching, loading, error, and empty states.

### Phase 3: Dashboard Integration

Wire all widgets into `DashboardPage.tsx`:
- Update layout to responsive 2-column grid
- Add all new widgets
- Update PullToRefresh to refetch all widget queries
- Ensure independent rendering (one slow/failed widget doesn't block others)

### Phase 4: Polish and Verification

- Verify mobile layout (vertical stack)
- Verify desktop layout (2-column grid)
- Verify empty states for all widgets
- Verify error isolation
- Verify refetch on navigation back to dashboard
- Build verification

---

## Detailed Tasks

### Phase 1: Package Summary Empty State

**1.1 Update PackageSummaryWidget to show empty state** — Size: S
- File: `frontend/src/components/features/packages/package-summary-widget.tsx`
- Remove the `if (!hasPackages) return null;` guard
- Add an empty state card: package icon + "No packages being tracked" text
- Keep existing loading state (`return null` while loading is acceptable, or show skeleton)
- Keep link to `/app/packages`
- **Acceptance**: Widget renders a card with empty message when all counts are 0

### Phase 2: Widget Components

**2.1 Create MealPlanWidget** — Size: M
- File: `frontend/src/components/features/meals/meal-plan-widget.tsx`
- Calculate week start (Monday) for today's date
- Use `usePlans({ starts_on: weekStartDate })` to fetch current week's plan
- Extract today's day-of-week, filter plan items for today
- Display populated slots: slot label (Breakfast, Lunch, etc.) + recipe name
- Each recipe name links to `/app/recipes/{recipeId}`
- Header links to `/app/meals`
- Loading: skeleton card
- Error: inline error indicator
- Empty: "No meals planned for today"
- Use `refetchOnMount: 'always'` or set on the query
- **Acceptance**: Shows today's meal slots with recipe links; empty state when no plan/slots

**2.2 Create CalendarWidget** — Size: M
- File: `frontend/src/components/features/calendar/calendar-widget.tsx`
- Calculate today's start/end timestamps
- Use `useCalendarEvents(todayStart, todayEnd)`
- Sort: all-day events first, then by start time
- Display: time (or "All Day"), title, user indicator (name or color dot)
- Header links to `/app/calendar`
- Loading: skeleton card
- Error: inline error indicator
- Empty: "No events today"
- Use `refetchOnMount: 'always'` or set on the query
- **Acceptance**: Shows today's events chronologically with user indicators; all-day events first

**2.3 Create HabitsWidget** — Size: S
- File: `frontend/src/components/features/trackers/habits-widget.tsx`
- Use `useTrackerToday()` (already exists, returns today's scheduled items + entries)
- Display each habit name with completion status (checkmark or empty circle)
- Header links to `/app/habits`
- Loading: skeleton card
- Error: inline error indicator
- Empty: "No habits scheduled for today"
- Use `refetchOnMount: 'always'` or set on the query
- **Acceptance**: Shows habits with check/unchecked indicators; empty state when none scheduled

**2.4 Create WorkoutWidget** — Size: S
- File: `frontend/src/components/features/workouts/workout-widget.tsx`
- Use `useWorkoutToday()` (already exists, returns isRestDay, items[])
- If rest day: show "Rest Day" indicator
- If not rest day: list exercise names with "done" indicator on exercises with logged performance
- Header links to `/app/workouts`
- Loading: skeleton card
- Error: inline error indicator
- Empty: "No workout planned for today" (when no items and not rest day)
- Use `refetchOnMount: 'always'` or set on the query
- **Acceptance**: Shows exercises with done indicators, rest day state, or empty state

### Phase 3: Dashboard Integration

**3.1 Update DashboardPage layout** — Size: M
- File: `frontend/src/pages/DashboardPage.tsx`
- Add imports for all four new widget components
- Restructure layout:
  - Weather widget (full width, as-is)
  - Responsive 2-column grid (`grid grid-cols-1 md:grid-cols-2 gap-4`) containing:
    - PackageSummaryWidget
    - CalendarWidget
    - MealPlanWidget
    - HabitsWidget
    - WorkoutWidget
    - Existing stat cards (tasks/reminders) — may need to be wrapped in a card
- Mobile: single column stack (handled by `grid-cols-1` default)
- **Acceptance**: Desktop shows 2-column grid; mobile shows vertical stack

**3.2 Update PullToRefresh handler** — Size: S
- File: `frontend/src/pages/DashboardPage.tsx`
- Extend `handleRefresh` to also refetch: meal plans, calendar events, tracker today, workout today, package summary
- Each refetch runs in parallel via `Promise.all`
- **Acceptance**: Pull-to-refresh refreshes all widgets including new ones

**3.3 Ensure refetch on mount** — Size: S
- Verify that each new widget's query uses `refetchOnMount: 'always'` (or the hook is configured that way)
- This ensures navigating back to dashboard shows fresh data
- May need to pass query options to existing hooks or wrap them
- **Acceptance**: Navigating away and back to dashboard triggers data refetch

### Phase 4: Polish and Verification

**4.1 Verify responsive layout** — Size: S
- Desktop: 2-column grid with sensible grouping
- Mobile: vertical stack, touch-friendly spacing
- All widgets readable at both breakpoints

**4.2 Verify error isolation** — Size: S
- One widget failing (API error) doesn't affect others
- Each widget shows its own error state inline

**4.3 Build verification** — Size: S
- Run TypeScript build (`npm run build` or equivalent in frontend)
- Verify no type errors or build failures

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Meal plan data structure differs from expectations | Low | Medium | Read existing meal plan components/types before implementing |
| Calendar event user indicator data not available | Low | Medium | Check calendar event response shape; fall back to name-only |
| Too many concurrent API calls on dashboard load | Low | Low | React Query deduplicates; staleTime prevents re-fetches within window |
| Layout looks cramped on small screens | Medium | Low | Test at mobile breakpoints; vertical stack prevents crowding |

## Success Metrics

- All 10 acceptance criteria from PRD section 10 pass
- Dashboard load time does not regress noticeably (widgets load independently)
- No build errors
- Empty states display correctly for all widgets

## Required Resources and Dependencies

- **Frontend only** — no backend changes
- Existing React Query hooks for all data sources
- shadcn/ui Card components
- lucide-react icons
- Tailwind CSS responsive utilities

## Timeline Estimate

| Phase | Effort |
|-------|--------|
| Phase 1: Package empty state | S |
| Phase 2: Four widget components | M (cumulative) |
| Phase 3: Dashboard integration | M |
| Phase 4: Polish & verification | S |
| **Total** | **M-L** |
