# URL-Addressable Views — Design

Task: task-058-url-addressable-views
PRD: [`prd.md`](prd.md) (approved)
Created: 2026-11-12
Scope: frontend only (`frontend/src`). No backend, API, or data-model change.

---

## 1. Intent

Make the active view on Habits, Ingredients, Grocery Lists, and Calendar a
URL path segment, and the viewed period on Habits, Calendar, and Meal Planner
a query parameter, so refresh, bookmarks, shared links, and Back/Forward all
land the user where they were. Follow the convention Workouts and Shopping
already use (`/app/workouts/today|week|…`, `/app/shopping/grocery|wish-list`).

Anti-goals: a flash of the default view on refresh; history flooded by
prev/next date clicks; any existing inbound link (sidebar, widgets, OAuth
return, legacy `/app/tracker`) breaking; the layout shell or
`TenantProvider` remounting on a view switch.

## 2. Current state (verified against source)

| Page | View state today | Date state today |
|---|---|---|
| `pages/TrackerPage.tsx` | `useState<View>("today")`, views `today/calendar/setup/report` | `useState(getLocalMonth)` (`YYYY-MM` string) |
| `pages/IngredientsPage.tsx` | `useState<"list"\|"categories"\|"bulk">("list")`; Categories/Bulk buttons toggle back to list when pressed again; "New" button forces `list` | — (`?category=` already in URL via `useSearchParams`) |
| `pages/ShoppingListsPage.tsx` | `useState<"active"\|"archived">("active")`, plain `<button>`s, no `aria-pressed` | — |
| `pages/CalendarPage.tsx` | `useState<CalendarViewMode>("week")`, `ViewModeToggle` component | `weekStart` (Sunday via `getStartOfWeek` on desktop; **yesterday** on mobile, stepping by `dayCount` = 3) and `monthAnchor` (`getStartOfMonth`); an `isDesktop` effect resets `weekStart` to "now" on viewport change |
| `pages/MealsPage.tsx` | single view | `useState<Date>(getLocalWeekStart)` (Monday), `WeekSelector` calls `onWeekChange` |

Routes (`App.tsx`, all under the `/app` layout route that renders
`ProtectedRoute` → `TenantProvider`/`AppShell`): flat routes `habits`,
`tracker` → `Navigate("/app/habits")`, `calendar`, `meals`, `ingredients`,
`ingredients/:id`, `shopping/grocery`, `shopping/grocery/:id`.

`CalendarPage`'s OAuth effect reads `?connected=true` / `?error=…`, toasts, then
calls `setSearchParams({}, { replace: true })` — it wipes **all** params.

Nav: `isNavItemActive(pathname, to, end)` in `nav-config.ts` matches `to` or
`to/…` unless `end`. None of the five in-scope nav items set `end`, so
FR-G9 already holds for every new sub-path (`/app/habits/setup`,
`/app/shopping/grocery/archived`, …). **No nav change needed.**

## 3. Decisions

### D1. Routing shape: static child routes + a `view` prop

Each view is its own static `<Route>` rendering the same page component with a
`view` prop. Bare and unknown URLs are handled by small redirect routes.

```tsx
<Route path="habits" element={<RedirectPreservingQuery to="/app/habits/today" />} />
<Route path="habits/today"    element={<TrackerPage view="today" />} />
<Route path="habits/calendar" element={<TrackerPage view="calendar" />} />
<Route path="habits/setup"    element={<TrackerPage view="setup" />} />
<Route path="habits/report"   element={<TrackerPage view="report" />} />
<Route path="habits/*"        element={<RedirectPreservingQuery to="/app/habits/today" />} />
<Route path="tracker"         element={<RedirectPreservingQuery to="/app/habits/today" />} />

<Route path="calendar"        element={<RedirectPreservingQuery to="/app/calendar/week" />} />
<Route path="calendar/week"   element={<CalendarPage view="week" />} />
<Route path="calendar/month"  element={<CalendarPage view="month" />} />
<Route path="calendar/*"      element={<RedirectPreservingQuery to="/app/calendar/week" />} />

<Route path="ingredients"            element={<RedirectPreservingQuery to="/app/ingredients/list" />} />
<Route path="ingredients/list"       element={<IngredientsPage view="list" />} />
<Route path="ingredients/categories" element={<IngredientsPage view="categories" />} />
<Route path="ingredients/bulk"       element={<IngredientsPage view="bulk" />} />
<Route path="ingredients/:id"        element={<IngredientDetailPage />} />   {/* unchanged */}

<Route path="shopping/grocery"          element={<ShoppingListsPage tab="active" />} />
<Route path="shopping/grocery/archived" element={<ShoppingListsPage tab="archived" />} />
<Route path="shopping/grocery/:id"      element={<ShoppingListDetailPage />} />  {/* unchanged */}

<Route path="meals" element={<MealsPage />} />   {/* unchanged; date only */}
```

