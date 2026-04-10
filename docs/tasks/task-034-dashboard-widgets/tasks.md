# Task 034: Dashboard "Today" Widgets — Task Checklist

Last Updated: 2026-04-10

---

## Phase 1: Package Summary Empty State

- [ ] **1.1** Update PackageSummaryWidget to show empty state card when no packages exist (S)

## Phase 2: New Widget Components

- [ ] **2.1** Create MealPlanWidget — fetch today's plan, show populated slots with recipe links (M)
- [ ] **2.2** Create CalendarWidget — fetch today's events, chronological with user indicators (M)
- [ ] **2.3** Create HabitsWidget — fetch today's habits with completion status indicators (S)
- [ ] **2.4** Create WorkoutWidget — fetch today's workout, rest day state, done indicators (S)

## Phase 3: Dashboard Integration

- [ ] **3.1** Update DashboardPage layout — 2-column responsive grid with all widgets (M)
- [ ] **3.2** Update PullToRefresh handler to refetch all widget queries (S)
- [ ] **3.3** Ensure refetchOnMount for all widget queries so navigation back shows fresh data (S)

## Phase 4: Polish and Verification

- [ ] **4.1** Verify responsive layout — desktop 2-col grid, mobile vertical stack (S)
- [ ] **4.2** Verify error isolation — one widget failure doesn't affect others (S)
- [ ] **4.3** Build verification — no TypeScript or build errors (S)

---

## PRD Acceptance Criteria Cross-Reference

- [ ] Package summary widget shows empty state card instead of hiding
- [ ] Meal plan widget shows today's slots with recipe name links to `/app/recipes/{id}`
- [ ] Meal plan widget shows empty state when no meals planned
- [ ] Calendar widget shows today's events chronologically with user indicators
- [ ] Calendar widget shows all-day events before timed events
- [ ] Calendar widget shows empty state when no events
- [ ] Habits widget shows today's habits with check/unchecked status
- [ ] Habits widget shows empty state when none scheduled
- [ ] Workout widget shows exercises with done indicators, or rest day state
- [ ] Workout widget shows empty state when no workout and not rest day
- [ ] Each widget header links to its full page
- [ ] All widgets have skeleton loading states
- [ ] All widgets handle API errors without affecting others
- [ ] Desktop: responsive grid layout; Mobile: vertical stack
- [ ] Pull-to-refresh refreshes all widgets
- [ ] Navigating back to dashboard refetches widget data
