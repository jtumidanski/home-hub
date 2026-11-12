# task-058 Context: URL-Addressable Views

This page is a quick reference for implementers. `plan.md` is authoritative. `design.md` holds the rationale.

## What this changes

- **View in the path:** Habits (`/app/habits/today|calendar|setup|report`), Ingredients (`/app/ingredients/list|categories|bulk`), Grocery (`/app/shopping/grocery` plus `/archived`), and Calendar (`/app/calendar/week|month`).
- **Period in a query param:** Habits `?month=YYYY-MM`, Calendar `?date=YYYY-MM-DD`, Meals `?week=YYYY-MM-DD`. The param is omitted for the current period.
- Frontend only.

## Key files

| File | Role | Notes |
|---|---|---|
| `frontend/src/App.tsx` | Providers + `BrowserRouter` | `TenantProvider` wraps `<Routes>` here, above the `/app` layout route. The route tree moves out in Task 3. |
| `frontend/src/app-routes.tsx` (new) | `AppRoutes`: the full `<Routes>` tree | Tested directly in `src/__tests__/app-routes.test.tsx`. |
| `frontend/src/components/common/redirect-preserving-query.tsx` (new) | `<Navigate replace>` that keeps `search` and `hash` | Required for OAuth `?connected` / `?error` and `?category=`. |
| `frontend/src/lib/hooks/use-query-param-state.ts` (new) | `QueryParamCodec<T>`, `useQueryParamState`, `withParam` | Read in render, write with replace, omit when current, canonicalize once. |
| `frontend/src/lib/date-utils.ts` | `+ parseLocalDateStr`, `+ formatLocalDateStr` | `formatLocalDateStr` replaces MealsPage's private `formatDateStr`. |
| `frontend/src/components/features/calendar/calendar-url.ts` (new) | `weekAnchorFor`, `calendarDateCodec` | Holds the mobile/desktop week-anchor math formerly in `focusDay`. |
| `frontend/src/test/location-probe.tsx` (new) | Test helper | Renders `pathname+search`, `useNavigationType()`, and a Back button. |
| `frontend/src/pages/TrackerPage.tsx` | Habits | 63 lines. `view` and `month` were `useState`. |
| `frontend/src/pages/IngredientsPage.tsx` | Ingredients | `view` `useState` at line 34. Toggle buttons at 119–135. `?category=` is already handled via `useSearchParams`. |
| `frontend/src/pages/ShoppingListsPage.tsx` | Grocery | `tab` `useState` at line 31. Plain `<button>`s at 102–123 have no `aria-pressed` yet. |
| `frontend/src/pages/CalendarPage.tsx` | Calendar | `weekStart`/`viewMode`/`monthAnchor` `useState` at 51–53. `isDesktop` reset effect at 70–81. OAuth effect at 83–100 currently wipes **all** params. `focusDay`/`goPrev`/`goNext`/`handleViewModeChange` at 152–201. |
| `frontend/src/pages/MealsPage.tsx` | Meals | `startsOn` `useState(getLocalWeekStart)` at line 57. `WeekSelector onWeekChange`. |
| `frontend/src/components/features/navigation/nav-config.ts` | `isNavItemActive(pathname, to, end)` | Prefix match unless `end`. The in-scope items don't set `end`, so no change is needed (guard test added). |

## Existing helpers to reuse

- `getStartOfWeek` (Sunday), `getStartOfMonth`, `addMonths`, and `isSameMonth` are in `components/features/calendar/calendar-utils.ts`.
- `getLocalWeekStart` (Monday, returns `Date`) and `getLocalMonth` (`YYYY-MM`) are in `lib/date-utils.ts`.
- `isoMondayOf(date): string` is in `lib/workout-week.ts`.
- Calendar mobile: `dayCount` is 3 below 768px. The window starts at "yesterday" and steps by 3 days.

## Decisions (from design.md §7)

1. Static child routes plus a typed `view`/`tab` prop. No `:view` param and no outlet shell. Do not use `key` props.
2. Toggles stay `<Button>`s with `aria-pressed`. Grocery tabs gain `aria-pressed`.
3. Nav and widget links stay on bare URLs and rely on redirects. Only `/app/tracker` is retargeted, to `/app/habits/today`.
4. A Habits view switch carries the whole query string.
5. Mobile Calendar `date` is the first displayed day and is **not** normalized to Sunday. This is the one PRD deviation; mention it in the PR.
6. Calendar OAuth cleanup deletes only `connected` and `error`.
7. Ingredients has no `ingredients/*` catch-all. Unknown segments fall through to `:id`.

## Testing notes

- Vitest config is in `frontend/vite.config.ts`. The setup file is `src/test/setup.ts`. Tests match `src/**/*.test.{ts,tsx}`.
- Existing tests to update: `pages/__tests__/TrackerPage.test.tsx` (renders `<TrackerPage />` with no router, so it must be rewritten) and `pages/__tests__/ShoppingListsPage.test.tsx` (pass `tab`).
- No `nav-config` test exists yet. Task 3 creates `components/features/navigation/__tests__/nav-config.test.ts`.
- Pin dates with `vi.useFakeTimers({ toFake: ["Date"] })` plus `vi.setSystemTime(new Date(2026, 10, 12))`. Faking only `Date` keeps RTL's async utilities working.
- Calendar uses `window.matchMedia` at module load. Stub it to desktop before importing `CalendarPage`.

## Gate

`tools/verify.sh` with no flags. Its frontend leg runs `npm ci && npm run build && npm test` and `npx eslint .`. No `shared/` Go change is made, but the full gate still has to exit 0.
