# URL-Addressable Views — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-11-12
---

## 1. Overview

Several Home Hub pages have child views or tabs: Habits has Today / Calendar /
Setup (and a Report view reached from Calendar), Ingredients has List /
Categories / Bulk, Grocery Lists has Active / Archived, and Calendar has Week /
Month. Today these pages keep the active view in React `useState`, so a
browser refresh, a bookmark, or a shared link always lands on the default view.
The date being viewed (the Habits month, the Calendar week/month, the Meal
Planner week) is also lost on refresh.

Workouts already solves this with route path segments
(`/app/workouts/today|week|exercises|taxonomy|review`) and Shopping uses
`/app/shopping/grocery` vs `/app/shopping/wish-list`. This task extends that
convention to the remaining pages: the active view becomes a path segment, and
the viewed date becomes a query parameter. A refresh returns the user to exactly
what they were looking at, and browser Back/Forward moves between views.

This is a frontend-only change. No backend service, API, or data model is
affected.

## 2. Goals

Primary goals:
- Every in-scope view/tab is addressable by URL path segment, matching the
  Workouts convention.
- The viewed date on Habits, Calendar, and Meal Planner survives a refresh.
- Switching views pushes a browser history entry so Back returns to the
  previous view; date navigation replaces the entry so history is not flooded.
- All existing inbound links (sidebar nav, dashboard widgets, OAuth redirect,
  in-app `navigate()` calls, the legacy `/app/tracker` redirect) keep working.

Non-goals:
- Persisting dialog, popover, sheet, or form state (create/edit/delete dialogs,
  calendar source picker, etc.) in the URL.
- Persisting Ingredients free-text search or pagination page in the URL
  (category filter is already persisted and must keep working).
- Cook mode's internal view mode (`cook-mode.tsx`), Workout Week filters
  (`themeId`/`regionId`), Workout Exercises form fields, Settings avatar style,
  Onboarding steps.
- Remembering a "last used view" per user (localStorage or server-side). A bare
  page URL always resolves to the default view.
- Any backend change.

## 3. User Stories

- As a household member, I want a browser refresh on Habits → Setup to keep me
  on Setup so I don't lose my place while configuring habits.
- As a household member viewing the Habits calendar for a past month, I want a
  refresh to keep showing that month.
- As a user on Calendar month view, I want a refresh to keep month view and the
  same month instead of jumping back to this week.
- As a user browsing archived grocery lists, I want a refresh to keep showing
  archived lists.
- As a user bulk-categorizing ingredients, I want a refresh to keep me on Bulk.
- As a user planning next week's meals, I want a refresh to keep showing next
  week.
- As a user, I want to bookmark or share a link to a specific view (e.g.
  `/app/habits/calendar`) and land on it directly.
- As a user, I want the browser Back button to return me to the previous tab I
  was on.

## 4. Functional Requirements

### 4.1 General rules (apply to every in-scope page)

- **FR-G1** The active view is derived from the URL on every render. The page
  holds no `useState` copy of the active view; the URL is the single source of
  truth.
- **FR-G2** A bare page URL (no view segment) redirects with `replace` to the
  default view, preserving any existing query string (`?…`). Exception: Grocery
  Lists, see 4.4.
- **FR-G3** An unrecognized view segment redirects with `replace` to the
  default view (no 404), except where it would collide with an existing
  `/:id` detail route (Ingredients; see 4.3).
- **FR-G4** Clicking a view/tab control navigates with a **push** (new history
  entry). Browser Back returns to the previous view.
- **FR-G5** Date navigation (prev/next, "Today", date picker) updates the date
  query param with **replace** (no new history entry).
- **FR-G6** Date query params are **omitted** when the viewed period is the
  current period (current month / week containing today). A URL without a
  date param always means "current period". Navigating away from the current
  period sets the param; navigating back to the current period removes it.
- **FR-G7** An invalid or unparseable date param falls back to the current
  period and the param is removed with `replace`. No error toast.
- **FR-G8** When switching views on a page whose views share a date (Habits
  calendar ↔ report, Calendar week ↔ month), the date param is carried over
  (transformed as defined per page) so the user stays on the same period.
- **FR-G9** Sidebar nav items stay highlighted as active on every sub-view
  (e.g. "Habits" is active on `/app/habits/setup`; "Grocery Lists" is active
  on `/app/shopping/grocery/archived`). Verify against
  `isNavItemActive` / the `end` flag in `nav-group.tsx` and `nav-config.ts`.
- **FR-G10** View controls keep their current visual design and accessibility
  semantics (`aria-pressed` / `role="group"` etc.). If a control is converted
  to a link, it must still convey the selected state to assistive tech.

### 4.2 Habits (`TrackerPage.tsx`)

Routes:

| URL | View |
|---|---|
| `/app/habits` | redirect → `/app/habits/today` (preserve query) |
| `/app/habits/today` | Today (default) |
| `/app/habits/calendar` | Calendar grid |
| `/app/habits/setup` | Setup |
| `/app/habits/report` | Month report |
| `/app/habits/<other>` | redirect → `/app/habits/today` |
| `/app/tracker` | redirect → `/app/habits/today` (existing legacy redirect, updated target) |