Alternatives considered:

- **`habits/:view` param route, page validates the param.** Fewer route lines,
  but puts string validation and an in-page `<Navigate>` inside every page,
  and cannot be used for Ingredients/Grocery where `:id` already owns the
  dynamic slot. Rejected for inconsistency.
- **Layout route + `<Outlet>` like `WorkoutShell`.** Workouts uses this because
  each view is a separate page component. Here the views share header, date
  state, and dialogs inside one page component; splitting them into an outlet
  shell is a larger refactor than the PRD asks for. Rejected.

Why this choice: the route table is the single, greppable list of valid URLs;
pages receive a typed `view` and never parse paths; React Router v7 ranks
static segments above `:id`, so `list|categories|bulk|archived` win without
tricks (FR-I1, FR-S2 — proven by tests, §6); `habits/*` and `calendar/*`
give FR-G3 redirects. Ingredients deliberately has **no** `ingredients/*`
catch-all: an unknown segment falls to `ingredients/:id` and the detail page's
existing not-found handling (FR-I1).

`view` is a required prop typed as the page's view union (exported from the
page file, e.g. `export type HabitsView = "today" | "calendar" | "setup" |
"report"`). The page holds no `useState` copy of it (FR-G1).

**Mount stability.** Switching between sibling routes that render the same
component type at the same position does not remount the `/app` layout
element, so `TenantProvider`/`AppShell` are untouched (NFR Performance). The
page component itself may or may not be reused by React reconciliation; data
comes from React Query caches either way, so no extra fetch occurs beyond
what today's staleness rules already allow. The builder must not add `key`
props to these route elements. A remount of the page *would* drop transient
local state such as Ingredients search text; if the test in §6 shows the page
remounts on view switch, that is acceptable (search is a non-goal, FR-I4) but
should be noted in the PR.

### D2. `RedirectPreservingQuery`

New component in `frontend/src/components/common/redirect-preserving-query.tsx`:

```tsx
export function RedirectPreservingQuery({ to }: { to: string }) {
  const { search, hash } = useLocation();
  return <Navigate to={{ pathname: to, search, hash }} replace />;
}
```

`<Navigate to="today">` drops `?…`, which would break the OAuth toast
(FR-C1) and `?category=` (FR-I2); this component is the single fix for FR-G2.
`<Navigate>` renders before paint of the target, so there is no flash of the
default view (NFR No flash): the default view is never rendered for a
non-default URL because the view comes from the route, not from state that
later updates.

The existing `/app/shopping` → `/app/shopping/grocery` redirect and
`RedirectSummaryToReview` stay as they are.

### D3. View controls stay buttons that call `navigate` (push)

PRD Open Question 2. Decision: keep every toggle a `<Button>`/`<button>` with
`aria-pressed` and call `navigate(target)` (push, FR-G4).

- `aria-pressed` toggle-button semantics are what the controls are today and
  what tests assert; converting to `<NavLink>` would change them to link
  semantics (`aria-current`) and require restyling shadcn `Button` via
  `asChild` on four pages. FR-G10 is satisfied trivially by not changing the
  element.
- Middle-click-to-new-tab is not a PRD goal.

Grocery's Active/History tab buttons get `aria-pressed` added (one attribute
each) so tests can assert the selected tab by role, matching the other pages.
Visual design unchanged.

### D4. Inbound links rely on redirects

PRD Open Question 1. Decision: **leave nav config and dashboard widget links
pointing at the bare page URLs** (`/app/habits`, `/app/calendar`,
`/app/ingredients`, `/app/meals`, `/app/shopping/grocery`). The redirects are
required anyway (bookmarks, OAuth return URL, `/app/tracker`), the bare URL
is the stable public contract, and it keeps the diff out of ~8 widget files.
Only `/app/tracker`'s target changes, to `/app/habits/today`, saving a hop.

In-page navigations that are new or changed by this task use explicit
targets (`/app/habits/report?month=…`, `/app/calendar/week?date=…`).

### D5. Shared hook for date query params

New hook `frontend/src/lib/hooks/use-query-param-state.ts`:

