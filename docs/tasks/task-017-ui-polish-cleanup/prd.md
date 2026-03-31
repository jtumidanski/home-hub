# UI Polish & Cleanup — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-03-31
---

## 1. Overview

This task covers a batch of nine quality-of-life improvements across the Home Hub frontend and one backend bug fix. The items address redundant UI elements, incorrect data, missing pagination, inconsistent labels, and UX misalignment between pages.

Each item is small in isolation, but together they meaningfully improve the polish and consistency of the application. The fixes span the Meal Planner, Recipes, Ingredients, Calendar, Tasks/Reminders, Dashboard, and Households pages.

## 2. Goals

Primary goals:
- Remove redundant or unnecessary UI elements (meal type badge, dashboard subtitle)
- Fix the cooklang time parser to correctly handle compound durations
- Surface ingredient pagination and total counts
- Add calendar-based week selection to Meal Planner and Calendar pages
- Ensure label consistency in filter dropdowns and recipe tags
- Align page header patterns across all detail/member pages

Non-goals:
- Reprocessing existing recipes to update stored times (user will resave manually)
- Redesigning the ingredients page layout beyond adding pagination
- Changing the structure of the filter bar component beyond the label fix

## 3. User Stories

- As a user viewing the meal planner, I don't need a redundant meal-type badge since the grid column already indicates the meal type.
- As a user, I want recipe total time to be correctly calculated so I can plan my cooking schedule.
- As a user browsing ingredients, I want to see how many ingredients exist and navigate through pages.
- As a user, I want to click the week label on the Meal Planner or Calendar page and pick any week from a calendar popup.
- As a user, I want filter dropdown labels to match their selected value consistently.
- As a user, I want recipe tags to be visually consistent with title-cased words.
- As a user, I don't need to see "You are the owner" on the dashboard — it adds no value.
- As a user, I expect all detail/list pages to have the same back-button and title pattern.

## 4. Functional Requirements

### 4.1 Remove Meal Type Badge from Meal Planner Grid (Item 1)

- Remove the `recipe_classification` badge rendered on recipe cards within the week grid.
- File: `frontend/src/components/features/meals/week-grid.tsx` (lines 113-116).
- The grid columns already indicate Breakfast/Lunch/Dinner/Snack, making the badge redundant.

### 4.2 Fix Cooklang Time Parser for Compound Durations (Item 2)

- Update `ParseMinutes()` in `services/recipe-service/internal/recipe/cooklang/parser.go` to handle compound durations.
- Must correctly parse formats like `"1h 20m"`, `"1h20m"`, `"1 hour 20 minutes"`, `"2h"`, `"45m"`, `"90 minutes"`, `"1.5 hours"`.
- `"1h 20m"` must return `80`, not `1`.
- Add test cases for all compound duration formats.
- The frontend total-time calculation (`prepTimeMinutes + cookTimeMinutes`) is correct and requires no changes.

### 4.3 Add Pagination and Total Count to Ingredients Page (Item 3)

- The backend already supports `page[number]` and `page[size]` query params with `meta.total` in the response.
- Add pagination controls to `frontend/src/pages/IngredientsPage.tsx`.
- Display total ingredient count (e.g., "Showing 1-20 of 142 ingredients").
- Pass `page` and `pageSize` params from the `useIngredients` hook.
- Default page size: 20 (matches backend default).

### 4.4 Calendar Popover for Meal Planner Week Selector (Item 4)

- In `frontend/src/components/features/meals/week-selector.tsx`, make the date range label a popover trigger.
- Clicking it opens a calendar (shadcn/ui `Calendar` component inside a `Popover`).
- Selecting any date navigates to the week containing that date.
- Existing prev/next week buttons remain unchanged.

### 4.5 Calendar Popover for Calendar Page "Today" Widget (Item 5)

- In `frontend/src/pages/CalendarPage.tsx`, make the "Today" button a popover trigger.
- Clicking it opens a calendar; selecting a date navigates to the week containing that date.
- Existing prev/next buttons remain unchanged.

### 4.6 Fix Status Filter Dropdown Label (Item 6)

- In `frontend/src/components/common/list-filter-bar.tsx`, when the "all" value is selected, the closed trigger should display "All statuses" (not the raw value "all").
- The `SelectTrigger` currently renders the value directly. Fix so the display label matches the `SelectItem` label.

### 4.7 Title-Case Recipe Tags (Item 7)

- Visually title-case recipe tag text (capitalize the first letter of each word).
- Apply in:
  - `frontend/src/components/features/recipes/recipe-card.tsx` (tag badges)
  - `frontend/src/pages/RecipeDetailPage.tsx` (tag badges)
  - `frontend/src/pages/RecipesPage.tsx` (tag filter badges)
- This is display-only — do not modify stored tag values.
- Use a shared utility function (e.g., `toTitleCase`) for consistency.

### 4.8 Remove Dashboard "You are the owner" Subtitle (Item 8)

- Remove the `resolvedRole` subtitle from `frontend/src/pages/DashboardPage.tsx` (lines 52-54).
- Keep the "Dashboard" heading.

### 4.9 Align Household Members Page Header (Item 9)

- Update `frontend/src/pages/HouseholdMembersPage.tsx` header to match the pattern used by `RecipeDetailPage` and `IngredientDetailPage`.
- Replace the current `<Link>`-wrapped icon-only back button with a `<Button variant="ghost" size="sm">` containing an arrow icon and "Households" text label.
- Use `useNavigate()` + `onClick` instead of `<Link>` to match other detail pages.

## 5. API Surface

No new endpoints. The existing `GET /ingredients` endpoint with `page[number]`, `page[size]` query params and `meta.total` response field is already sufficient.

## 6. Data Model

No data model changes required.

## 7. Service Impact

| Service | Impact |
|---------|--------|
| **frontend** | Items 1, 3, 4, 5, 6, 7, 8, 9 — UI changes across multiple pages and components |
| **recipe-service** | Item 2 — Fix `ParseMinutes` in `cooklang/parser.go`, add tests |

## 8. Non-Functional Requirements

- All changes are backward-compatible; no migrations needed.
- Calendar popover should be responsive on mobile (full-width on small screens).
- Pagination controls should handle edge cases (first/last page disabled states).
- Title-case utility must handle edge cases (hyphens, single-letter words).

## 9. Open Questions

None — all questions resolved during spec review.

## 10. Acceptance Criteria

- [ ] Meal planner grid cards no longer show meal-type classification badges
- [ ] `ParseMinutes("1h 20m")` returns `80`; all compound duration formats tested
- [ ] Ingredients page shows "Showing X-Y of Z ingredients" and pagination controls
- [ ] Clicking the week date range on Meal Planner opens a calendar popover; selecting a date navigates to that week
- [ ] Clicking "Today" on Calendar page opens a calendar popover; selecting a date navigates to that week
- [ ] Status filter dropdown shows "All statuses" when closed with "all" selected
- [ ] Recipe tags display with title-cased text on Recipes page, Recipe Detail page, and recipe cards
- [ ] Dashboard no longer shows "You are the owner/member" subtitle
- [ ] Household Members page uses the same back-button + text label pattern as Recipe/Ingredient detail pages
