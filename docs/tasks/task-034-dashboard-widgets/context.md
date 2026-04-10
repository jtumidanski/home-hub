# Task 034: Dashboard "Today" Widgets — Context

Last Updated: 2026-04-10

---

## Key Files

### Files to Modify

| File | Change |
|------|--------|
| `frontend/src/pages/DashboardPage.tsx` | Add new widgets, update layout to 2-col grid, extend PullToRefresh |
| `frontend/src/components/features/packages/package-summary-widget.tsx` | Add empty state instead of hiding |

### Files to Create

| File | Purpose |
|------|---------|
| `frontend/src/components/features/meals/meal-plan-widget.tsx` | Meal plan today widget |
| `frontend/src/components/features/calendar/calendar-widget.tsx` | Calendar events today widget |
| `frontend/src/components/features/trackers/habits-widget.tsx` | Habits today widget |
| `frontend/src/components/features/workouts/workout-widget.tsx` | Workout today widget |

### Reference Files (read-only, for patterns and types)

| File | Why |
|------|-----|
| `frontend/src/components/features/weather/weather-widget.tsx` | Widget pattern reference (loading, error, empty states) |
| `frontend/src/lib/hooks/api/use-meals.ts` | `usePlans()` hook — fetches meal plans by starts_on |
| `frontend/src/lib/hooks/api/use-calendar.ts` | `useCalendarEvents(start, end)` hook |
| `frontend/src/lib/hooks/api/use-trackers.ts` | `useTrackerToday()` hook |
| `frontend/src/lib/hooks/api/use-workouts.ts` | `useWorkoutToday()` hook |
| `frontend/src/lib/hooks/api/use-packages.ts` | `usePackageSummary()` hook |
| `frontend/src/components/ui/card.tsx` | Card composition API |
| `frontend/src/components/common/error-card.tsx` | Error card pattern |
| `frontend/src/components/common/pull-to-refresh.tsx` | PullToRefresh pattern |
| `frontend/src/context/tenant-context.tsx` | Tenant/household context |

### Existing "Today" Pages (for data shape reference)

| File | Why |
|------|-----|
| `frontend/src/pages/TrackerPage.tsx` (or TodayView component) | How tracker today data is consumed |
| `frontend/src/pages/WorkoutTodayPage.tsx` | How workout today data is consumed |
| `frontend/src/pages/MealPlanPage.tsx` (or similar) | How meal plan data is structured |
| `frontend/src/pages/CalendarPage.tsx` | How calendar events are rendered |

## Key Decisions

1. **No new API hooks needed for core data** — all four data sources have existing hooks (`usePlans`, `useCalendarEvents`, `useTrackerToday`, `useWorkoutToday`). The meal plan widget may need a helper to calculate week start date and extract today's items.

2. **Independent widget fetching** — each widget owns its query lifecycle. No shared loading gate. React Query handles deduplication if the same data is used elsewhere.

3. **refetchOnMount: 'always'** — ensures navigating back to dashboard shows fresh data after interacting on detail pages. This may need to be passed as an option if the existing hooks don't set it.

4. **2-column responsive grid** — desktop uses `md:grid-cols-2`, mobile defaults to single column. Weather stays full-width above the grid.

5. **Empty state pattern** — consistent across all widgets: icon + descriptive text in a Card. Package widget adopts this pattern instead of hiding.

## Dependencies

- No backend changes required
- No new npm packages required
- All UI primitives (Card, Skeleton, icons) already available
- All API hooks already exist

## Data Shape Notes

### Meal Plans
- `usePlans({ starts_on })` returns plans for a week starting on that date
- Plan items have day-of-week and slot (breakfast, lunch, dinner, snack, side)
- Need to compute Monday of current week as `starts_on` parameter
- Need to filter items for today's day-of-week

### Calendar Events
- `useCalendarEvents(start, end)` with ISO date strings
- Events have: start/end times, title, allDay flag, user info (name/color)
- Sort: all-day first, then by start time

### Tracker Today
- `useTrackerToday()` returns scheduled items with entry data
- An item is "complete" if it has an entry for today
- Items include: name, color, scale type

### Workout Today
- `useWorkoutToday()` returns: `{ date, isRestDay, items[], weekStartDate }`
- Items include: exercise name, planned details, performance data
- Performance presence indicates "done"