```ts
interface QueryParamCodec<T> {
  key: string;
  parse: (raw: string) => T | null;   // null = invalid
  serialize: (value: T) => string;    // canonical form
  current: () => T;                   // "current period" value, evaluated per render
  equals: (a: T, b: T) => boolean;
}

function useQueryParamState<T>(codec: QueryParamCodec<T>): [T, (next: T) => void]
```

Behaviour:

- **Read:** `raw = searchParams.get(key)`; `value = raw != null ? parse(raw) ?? current() : current()`.
  Computed during render, so fallback is immediate (no flash, FR-G7).
- **Write:** `setValue(next)` calls
  `setSearchParams(prev => { const p = new URLSearchParams(prev); equals(next, current()) ? p.delete(key) : p.set(key, serialize(next)); return p; }, { replace: true })`.
  Functional update preserves unrelated params (`category`, OAuth params).
  Replace satisfies FR-G5; delete-when-current satisfies FR-G6.
- **Canonicalize:** one `useEffect` keyed on `raw`: if `raw != null` and
  (`parse(raw) == null` **or** `serialize(value) !== raw` **or**
  `equals(value, current())`), rewrite the param via the same write path
  (`replace`). This covers invalid input (FR-G7), non-canonical input (e.g. a
  mid-week date for Calendar week view, FR-C2 "normalized"), and a hand-typed
  current-period param (FR-G6), with one rule.

Codecs that need view/device context (Calendar) are built with `useMemo` per
render; the hook must not assume the codec is referentially stable — it only
depends on `raw` and the codec's pure functions.

Pages also need to build URLs for **push** navigations that carry a date
(view switches). Export a pure helper alongside the hook:

```ts
function withParam<T>(search: URLSearchParams, codec: QueryParamCodec<T>, value: T): string
// returns "?…" with key set/deleted by the same omit-when-current rule
```