- **FR-H1** Query param `month=YYYY-MM` applies to `calendar` and `report`.
  Omitted = current local month (`getLocalMonth`). Ignored on `today` and
  `setup` (not written, not required to be stripped).
- **FR-H2** Calendar month prev/next (`onMonthChange`) updates `month` with
  replace (FR-G5/G6).
- **FR-H3** The "View report" action in `CalendarGrid` navigates (push) to
  `/app/habits/report?month=<same month>`. The "Open setup" action navigates
  (push) to `/app/habits/setup`.
- **FR-H4** The header view toggle continues to show Today / Calendar / Setup
  only; on `report` none of the three is pressed (current behavior). Clicking
  Calendar from report goes to `/app/habits/calendar` carrying `month`.
- **FR-H5** The width class (`max-w-none` / `max-w-6xl` / `max-w-3xl`) is
  derived from the URL view exactly as today.

### 4.3 Ingredients (`IngredientsPage.tsx`)

| URL | View |
|---|---|
| `/app/ingredients` | redirect → `/app/ingredients/list` (preserve query, incl. `?category=`) |
| `/app/ingredients/list` | List (default) |
| `/app/ingredients/categories` | Categories |
| `/app/ingredients/bulk` | Bulk categorize |
| `/app/ingredients/:id` | Ingredient detail (unchanged) |

- **FR-I1** Static segments `list`, `categories`, `bulk` must take precedence
  over `:id` (React Router v7 ranks static over dynamic segments; add a test
  that proves it). Ingredient IDs are UUIDs, so no real collision exists.
  An unknown non-UUID segment falls through to the detail page's existing
  not-found handling. FR-G3 does not apply here.
- **FR-I2** The existing `?category=` filter keeps working on `list` and is
  preserved across the bare-URL redirect. Switching views does not need to
  carry `category` (it only affects `list`), but must not throw if present.
- **FR-I3** Links from ingredient detail back to the list
  (`/app/ingredients`) keep working via the redirect.
- **FR-I4** Search text and `page` remain local state (non-goal).

### 4.4 Grocery Lists (`ShoppingListsPage.tsx`)

| URL | View |
|---|---|
| `/app/shopping/grocery` | Active lists (default; **no redirect**) |
| `/app/shopping/grocery/archived` | Archived lists |
| `/app/shopping/grocery/:id` | List detail (unchanged) |

- **FR-S1** Active stays at the bare `/app/shopping/grocery` URL because nav
  config, `ShoppingListDetailPage` (4 `navigate` calls), and the `/app/shopping`
  redirect all target it. This is a deliberate exception to FR-G2.
- **FR-S2** Static `archived` must take precedence over `:id` (test required).
- **FR-S3** Active/Archived buttons navigate with push (FR-G4).

### 4.5 Calendar (`CalendarPage.tsx`)

| URL | View |
|---|---|
| `/app/calendar` | redirect → `/app/calendar/week` (**must preserve query**) |
| `/app/calendar/week` | Week view (default) |
| `/app/calendar/month` | Month view |
| `/app/calendar/<other>` | redirect → `/app/calendar/week` |

- **FR-C1** The Google OAuth reauthorize flow returns to
  `window.location.origin + "/app/calendar"` with `?connected=true` or
  `?error=…` (`reauthorize-banner.tsx`, `connection-status.tsx`). The redirect
  must preserve these params so the existing toast-and-clear effect in
  `CalendarPage` still fires. Do not change the backend-registered return URL.
- **FR-C2** Query param `date=YYYY-MM-DD`:
  - On `week`: the start of the displayed week (as computed by
    `getStartOfWeek`). Any date given is normalized to its week start.
  - On `month`: the first day of the displayed month (`getStartOfMonth`). Any
    date given is normalized to its month start.
  - Omitted = current week / current month (FR-G6).
- **FR-C3** Prev/next, "Today", and the date picker update `date` with replace.
- **FR-C4** The view-mode toggle navigates with push and carries the date using
  the existing transition logic in `handleViewModeChange`: week → month uses
  the month containing the week start; month → week uses today if today is in
  the anchored month, otherwise the month's first week.
- **FR-C5** Clicking a day in month view (`handleDayClick`) navigates (push) to
  `/app/calendar/week?date=<week start of that day>`.
- **FR-C6** Mobile day-count behavior (`dayCount` / `useSyncExternalStore`) is
  unchanged; `date` still represents the week start.

### 4.6 Meal Planner (`MealsPage.tsx`) — date only

- **FR-M1** No view segments (single view). Query param `week=YYYY-MM-DD` holds
  the displayed week start (`getLocalWeekStart` semantics). Omitted = current
  week. Prev/next/today update it with replace.
- **FR-M2** Dashboard widgets linking to `/app/meals` keep landing on the
  current week.

### 4.7 Inbound links

All of the following must continue to resolve correctly (via redirects or
updated targets). Updating targets to the explicit default view is optional;
redirects are sufficient:

- `nav-config.ts` entries for Calendar, Meal Planner, Ingredients, Habits,
  Grocery Lists.
