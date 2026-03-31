# Task 017 — UI Polish & Cleanup: Task Checklist

Last Updated: 2026-03-31

## Phase 1: Quick Wins

### 1.1 Remove meal type badge from week grid [S]
- [ ] Delete the `recipe_classification` badge block in `week-grid.tsx` (lines 113-116)
- **Acceptance**: Meal planner grid cards show no meal-type badge
- **Deps**: None

### 1.2 Fix status filter dropdown label [S]
- [ ] In `list-filter-bar.tsx`, ensure the `SelectTrigger` displays "All statuses" when value is "all" (not the raw value)
- **Acceptance**: Closed dropdown shows "All statuses"; open dropdown items unchanged
- **Deps**: None

### 1.3 Title-case recipe tags [S]
- [ ] Create `toTitleCase()` utility in `frontend/src/lib/utils.ts` (or similar)
- [ ] Apply to tag text in `recipe-card.tsx`, `RecipeDetailPage.tsx`, and `RecipesPage.tsx`
- **Acceptance**: Tags like "quick meal" render as "Quick Meal" across all three locations
- **Deps**: None

### 1.4 Remove dashboard role subtitle [S]
- [ ] Delete the `resolvedRole` paragraph in `DashboardPage.tsx` (lines 52-54)
- **Acceptance**: Dashboard shows only "Dashboard" heading, no subtitle
- **Deps**: None

### 1.5 Align household members page header [S]
- [ ] Replace `<Link>`-wrapped icon button with `<Button variant="ghost" size="sm">` + arrow + "Households" text
- [ ] Switch from `<Link>` to `useNavigate()` + `onClick` to match recipe/ingredient pattern
- **Acceptance**: Header matches `IngredientDetailPage` / `RecipeDetailPage` pattern
- **Deps**: None

## Phase 2: Backend Bug Fix

### 2.1 Fix ParseMinutes for compound durations [M]
- [ ] Rewrite `ParseMinutes()` in `cooklang/parser.go` to handle: `"1h 20m"` → 80, `"1h20m"` → 80, `"2h"` → 120, `"45m"` → 45, `"90 minutes"` → 90, `"1.5 hours"` → 90, `"1 hour 20 minutes"` → 80
- [ ] Add test cases for all above formats in `parser_test.go`
- [ ] Run `go test ./...` for recipe-service
- **Acceptance**: All new test cases pass; existing tests still pass
- **Deps**: None

## Phase 3: Calendar Popovers

### 3.1 Install shadcn Calendar and Popover components [S]
- [ ] Run `npx shadcn@latest add calendar popover` in `frontend/`
- [ ] Verify `react-day-picker` and `date-fns` are added to dependencies
- [ ] Verify build succeeds
- **Acceptance**: `Calendar` and `Popover` components available in `frontend/src/components/ui/`
- **Deps**: None

### 3.2 Add calendar popover to meal planner week selector [M]
- [ ] Wrap the date range label in `week-selector.tsx` with a Popover
- [ ] Popover content: shadcn Calendar component
- [ ] On date select: compute the Monday of that week, call `onWeekChange`, close popover
- [ ] Keep existing prev/next buttons unchanged
- **Acceptance**: Clicking date range opens calendar; selecting any date navigates to that week
- **Deps**: 3.1

### 3.3 Add calendar popover to calendar page "Today" widget [M]
- [ ] Replace the "Today" button in `CalendarPage.tsx` with a Popover trigger
- [ ] Popover content: shadcn Calendar component
- [ ] On date select: navigate to week containing that date, close popover
- [ ] Keep existing prev/next buttons unchanged
- **Acceptance**: Clicking "Today" area opens calendar; selecting a date navigates to that week
- **Deps**: 3.1

## Phase 4: Ingredients Pagination

### 4.1 Add pagination to ingredients page [M]
- [ ] Add `page` state to `IngredientsPage.tsx`
- [ ] Pass `page` and `pageSize` to `useIngredients` hook
- [ ] Extract `meta.total`, `meta.page`, `meta.pageSize` from API response
- [ ] Display "Showing X-Y of Z ingredients" summary text
- [ ] Add prev/next page buttons with disabled states for first/last page
- [ ] Reset to page 1 when search query changes
- **Acceptance**: Pagination controls visible; count displayed; navigating pages works; search resets to page 1
- **Deps**: None

## Verification

- [ ] `recipe-service` Go tests pass
- [ ] Frontend builds with no type errors
- [ ] Docker builds succeed for frontend and recipe-service