Date parsing helper added to `lib/date-utils.ts`:
`parseLocalDateStr(s: string): Date | null` — accepts only `^\d{4}-\d{2}-\d{2}$`,
builds `new Date(y, m-1, d)` (local, no UTC — NFR Timezone), and rejects
non-round-tripping dates (`2026-02-31`). `formatLocalDateStr(d: Date): string`
also moves here (MealsPage's private `formatDateStr` becomes this), so all
three pages share one serializer.

Alternative considered: three ad-hoc `useSearchParams` blocks. Rejected — the
omit/replace/canonicalize rules are the subtle part and would be repeated
three times with three slightly different bugs.

### D6. Per-page behaviour

#### Habits (`TrackerPage`)

- Props: `{ view: HabitsView }`. Remove both `useState`s.
- Codec `month`: `parse` accepts `^\d{4}-(0[1-9]|1[0-2])$`; `serialize`
  identity; `current = getLocalMonth`. Used on all views but only consumed by
  `calendar` and `report` (FR-H1). It is read unconditionally (hooks rule),
  which means a stray invalid `month` on `today`/`setup` is also cleaned —
  harmless and simpler than gating.
- Header toggle (Today/Calendar/Setup): `navigate({ pathname: \`/app/habits/${id}\`, search: location.search })`.
  Carrying the whole search is my call: it carries `month` to calendar
  (FR-H4) and keeps a month across a detour through Setup; FR-H1 does not
  require stripping it. On `report` no button is pressed (unchanged).
- `CalendarGrid` props unchanged: `onMonthChange = setMonth` (replace,
  FR-H2); `onViewReport` → push `/app/habits/report` with current search;
  `onOpenSetup` → push `/app/habits/setup` (FR-H3).
- `widthClass` derived from `view` exactly as today (FR-H5).

#### Ingredients (`IngredientsPage`)

- Props: `{ view: IngredientsView }`. Remove view `useState`.
- Categories/Bulk buttons keep their toggle semantics: pressing the active
  one navigates back to `list`. Target is
  `/app/ingredients/<view>` with **current search preserved** (so `?category=`
  survives a round trip; it's ignored off `list`, FR-I2).
- "New ingredient" button: currently `setView("list")` + open create form →
  navigate to `list` (push only if not already on `list`) and open the form.
- Search text and `page` remain `useState` (FR-I4).
- Detail page's back link to `/app/ingredients` keeps working via redirect
  (FR-I3); not changed.

#### Grocery Lists (`ShoppingListsPage`)

- Props: `{ tab: "active" | "archived" }`. Remove `useState`.
- Buttons push `/app/shopping/grocery` or `/app/shopping/grocery/archived`
  (FR-S3); add `aria-pressed` (D3).
- `ShoppingListDetailPage` navigations untouched (FR-S1).

#### Calendar (`CalendarPage`)

- Props: `{ view: CalendarViewMode }`. Remove `viewMode`, `weekStart`,
  `monthAnchor` `useState`s and the `isDesktop` reset effect.
- One query param `date`, codec chosen by view and viewport:

  | view | `current()` | normalize on parse |
  |---|---|---|
  | `month` | `getStartOfMonth(today)` | `getStartOfMonth(d)` |
  | `week`, desktop (`dayCount` 7) | `getStartOfWeek(today)` (Sunday) | `getStartOfWeek(d)` |
  | `week`, mobile (`dayCount` 3) | yesterday at 00:00 (today's behaviour) | none — first displayed day |

  `monthAnchor = view === "month" ? value : getStartOfMonth(value)`;
  `weekStart = view === "week" ? value : …` — the page simply uses `value` as
  the anchor for whichever view is active.

  **Mobile deviation (my call, flagged for review).** PRD FR-C2 says `date`
  on `week` is normalized to `getStartOfWeek`, and FR-C6 says mobile
  behaviour is unchanged. Those conflict: mobile shows a 3-day window
  starting yesterday and steps by 3 days, so its window start is usually not
  a Sunday; normalizing would snap every prev/next back to Sunday and break
  mobile navigation. Resolution: on mobile, `date` is the first displayed day,
  stored verbatim. Desktop behaviour is exactly FR-C2. Crossing the
  breakpoint needs no effect: desktop re-derives `getStartOfWeek(date)`;
  mobile shows the 3 days starting at `date`. The canonicalize effect will
  rewrite a mobile mid-week date to its Sunday when the viewport becomes
  desktop — acceptable.

- `goPrev`/`goNext`/date picker/`focusDay` compute the new anchor exactly as
  today and call `setDate(next)` (replace, FR-C3).
- `handleViewModeChange(mode)` keeps today's transition math (FR-C4) but
  pushes: week → month `navigate({ pathname: "/app/calendar/month", search: withParam(…, monthCodec, getStartOfMonth(weekStart)) })`;
  month → week computes `focusDay(isSameMonth(today, monthAnchor) ? today : monthAnchor)`
  as a value (refactor `focusDay` into a pure `weekAnchorFor(day, isDesktop)`
  returning the Date) and pushes `/app/calendar/week` with that date.
- `handleDayClick(day)` pushes `/app/calendar/week?date=<weekAnchorFor(day)>`
  (FR-C5); Back returns to month view.
- OAuth effect: replace `setSearchParams({}, …)` with a functional update
  that deletes only `connected` and `error` (replace). Prevents the toast
  cleanup from wiping `date` and keeps FR-C1 working after the
  `RedirectPreservingQuery` hop.
- `ViewModeToggle` component unchanged; it receives `mode={view}`.

#### Meal Planner (`MealsPage`)

- Codec `week`: `parse` via `parseLocalDateStr` then normalize to Monday
  (reuse `isoMondayOf` from `lib/workout-week.ts`, or a Date-returning
  sibling added next to it — builder's choice); `current = getLocalWeekStart`;
  `serialize = formatLocalDateStr`; `equals` by serialized string.
- `startsOn` comes from the hook; `WeekSelector onWeekChange={setWeek}`
  (replace, FR-M1). `startsOnStr` and all queries unchanged.
- Widgets linking `/app/meals` carry no param → current week (FR-M2).

## 4. Files touched

| File | Change |
|---|---|
| `frontend/src/App.tsx` | Route table per D1; `/app/tracker` target; import `AppRoutes` (below). |
| `frontend/src/app-routes.tsx` (new) | The `<Routes>` tree extracted from `App.tsx` unchanged except for D1, so tests can render the real table in a `MemoryRouter`. `App.tsx` keeps providers + `BrowserRouter`. |
| `frontend/src/components/common/redirect-preserving-query.tsx` (new) | D2. |
| `frontend/src/lib/hooks/use-query-param-state.ts` (new) | D5 hook + `withParam`. |
| `frontend/src/lib/date-utils.ts` | `parseLocalDateStr`, `formatLocalDateStr`. |
| `frontend/src/pages/TrackerPage.tsx` | D6 Habits. |
| `frontend/src/pages/IngredientsPage.tsx` | D6 Ingredients. |
| `frontend/src/pages/ShoppingListsPage.tsx` | D6 Grocery. |
| `frontend/src/pages/CalendarPage.tsx` | D6 Calendar. |
| `frontend/src/pages/MealsPage.tsx` | D6 Meals. |
| Nav, widgets, `ShoppingListDetailPage`, `reauthorize-banner.tsx`, `connection-status.tsx` | **No change** (D4, FR-C1). |

`RedirectSummaryToReview` moves with the route tree into `app-routes.tsx`.

## 5. Risks

- **Mobile calendar `date` semantics** differ from desktop (D6). Flagged above;
  it's the only place this design departs from PRD wording.
- **Canonicalize effect loops.** If `serialize(parse(raw))` is not a fixed
  point, the effect would rewrite forever. Codec unit tests must assert
  `serialize(parse(serialize(x))) === serialize(x)` and that the effect runs
  at most once per bad URL.
- **"Current period" at midnight.** `current()` is evaluated per render, so a
  page left open over midnight treats the old period as non-current on the
  next navigation and writes the param. Acceptable; same as Workouts.
- **Page remount on sibling-route switch** — see D1 note; test it.

## 6. Test plan

Vitest + Testing Library, `MemoryRouter`.

1. **`lib/hooks/__tests__/use-query-param-state.test.tsx`** — read default when
   absent; parse valid; invalid → current + param removed (replace); current
   value written → param deleted; non-current → set; other params preserved;
   writes use replace (history length unchanged via a location-recording
   probe); idempotent canonicalization.
2. **`lib/__tests__/date-utils.test.ts`** — `parseLocalDateStr` rejects
   `2026-02-31`, `2026-1-5`, garbage; round-trips with `formatLocalDateStr`.
3. **`__tests__/app-routes.test.tsx`** — render `AppRoutes` with page modules
   and `ProtectedRoute`/`AppShell` mocked to stubs that print their props /
   render `<Outlet/>`. Cases:
   - each `/app/habits/{today,calendar,setup,report}` renders `TrackerPage`
     with that view; `/app/habits`, `/app/habits/bogus`, `/app/tracker` land
     on `/app/habits/today`; `/app/habits?month=2026-01` keeps the query.
   - `/app/calendar/{week,month}`; `/app/calendar` and `/app/calendar/bogus`
     → `/app/calendar/week`; `/app/calendar?connected=true` and
     `?error=auth_failed` keep their query.
   - `/app/ingredients/{list,categories,bulk}` render `IngredientsPage`;
     `/app/ingredients/<uuid>` renders `IngredientDetailPage`;
     `/app/ingredients?category=x` → `/app/ingredients/list?category=x`.
   - `/app/shopping/grocery` → active, `/archived` → archived,
     `/app/shopping/grocery/<id>` → detail.
   - layout stub mount counter does not increment on a view switch.
4. **Page tests** (update `TrackerPage.test.tsx`, `ShoppingListsPage.test.tsx`;
   add `IngredientsPage`, `CalendarPage`, `MealsPage` route-behaviour tests as
   needed): button `aria-pressed` reflects `view` prop; clicking pushes the
   right path (assert via a location probe and that `navigate(-1)` returns);
   Habits "View report" carries `month`; Calendar prev/next sets/removes
   `date` without pushing; week↔month carries date per FR-C4; day click →
   week for that day; OAuth toast fires and only `connected`/`error` are
   removed; Meals `?week=` restores week, current week has no param, a
   mid-week date normalizes to Monday.
5. **Nav:** one `isNavItemActive` case each for `/app/habits/setup` and
   `/app/shopping/grocery/archived` in the existing nav-config test (FR-G9
   regression guard).

Done = flagless `tools/verify.sh` exits 0.

## 7. Calls made by the designer (for review)

1. Static child routes + `view` prop rather than `:view` param or outlet shells (D1).
2. Toggles stay buttons with `aria-pressed`; Grocery tabs gain `aria-pressed` (D3).
3. Inbound links stay on bare URLs and rely on redirects; only `/app/tracker` retargeted (D4).
4. Habits view switches carry the whole query string (D6).
5. Mobile Calendar `date` = first displayed day, not normalized to Sunday (D6) — the one PRD deviation.
6. Calendar OAuth cleanup deletes only `connected`/`error` (D6).
7. `<Routes>` tree extracted to `app-routes.tsx` so route precedence and redirects are tested against the real table (§4, §6).

## 8. Left to the builder

- Exact test file split and naming for page tests.
- Whether Meals Monday normalization reuses `isoMondayOf` or adds a
  Date-returning sibling in `lib/workout-week.ts` / `date-utils.ts`.
- Whether `withParam` lives in the hook file or `lib/`.
- Follow `frontend-dev-guidelines` throughout.