- Dashboard widgets: `habits-widget.tsx`, `calendar-widget.tsx`,
  `calendar-tomorrow-adapter.tsx`, `meal-plan-widget.tsx`,
  `meal-plan-today-detail.tsx`.
- `App.tsx` redirects: `/app/tracker`, `/app/shopping`.
- `ShoppingListDetailPage.tsx` navigations to `/app/shopping/grocery`.

## 5. API Surface

None. No backend endpoints are added or changed. Frontend routes are listed
in section 4.

## 6. Data Model

None. No entities, migrations, or persisted preferences.

## 7. Service Impact

| Area | Change |
|---|---|
| `frontend/src/App.tsx` | Add nested/sub routes and bare-URL redirects for habits, ingredients, grocery archived, calendar; update `/app/tracker` redirect. |
| `frontend/src/pages/TrackerPage.tsx` | View from route; `month` from query. |
| `frontend/src/pages/IngredientsPage.tsx` | View from route. |
| `frontend/src/pages/ShoppingListsPage.tsx` | Tab from route. |
| `frontend/src/pages/CalendarPage.tsx` | View mode from route; anchor from `date` query; preserve OAuth query handling. |
| `frontend/src/pages/MealsPage.tsx` | Week from `week` query. |
| Navigation (`nav-group.tsx` / `nav-config.ts`) | Only if needed for FR-G9. |
| Tests (`frontend/src/pages/__tests__/TrackerPage.test.tsx`, `ShoppingListsPage.test.tsx`, new tests) | Update for route-driven views; add coverage per section 10. |
| Backend services | None. |

The implementation may add a small shared helper or hook (e.g. for reading and
writing a date query param with replace and omit-when-current semantics) under
`frontend/src/lib/` if it removes duplication across the three date-bearing
pages. Follow `frontend-dev-guidelines`.

## 8. Non-Functional Requirements

- **Performance:** Switching views must not refetch data that the current
  implementation does not refetch. In particular, toggling Habits views or
  Calendar week/month must not unmount and remount the `TenantProvider` /
  layout shell. Only the page content changes.
- **No flash:** A bare-URL redirect or refresh on a sub-view must render the
  target view directly, without briefly rendering the default view first.
- **Multi-tenancy:** URLs contain no tenant or household identifiers. Switching
  household with a view URL open keeps the view and reloads the data under the
  new tenant as today.
- **Timezone:** Date params are interpreted in the user's local date (the same
  helpers the pages use today: `getLocalMonth`, `getLocalWeekStart`,
  `getStartOfWeek`, `getStartOfMonth`). No UTC conversion of the param.
- **Accessibility:** View controls keep their pressed/selected semantics and
  keyboard operability (FR-G10).

## 9. Open Questions

- Should nav links and dashboard widgets be updated to point at explicit
  default views (`/app/habits/today`), or rely on redirects? This PRD allows
  either. Design should pick one for consistency.
- Should view toggles become `<Link>`/`<NavLink>` elements (middle-click opens
  in a new tab) or stay as buttons that call `navigate`? Either satisfies the
  requirements. Design should decide.

## 10. Acceptance Criteria

- [ ] `/app/habits/{today,calendar,setup,report}` each render the matching
      view directly, and a refresh on each stays on it.
- [ ] `/app/habits` and `/app/tracker` redirect to `/app/habits/today`.
- [ ] On `/app/habits/calendar`, navigating to a previous month sets
      `?month=YYYY-MM`; a refresh shows that month; returning to the current
      month removes the param; month navigation adds no history entries.
- [ ] "View report" from the habits calendar opens `/app/habits/report` for the
      same month.
- [ ] `/app/ingredients/{list,categories,bulk}` each render the matching view;
      `/app/ingredients` redirects to `list` preserving `?category=`;
      `/app/ingredients/<uuid>` still renders ingredient detail.
- [ ] `/app/shopping/grocery/archived` shows archived lists and survives
      refresh; `/app/shopping/grocery/<id>` still renders list detail.
- [ ] `/app/calendar/{week,month}` render the matching mode; `?date=` restores
      the same week/month on refresh; `/app/calendar` redirects to `week`.
- [ ] `/app/calendar?connected=true` and `/app/calendar?error=…` still show
      their toasts after the redirect.
- [ ] Clicking a day in calendar month view lands on week view for that day's
      week, and Back returns to month view.
- [ ] `/app/meals?week=YYYY-MM-DD` shows that week on refresh; current week
      has no param.
- [ ] Switching any view/tab adds a history entry; browser Back returns to the
      previous view.
- [ ] Invalid date params and unknown view segments fall back as specified,
      with no errors.
- [ ] Sidebar nav highlights the parent item on every sub-view.
- [ ] All nav and dashboard-widget links listed in 4.7 still land on the right
      page.
- [ ] Unit tests cover route-to-view mapping, redirects (incl. query
      preservation), static-over-`:id` precedence for ingredients and grocery,
      and date-param parse/omit/fallback for each date-bearing page.
- [ ] Flagless `tools/verify.sh` exits 0.
