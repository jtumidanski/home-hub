# Dashboard "Today" Widgets — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-04-10
---

## 1. Overview

The Dashboard page currently shows a weather widget, a package summary (hidden when empty), and task/reminder count cards. Users have to navigate to individual pages to see what their day looks like across meals, calendar, habits, and workouts.

This feature adds four new "today at a glance" widgets to the dashboard — today's meal plan, today's calendar events, today's habits, and today's workouts — and updates the existing package summary widget to show an empty state instead of hiding. Each widget is read-only and links to its respective full page for interaction.

The goal is to make the dashboard a single place to see the shape of the day without navigating away.

## 2. Goals

Primary goals:
- Give users a complete "today" overview on a single page
- Each widget links to its full page for deeper interaction
- Consistent empty states when no data exists for today
- Responsive layout: sensible multi-column on desktop, vertical stack on mobile

Non-goals:
- No interactive features (editing meals, logging habits, logging workouts) from the dashboard
- No widget customization, reordering, or hide/show preferences
- No new backend API endpoints (all required endpoints already exist)
- No push notifications or real-time updates

## 3. User Stories

- As a user, I want to see today's meal plan on the dashboard so I know what to cook without navigating to the meals page
- As a user, I want to see today's calendar events across the household so I know what everyone has scheduled
- As a user, I want to see my habits for today and their completion status so I can track progress at a glance
- As a user, I want to see my workout plan for today so I know what exercises are planned
- As a user, I want to see an empty state for packages when none are tracked, so the widget area isn't mysteriously absent

## 4. Functional Requirements

### 4.1 Package Summary Widget Update

- **Current behavior**: Widget is hidden when no packages exist (`arrivingTodayCount`, `inTransitCount`, and `exceptionCount` are all 0)
- **New behavior**: Show an empty state card when no packages exist (e.g., "No packages being tracked" with a package icon)
- The widget continues to link to `/app/packages`

### 4.2 Meal Plan Widget

- Fetch today's meal plan by calculating the week start date for today and querying `GET /api/v1/meals/plans?starts_on={weekStart}`. There may be gaps — not every week has a plan.
- Display only populated meal slots for today (breakfast, lunch, dinner, snack, side)
- Each meal item links to its recipe page (`/app/recipes/{recipeId}`)
- Show slot label (e.g., "Lunch") and recipe name for each item
- **Empty state**: "No meals planned for today" when no plan covers today or no slots are populated for today
- Widget header links to `/app/meals`

### 4.3 Calendar Widget

- Fetch today's events via `GET /api/v1/calendar/events?start={todayStart}&end={todayEnd}`
- Display events in a flat chronological timeline
- Each event shows: time (or "All Day"), title, and user indicator (name or color dot) so the user can distinguish whose event it is
- All-day events appear first, then time-based events sorted by start time
- Silently show whatever is available — do not indicate missing or disconnected calendars
- **Empty state**: "No events today" when no calendar events exist for today
- Widget header links to `/app/calendar`

### 4.4 Habits Widget

- Fetch today's habits via `GET /api/v1/trackers/today`
- Display each scheduled habit name with completion status (checkmark or empty indicator)
- A habit is "complete" if it has a corresponding entry in the today response
- **Empty state**: "No habits scheduled for today" when no habits are scheduled
- Widget header links to `/app/habits`
- Scoped to the current user (the `/trackers/today` endpoint is already user-scoped via auth token)

### 4.5 Workout Widget

- Fetch today's workout via `GET /api/v1/workouts/today`
- If it is a rest day, display a "Rest Day" indicator
- If not a rest day, list exercise names for today's planned items
- Show a subtle "done" indicator on exercises that have logged performance
- **Empty state**: "No workout planned for today" when no items exist and it's not a rest day
- Widget header links to `/app/workouts`
- Scoped to the current user (the `/workouts/today` endpoint is already user-scoped via auth token)

## 5. API Surface

No new endpoints required. Existing endpoints used:

| Endpoint | Service | Purpose |
|----------|---------|---------|
| `GET /api/v1/packages/summary` | package-service | Package counts (existing) |
| `GET /api/v1/meals/plans` | recipe-service | Find meal plan covering today |
| `GET /api/v1/meals/plans/{planId}` | recipe-service | Get plan items by day/slot |
| `GET /api/v1/calendar/events?start=&end=` | calendar-service | Today's events across household |
| `GET /api/v1/trackers/today` | tracker-service | Today's habits + entries |
| `GET /api/v1/workouts/today` | workout-service | Today's workout plan |

## 6. Data Model

No data model changes. All data is fetched from existing APIs.

## 7. Service Impact

| Service | Change |
|---------|--------|
| **Frontend** | New widget components, new API hooks (meals today, calendar today, habits today, workouts today), updated `DashboardPage` layout, updated `PackageSummaryWidget` empty state |
| **Backend** | None |

## 8. Non-Functional Requirements

- **Performance**: Each widget fetches data independently. Use TanStack React Query with 60-second `staleTime` consistent with existing dashboard hooks. Widgets render independently (one slow API doesn't block others).
- **Refetch on navigation**: When the user navigates back to the dashboard (e.g., after logging a habit or workout), queries should refetch rather than serving stale cache. Use `refetchOnMount: 'always'` or equivalent so returning to the dashboard always shows fresh data.
- **Loading states**: Each widget shows a skeleton loader while its data loads (consistent with existing patterns).
- **Error states**: Each widget handles its own errors independently with an inline error indicator. One widget failing doesn't affect others.
- **Mobile**: Vertical stack layout. Widgets should be touch-friendly.
- **Desktop**: Use a responsive grid layout (e.g., 2-column) that makes good use of horizontal space while maintaining content readability. Group widgets with similar content density together.

## 9. Open Questions

None — all questions resolved during scoping.

## 10. Acceptance Criteria

- [ ] Package summary widget shows an empty state card when no packages are tracked, instead of hiding
- [ ] Meal plan widget shows today's populated meal slots with recipe names linking to `/app/recipes/{id}`
- [ ] Meal plan widget shows empty state when no meals are planned for today
- [ ] Calendar widget shows today's events in chronological order with user name/color indicators
- [ ] Calendar widget shows all-day events before timed events
- [ ] Calendar widget shows empty state when no events exist today
- [ ] Habits widget shows today's scheduled habits with completion status (check/unchecked)
- [ ] Habits widget shows empty state when no habits are scheduled
- [ ] Workout widget shows exercise names for today, with a "done" indicator on exercises with logged performance, or a "Rest Day" indicator
- [ ] Workout widget shows empty state when no workout is planned and it's not a rest day
- [ ] Each widget header links to its respective full page
- [ ] All widgets have skeleton loading states
- [ ] All widgets handle API errors gracefully without affecting other widgets
- [ ] Desktop layout uses a responsive grid; mobile layout is a vertical stack
- [ ] Pull-to-refresh on the dashboard refreshes all widgets including new ones
- [ ] Navigating back to the dashboard refetches widget data (no stale cache from prior visit)
