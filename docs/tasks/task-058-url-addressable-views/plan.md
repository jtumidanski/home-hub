# URL-Addressable Views Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Put the active view of Habits, Ingredients, Grocery Lists, and Calendar in the URL path, and the viewed period of Habits, Calendar, and Meal Planner in a query param. Refresh, bookmarks, and Back/Forward then land on the same view and period.

**Architecture:** Each view gets a static `<Route>` that renders the same page component with a typed `view`/`tab` prop. Bare and unknown URLs go through `RedirectPreservingQuery`. Date params go through one hook, `useQueryParamState`, driven by a per-page codec. The hook reads during render, writes with `replace`, omits the param for the current period, and rewrites invalid or non-canonical values once. The `<Routes>` tree moves to `app-routes.tsx` so tests can run against the real route table.

**Tech Stack:** React 19, TypeScript, React Router DOM v7 (`^7.13.2`), Vitest 4 + Testing Library, shadcn/ui.

**Spec:** `docs/tasks/task-058-url-addressable-views/design.md` (PRD: `prd.md`). Read both. When this plan is silent, the design's D1–D6 govern.

## Global Constraints

- Frontend only. All changes are under `frontend/src`. No backend, API, or data-model change.
- Follow the `frontend-dev-guidelines` skill.
- The URL is the single source of truth for the view: no `useState` copy of it in any page (FR-G1).
- Clicking a view/tab pushes a history entry (FR-G4). Date navigation uses `replace` (FR-G5).
- Omit the date param for the current period (FR-G6). Invalid params fall back to the current period and the param is removed with `replace`, with no toast (FR-G7).
- Interpret dates as local dates with `new Date(y, m-1, d)`, never UTC parsing (NFR Timezone).
- Do not put `key` props on the new route elements (design D1, mount stability).
- Do not change nav config, dashboard widget links, `ShoppingListDetailPage`, `reauthorize-banner.tsx`, or `connection-status.tsx` (D4, FR-C1).
- View toggles stay buttons with `aria-pressed` (D3). Do not convert them to links.
- Param names, verbatim: Habits `month=YYYY-MM`, Calendar `date=YYYY-MM-DD`, Meals `week=YYYY-MM-DD`.
- Done means flagless `tools/verify.sh` exits 0. Its frontend leg runs `npm ci && npm run build && npm test` and `npx eslint .`.
- Run per-task tests from `frontend/`: `npx vitest run <path>`. Run the type check with `npx tsc -b`.

## Review Focus

1. **OAuth return with a date already in the URL.** `/app/calendar?connected=true` must toast once and leave only the page's own params. `?connected=true&date=…` must keep `date`. Test owned by Task 8.
2. **Empty or malformed params** (`?month=`, `?month=2026-13`, `?week=2026-02-31`). Each must fall back to the current period and drop the param in one `replace`, with no render loop. Tests owned by Tasks 2, 6, and 7.
3. **Non-canonical but valid dates.** Calendar month with `?date=2026-03-17` must normalize to `2026-03-01`. Meals with a Wednesday must normalize to that week's Monday. Each needs exactly one rewrite. Tests owned by Tasks 7 and 8.
4. **`?category=` round trip.** List → Categories → List must keep `?category=x`, and the bare `/app/ingredients?category=x` must redirect with it. Tests owned by Tasks 3 and 5.
5. **Back after a day click.** Month view at a past month → click a day → week view; Back must return to month view at that same past month. Test owned by Task 8.

---

## Shared test helper (used by Tasks 3–8)

Task 3 creates `frontend/src/test/location-probe.tsx`:

```tsx
export function LocationProbe(): React.JSX.Element
// renders <div data-testid="location">{pathname}{search}</div>
//         <div data-testid="nav-type">{useNavigationType()}</div>   // "PUSH" | "REPLACE" | "POP"
//         <button onClick={() => navigate(-1)}>probe-back</button>
```

Page tests render a local route table inside `MemoryRouter`, plus `<LocationProbe/>`. They assert on the `location` and `nav-type` test IDs. "Push" means `nav-type` is `PUSH`. "Replace" means it is `REPLACE`. "Back works" means clicking `probe-back` restores the previous `location`.

---

### Task 1: Local date-string helpers

**Files:**
- Modify: `frontend/src/lib/date-utils.ts` (append)
- Modify: `frontend/src/pages/MealsPage.tsx:35-40` (delete private `formatDateStr` and use `formatLocalDateStr` at its call sites)
- Test: `frontend/src/lib/__tests__/date-utils.test.ts` (append a `describe`)

**Interfaces:**
- Produces: `formatLocalDateStr(d: Date): string` (local `YYYY-MM-DD`) and `parseLocalDateStr(s: string): Date | null` (local midnight, or `null`).

- [ ] **Step 1: Write the failing tests** in a `describe("parseLocalDateStr / formatLocalDateStr")` block:
  - `parseLocalDateStr("2026-03-17")` → date with `getFullYear()===2026`, `getMonth()===2`, `getDate()===17`, `getHours()===0`.
  - Returns `null` for `"2026-02-31"`, `"2026-1-5"`, `"2026-13-01"`, `""`, `"garbage"`, `"2026-03-17T00:00"`.
  - `formatLocalDateStr(new Date(2026, 0, 5))` → `"2026-01-05"`.
  - Round trip: for `"2024-02-29"` and `"2026-11-01"`, `formatLocalDateStr(parseLocalDateStr(s)!) === s`.
- [ ] **Step 2: Run the tests and confirm they fail.** Run `npx vitest run src/lib/__tests__/date-utils.test.ts`. Expected: FAIL, functions not exported.
- [ ] **Step 3: Implement both functions.** `parseLocalDateStr` accepts only `^\d{4}-\d{2}-\d{2}$`, builds `new Date(y, m-1, d)`, and returns `null` unless `formatLocalDateStr(result) === s`. Then replace MealsPage's `formatDateStr` with an import of `formatLocalDateStr`.
- [ ] **Step 4: Run the tests and the type check.** Run `npx vitest run src/lib/__tests__/date-utils.test.ts && npx tsc -b`. Expected: PASS, no type errors.
- [ ] **Step 5: Commit.** Message: `feat(frontend): local date-string parse/format helpers`.

---

### Task 2: `useQueryParamState` hook and `withParam`

**Files:**
- Create: `frontend/src/lib/hooks/use-query-param-state.ts`
- Test: `frontend/src/lib/hooks/__tests__/use-query-param-state.test.tsx`

**Interfaces:**
- Consumes: nothing from earlier tasks. Uses `useSearchParams` from `react-router-dom`.
- Produces (all exported):
  ```ts
  export interface QueryParamCodec<T> {
    key: string;
    parse: (raw: string) => T | null;   // null = invalid; may normalize
    serialize: (value: T) => string;    // canonical form
    current: () => T;                   // "current period", evaluated per call
    equals: (a: T, b: T) => boolean;
  }
  export function useQueryParamState<T>(codec: QueryParamCodec<T>): [T, (next: T) => void];
  export function withParam<T>(search: URLSearchParams, codec: QueryParamCodec<T>, value: T): string; // "" or "?…"
  ```

Behavior is specified in design D5 (Read / Write / Canonicalize). The codec may be rebuilt on every render. The canonicalize effect depends on `raw` only. Codec functions must be pure.

- [ ] **Step 1: Write the failing tests.** Use a test codec `monthCodec` (`key: "m"`, `parse` matches `^\d{4}-(0[1-9]|1[0-2])$`, `serialize` identity, `current: () => "2026-11"`, `equals` `===`). Add a probe component that renders `value` and a button calling `set(x)`. Render both inside `MemoryRouter initialEntries={[url]}` with a `LocationProbe` defined inline in the test file, since Task 3 creates the shared helper later. Tests:
  - `reads current when absent`: `/p` → value `2026-11`, location `/p`.
  - `reads a valid param`: `/p?m=2026-01` → value `2026-01`, location unchanged.
  - `invalid param falls back and is removed with replace`: `/p?m=2026-13&x=1` → value `2026-11`, location `/p?x=1`, nav-type `REPLACE`.
  - `empty param is removed`: `/p?m=` → location `/p`.
  - `current value in URL is removed`: `/p?m=2026-11` → location `/p`.
  - `set non-current writes param with replace and keeps others`: `/p?x=1`, set `2026-02` → location `/p?x=1&m=2026-02`, nav-type `REPLACE`.
  - `set current deletes param`: `/p?m=2026-02`, set `2026-11` → location `/p`.
  - `normalizing codec rewrites once`: use a codec whose `parse` maps `"2026-03-17"` to `"2026-03-01"`. Count `serialize` calls, or count location changes with a render spy. Assert that `/p?m=2026-03-17` settles at `/p?m=2026-03-01` and stays there after `await` of two more ticks.
  - `withParam`: `withParam(new URLSearchParams("x=1"), monthCodec, "2026-02")` → `"?x=1&m=2026-02"`. `withParam(new URLSearchParams("m=2026-02"), monthCodec, "2026-11")` → `""`.
- [ ] **Step 2: Run the tests and confirm they fail.** Run `npx vitest run src/lib/hooks/__tests__/use-query-param-state.test.tsx`. Expected: FAIL, module not found.
- [ ] **Step 3: Implement the hook and `withParam`** per design D5. Write through `setSearchParams(prev => …, { replace: true })`. The canonicalize `useEffect([raw])` rewrites when `raw != null` and any of these hold: `parse(raw) == null`, `serialize(value) !== raw`, or `equals(value, current())`.
- [ ] **Step 4: Run the tests and confirm they pass.** Same command. Expected: PASS.
- [ ] **Step 5: Commit.** Message: `feat(frontend): useQueryParamState hook for URL-backed date params`.

---

### Task 3: Route-table extraction, `RedirectPreservingQuery`, test harness, nav guard

This task changes no behavior. It moves the route table and adds the redirect component and test scaffolding. Pages are rewired in Tasks 4–8.

**Files:**
- Create: `frontend/src/app-routes.tsx`. Move the entire `<Routes>…</Routes>` tree and `RedirectSummaryToReview` out of `App.tsx` and into `export function AppRoutes(): React.JSX.Element`. Move the page imports with them.
- Modify: `frontend/src/App.tsx`. It keeps the providers, `BrowserRouter`, and `ThemedToaster`, and renders `<AppRoutes />` where `<Routes>` was. Change the `tracker` route to `<RedirectPreservingQuery to="/app/habits" />` for now; Task 6 retargets it.
- Create: `frontend/src/components/common/redirect-preserving-query.tsx`: `export function RedirectPreservingQuery({ to }: { to: string }): React.JSX.Element`. Body as in design D2.
- Create: `frontend/src/test/location-probe.tsx` (see "Shared test helper").
- Test: `frontend/src/components/common/__tests__/redirect-preserving-query.test.tsx`
- Test: `frontend/src/__tests__/app-routes.test.tsx`
- Test: `frontend/src/components/features/navigation/__tests__/nav-config.test.ts`

**Interfaces:**
- Produces: `AppRoutes`, `RedirectPreservingQuery`, and `LocationProbe`. It also produces the `app-routes.test.tsx` harness that Tasks 4–8 extend:
  - `vi.mock("@/components/features/navigation/protected-route", …)` → `ProtectedRoute: ({children}) => <>{children}</>`
  - `vi.mock("@/components/features/navigation/app-shell", …)` → `AppShell` stub that increments a module-level `shellMounts` counter in a mount-only `useEffect` and renders `<Outlet/>`.
  - Each page under test is mocked to a stub that renders `<div data-testid="page">{name}:{view ?? tab ?? ""}</div>`. Mock at least `TrackerPage`, `CalendarPage`, `IngredientsPage`, `IngredientDetailPage`, `ShoppingListsPage`, `ShoppingListDetailPage`, and `MealsPage`. Mock other pages only if importing them fails under jsdom.
  - Helper `renderAt(url: string)` renders `<MemoryRouter initialEntries={[url]}><AppRoutes/><LocationProbe/></MemoryRouter>`.

- [ ] **Step 1: Write the failing tests.**
  - `redirect-preserving-query.test.tsx`: at `/old?a=1#h` with route `/old` → `<RedirectPreservingQuery to="/new"/>`, the location becomes `/new?a=1` with hash `#h`, and nav-type is `REPLACE`.
  - `app-routes.test.tsx` (baseline that stays valid after this task):
    - `/app/tracker?x=1` → location `/app/habits?x=1`.
    - `/app/shopping` → location `/app/shopping/grocery`.
    - `/app/ingredients/0b6c…-uuid` → page `IngredientDetailPage:`.
    - `/app/shopping/grocery/list-1` → page `ShoppingListDetailPage:`.
  - `nav-config.test.ts`:
    - `isNavItemActive("/app/habits/setup", "/app/habits")` → `true`.
    - `isNavItemActive("/app/shopping/grocery/archived", "/app/shopping/grocery")` → `true`.
    - `isNavItemActive("/app/calendar/month", "/app/calendar")` → `true`.
    - `isNavItemActive("/app/ingredients/bulk", "/app/ingredients")` → `true`.
    - Read `nav-config.ts`. For each of those four `to` values, assert the nav item's `end` is falsy.
- [ ] **Step 2: Run the tests and confirm they fail.** Run `npx vitest run src/components/common/__tests__/redirect-preserving-query.test.tsx src/__tests__/app-routes.test.tsx src/components/features/navigation/__tests__/nav-config.test.ts`. Expected: FAIL on missing modules. The nav-config test may already pass, which is fine because it is a regression guard.
- [ ] **Step 3: Implement.** Create the component and the probe, extract `AppRoutes`, and rewire `App.tsx`.
- [ ] **Step 4: Run the new tests, the full suite, and the type check.** Run `npx vitest run && npx tsc -b`. Expected: PASS.
- [ ] **Step 5: Commit.** Message: `refactor(frontend): extract AppRoutes, add RedirectPreservingQuery`.

---

### Task 4: Grocery Lists — `tab` from route

**Files:**
- Modify: `frontend/src/pages/ShoppingListsPage.tsx`. Remove `tab` `useState` (line 31). Add props. Tab buttons (lines 102–123) navigate and get `aria-pressed`.
- Modify: `frontend/src/app-routes.tsx`. Add `shopping/grocery/archived` before `shopping/grocery/:id`. `shopping/grocery` passes `tab="active"`.
- Test: `frontend/src/pages/__tests__/ShoppingListsPage.test.tsx`. Update `renderPage` to pass `tab="active"`, then add a `describe("tabs")`.
- Test: `frontend/src/__tests__/app-routes.test.tsx`

**Interfaces:**
- Produces: `export type GroceryTab = "active" | "archived"` and `ShoppingListsPage(props: { tab: GroceryTab })`.

- [ ] **Step 1: Write the failing tests.**
  - app-routes: `/app/shopping/grocery` → page `ShoppingListsPage:active`. `/app/shopping/grocery/archived` → `ShoppingListsPage:archived` (static wins over `:id`, FR-S2). `/app/shopping/grocery/list-1` → `ShoppingListDetailPage:`.
  - Page: render the routes `/app/shopping/grocery` (`tab="active"`) and `/app/shopping/grocery/archived` (`tab="archived"`) with `LocationProbe`. Assertions:
    - At `/app/shopping/grocery`, `Active (…)` has `aria-pressed="true"` and `History (…)` has `"false"`.
    - Clicking History → location `/app/shopping/grocery/archived`, nav-type `PUSH`, and History is now pressed.
    - `probe-back` → location `/app/shopping/grocery`.
    - Starting at `/archived` renders the archived list. Make the `useShoppingLists` mock return different data per status argument.
- [ ] **Step 2: Run the tests and confirm they fail.** Run `npx vitest run src/pages/__tests__/ShoppingListsPage.test.tsx src/__tests__/app-routes.test.tsx`. Expected: FAIL.
- [ ] **Step 3: Implement.** Buttons call `navigate("/app/shopping/grocery")` or `navigate("/app/shopping/grocery/archived")`. Visual classes stay unchanged.
- [ ] **Step 4: Run the tests and the type check.** Same command plus `npx tsc -b`. Expected: PASS.
- [ ] **Step 5: Commit.** Message: `feat(frontend): grocery Active/History tabs addressable by URL`.

---

### Task 5: Ingredients — `view` from route

**Files:**
- Modify: `frontend/src/pages/IngredientsPage.tsx`. Remove the `view` `useState` (line 34) and add props. Rewire lines 119–135.
- Modify: `frontend/src/app-routes.tsx`. Replace `ingredients` with a `RedirectPreservingQuery` to `/app/ingredients/list`. Add the `ingredients/list|categories|bulk` routes. Keep `ingredients/:id`. Add **no** `ingredients/*` route.
- Test: `frontend/src/pages/__tests__/IngredientsPage.test.tsx` (new). Mock the `@/lib/hooks/api/*` hooks the page uses, `CategoryManager`, and `BulkCategorize`.
- Test: `frontend/src/__tests__/app-routes.test.tsx`

**Interfaces:**
- Produces: `export type IngredientsView = "list" | "categories" | "bulk"` and `IngredientsPage(props: { view: IngredientsView })`.

- [ ] **Step 1: Write the failing tests.**
  - app-routes:
    - `/app/ingredients` → location `/app/ingredients/list`.
    - `/app/ingredients?category=abc` → `/app/ingredients/list?category=abc`.
    - `/app/ingredients/{list,categories,bulk}` → `IngredientsPage:<view>`.
    - `/app/ingredients/3f2b8c1e-0000-4000-8000-000000000000` → `IngredientDetailPage:`.
    - `/app/ingredients/not-a-view` → `IngredientDetailPage:` (FR-I1 fall-through).
  - Page, with routes for list, categories, and bulk:
    - At `/app/ingredients/list?category=abc`, clicking Categories → location `/app/ingredients/categories?category=abc`, nav-type `PUSH`, and the category-manager stub is visible.
    - Clicking Categories again → `/app/ingredients/list?category=abc`.
    - At `/categories`, clicking Bulk → `/app/ingredients/bulk`.
    - At `/categories`, clicking "New ingredient" → location `/app/ingredients/list` **and** the create form input is visible.
- [ ] **Step 2: Run the tests and confirm they fail.** Run `npx vitest run src/pages/__tests__/IngredientsPage.test.tsx src/__tests__/app-routes.test.tsx`. Expected: FAIL.
- [ ] **Step 3: Implement.** Toggle target: `navigate({ pathname: \`/app/ingredients/${next}\`, search: location.search })`, where `next` is `list` when the pressed view is already active. For "New": if `view !== "list"`, navigate to list (push) and open the form. If the "New" test shows the form is lost because the page remounts, navigate with `state: { openCreate: true }` and seed `showCreate` from `location.state?.openCreate`. Note the remount in the PR (design D1).
- [ ] **Step 4: Run the tests and the type check.** Same command plus `npx tsc -b`. Expected: PASS.
- [ ] **Step 5: Commit.** Message: `feat(frontend): ingredients list/categories/bulk addressable by URL`.

---

### Task 6: Habits — `view` from route, `month` from query

**Files:**
- Modify: `frontend/src/pages/TrackerPage.tsx`. Remove both `useState`s and add the prop and the codec.
- Modify: `frontend/src/app-routes.tsx`. Add the Habits routes exactly as in design D1. Retarget `tracker` to `/app/habits/today`.
- Test: `frontend/src/pages/__tests__/TrackerPage.test.tsx`. Rewrite it. The calendar-grid mock exposes buttons for `onMonthChange("2026-01")`, `onViewReport`, and `onOpenSetup`, and renders its `month` prop. The month-report mock renders its `month` prop.
- Test: `frontend/src/__tests__/app-routes.test.tsx`

**Interfaces:**
- Consumes: `useQueryParamState` and `withParam` (Task 2), `getLocalMonth` (`lib/date-utils`).
- Produces: `export type HabitsView = "today" | "calendar" | "setup" | "report"` and `TrackerPage(props: { view: HabitsView })`. The `month` codec is local to the page (key `month`, regex `^\d{4}-(0[1-9]|1[0-2])$`, `current = getLocalMonth`).

- [ ] **Step 1: Write the failing tests.** Fake the clock to `2026-11-12` with `vi.useFakeTimers({ toFake: ["Date"] })` and `vi.setSystemTime`.
  - app-routes:
    - `/app/habits/{today,calendar,setup,report}` → `TrackerPage:<view>`.
    - `/app/habits`, `/app/habits/bogus`, and `/app/tracker` → location `/app/habits/today`.
    - `/app/habits?month=2026-01` → `/app/habits/today?month=2026-01`.
    - Mount stability: render `/app/habits/today`, push `/app/habits/calendar` through a probe button, and `shellMounts` stays `1`.
  - Page:
    - The `aria-pressed` of Today/Calendar/Setup matches the `view` prop. On `report`, all three are `"false"`.
    - Clicking Calendar from `/app/habits/today?month=2026-01` → `/app/habits/calendar?month=2026-01`, nav-type `PUSH`. Then `probe-back` → `/app/habits/today?month=2026-01`.
    - At `/app/habits/calendar`, the grid shows month `2026-11`. `onMonthChange("2026-01")` → location `/app/habits/calendar?month=2026-01`, nav-type `REPLACE`. Then `onMonthChange("2026-11")` → `/app/habits/calendar`.
    - `/app/habits/calendar?month=2026-13` → grid month `2026-11` and location `/app/habits/calendar`.
    - At `/app/habits/calendar?month=2026-01`, `onViewReport` → `/app/habits/report?month=2026-01` (PUSH) and the report shows `2026-01`.
    - `onOpenSetup` → `/app/habits/setup` (PUSH) and the setup stub is visible.
    - Width class: the container has `max-w-none` on calendar, `max-w-6xl` on report, and `max-w-3xl` on today and setup.
- [ ] **Step 2: Run the tests and confirm they fail.** Run `npx vitest run src/pages/__tests__/TrackerPage.test.tsx src/__tests__/app-routes.test.tsx`. Expected: FAIL.
- [ ] **Step 3: Implement** per design D6 "Habits". The header buttons and `onOpenSetup` navigate `{ pathname: \`/app/habits/${id}\`, search: location.search }`. `onViewReport` does the same with `report`.
- [ ] **Step 4: Run the tests and the type check.** Same command plus `npx tsc -b`. Expected: PASS.
- [ ] **Step 5: Commit.** Message: `feat(frontend): habits views and month addressable by URL`.

---

### Task 7: Meal Planner — `week` from query

**Files:**
- Modify: `frontend/src/pages/MealsPage.tsx:57`. Replace `useState<Date>(getLocalWeekStart)` with the hook.
- Test: `frontend/src/pages/__tests__/MealsPage.test.tsx` (new). Mock `usePlans` and the other API hooks the page calls. Assert `usePlans` was called with `{ starts_on: … }`. Stub `WeekSelector` as buttons that call `onWeekChange` with fixed dates and render `startsOn`.

**Interfaces:**
- Consumes: `useQueryParamState` (Task 2), `parseLocalDateStr`/`formatLocalDateStr` (Task 1), `getLocalWeekStart`, and `isoMondayOf` (`lib/workout-week.ts`).
- Produces: a page-local `week` codec. `key: "week"`. `parse` = `parseLocalDateStr`, then normalize to that week's Monday (e.g. `parseLocalDateStr(isoMondayOf(d))`). `serialize = formatLocalDateStr`. `current = () => getLocalWeekStart()`. `equals` compares serialized strings.

- [ ] **Step 1: Write the failing tests.** Fake the date to Thu `2026-11-12`, so the current Monday is `2026-11-09`.
  - `/app/meals` → `usePlans` called with `starts_on: "2026-11-09"`, and location stays `/app/meals`.
  - `/app/meals?week=2026-11-02` → `starts_on: "2026-11-02"`, and the location is unchanged.
  - `/app/meals?week=2026-11-04` (Wednesday) → `starts_on: "2026-11-02"`, location `/app/meals?week=2026-11-02`, nav-type `REPLACE`.
  - `/app/meals?week=2026-02-31` and `?week=2026-11-09` both → location `/app/meals`.
  - Selector "next week" → location `/app/meals?week=2026-11-16`, nav-type `REPLACE`. "This week" → `/app/meals`.
- [ ] **Step 2: Run the tests and confirm they fail.** Run `npx vitest run src/pages/__tests__/MealsPage.test.tsx`. Expected: FAIL.
- [ ] **Step 3: Implement.** `const [startsOn, setStartsOn] = useQueryParamState(weekCodec)`. Everything downstream stays unchanged.
- [ ] **Step 4: Run the tests and the type check.** Same command plus `npx tsc -b`. Expected: PASS.
- [ ] **Step 5: Commit.** Message: `feat(frontend): meal planner week addressable by URL`.

---

### Task 8: Calendar — `view` from route, `date` from query, OAuth preserved

**Files:**
- Create: `frontend/src/components/features/calendar/calendar-url.ts`
- Test: `frontend/src/components/features/calendar/__tests__/calendar-url.test.ts`
- Modify: `frontend/src/pages/CalendarPage.tsx`. Remove the `weekStart`, `viewMode`, and `monthAnchor` `useState`s (lines 51–53) and the `isDesktop` effect (lines 70–81). Rewrite `focusDay`, `goPrev`/`goNext`, `handleCalendarSelect`, `handleDayClick`, `handleViewModeChange`, and the OAuth effect (lines 83–100).
- Modify: `frontend/src/app-routes.tsx`. Add the Calendar routes exactly as in design D1.
- Test: `frontend/src/pages/__tests__/CalendarPage.test.tsx` (new). Mock `use-calendar`, `use-packages`, and `sonner` `toast`. Stub `CalendarGrid` to render `weekStart`. Stub `MonthGrid` to render `monthAnchor` and a button that calls its day-click prop with `new Date(2026, 2, 18)`. Stub the dialogs to null.
- Test: `frontend/src/__tests__/app-routes.test.tsx`

**Interfaces:**
- Consumes: `useQueryParamState`/`withParam` (Task 2), `parseLocalDateStr`/`formatLocalDateStr` (Task 1), `getStartOfWeek`/`getStartOfMonth`/`isSameMonth` (`calendar-utils`), and `CalendarViewMode` (`view-mode-toggle`).
- Produces (from `calendar-url.ts`):
  ```ts
  export function weekAnchorFor(day: Date, isDesktop: boolean): Date;
  // desktop: getStartOfWeek(day); mobile: day − 1 at 00:00 (today's focusDay math)
  export function calendarDateCodec(view: CalendarViewMode, isDesktop: boolean): QueryParamCodec<Date>;
  // key "date"; parse = parseLocalDateStr then normalize per design D6 table; serialize = formatLocalDateStr;
  // current: month → getStartOfMonth(new Date()); week → weekAnchorFor(new Date(), isDesktop); equals by serialized string
  ```
- Produces: `CalendarPage(props: { view: CalendarViewMode })`.

- [ ] **Step 1: Write the failing unit tests** in `calendar-url.test.ts`. Fake the date to Thu `2026-11-12`.
  - `weekAnchorFor(new Date(2026,10,12), true)` → `2026-11-08`. With `false` → `2026-11-11`.
  - Codec `month`: `parse("2026-03-17")` → `2026-03-01`. `current()` → `2026-11-01`.
  - Codec `week` desktop: `parse("2026-03-18")` → `2026-03-15`. `current()` → `2026-11-08`.
  - Codec `week` mobile: `parse("2026-03-18")` → `2026-03-18`, not normalized. `current()` → `2026-11-11`.
  - Every codec: `parse("2026-02-31")` → `null`. For each view, `serialize(parse(serialize(x))!) === serialize(x)` (fixed point, design §5).
- [ ] **Step 2: Run the unit tests and confirm they fail.** Run `npx vitest run src/components/features/calendar/__tests__/calendar-url.test.ts`. Expected: FAIL.
- [ ] **Step 3: Implement `calendar-url.ts`.**
- [ ] **Step 4: Run the unit tests.** Same command. Expected: PASS.
- [ ] **Step 5: Write the failing route and page tests.** jsdom `matchMedia`: stub it to desktop (`matches: true`) in the test setup for this file. Check whether `src/test/setup.ts` already stubs it. Fake date `2026-11-12`.
  - app-routes:
    - `/app/calendar/{week,month}` → `CalendarPage:<view>`.
    - `/app/calendar` and `/app/calendar/bogus` → `/app/calendar/week`.
    - `/app/calendar?connected=true` → `/app/calendar/week?connected=true`.
    - `/app/calendar?error=auth_failed` → `/app/calendar/week?error=auth_failed`.
  - Page:
    - `/app/calendar/week` → grid `weekStart` is `2026-11-08`. "Next period" → location `/app/calendar/week?date=2026-11-15`, nav-type `REPLACE`. "Previous period" → `/app/calendar/week`.
    - `/app/calendar/month?date=2026-03-17` → month grid anchor `2026-03-01`, location `/app/calendar/month?date=2026-03-01` (REPLACE, one rewrite).
    - Week → Month toggle at `/app/calendar/week?date=2026-03-15` → `/app/calendar/month?date=2026-03-01`, nav-type `PUSH`.
    - Month → Week at `/app/calendar/month?date=2026-03-01` → `/app/calendar/week?date=2026-03-01`. This is week 1 of March, because today is not in March (FR-C4).
    - Month → Week at `/app/calendar/month` → `/app/calendar/week`, because the current week omits the param.
    - Day click at `/app/calendar/month?date=2026-03-01` → `/app/calendar/week?date=2026-03-15` (PUSH). Then `probe-back` → `/app/calendar/month?date=2026-03-01` (FR-C5, Review Focus 5).
    - `/app/calendar/week?connected=true` → `toast.success` called once, location `/app/calendar/week`.
    - `/app/calendar/week?connected=true&date=2026-03-15` → toast once, location `/app/calendar/week?date=2026-03-15`.
    - `/app/calendar/week?error=auth_failed` → `toast.error("Google Calendar authorization failed. Please try again.")`.
- [ ] **Step 6: Run the route and page tests and confirm they fail.** Run `npx vitest run src/pages/__tests__/CalendarPage.test.tsx src/__tests__/app-routes.test.tsx`. Expected: FAIL.
- [ ] **Step 7: Implement CalendarPage** per design D6 "Calendar":
  - `const codec = useMemo(() => calendarDateCodec(view, isDesktop), [view, isDesktop])`, then `const [anchor, setAnchor] = useQueryParamState(codec)`.
  - `weekStart = view === "week" ? anchor : getStartOfWeek(anchor)` and `monthAnchor = view === "month" ? anchor : getStartOfMonth(anchor)`.
  - `goPrev`/`goNext` call `setAnchor` with `addMonths(±1)` on month, or `±dayCount` days on week.
  - `handleCalendarSelect` → `setAnchor(view === "month" ? getStartOfMonth(d) : weekAnchorFor(d, isDesktop))`.
  - View change and day click `navigate({ pathname, search: withParam(searchParams, calendarDateCodec(target, isDesktop), value) })`.
  - In the OAuth effect, both `setSearchParams` calls become `setSearchParams(prev => { const p = new URLSearchParams(prev); p.delete("connected"); p.delete("error"); return p; }, { replace: true })`.
  - Replace `viewMode` with `view` everywhere and pass `mode={view}` to `ViewModeToggle`.
- [ ] **Step 8: Run the Calendar tests, the full suite, and the type check.** Run `npx vitest run && npx tsc -b && npx eslint src`. Expected: PASS, no lint errors. If the removed effect leaves an unused `eslint-disable` comment, remove the comment too.
- [ ] **Step 9: Commit.** Message: `feat(frontend): calendar week/month and date addressable by URL`.

---

### Task 9: Repo gate

**Files:** none, unless the gate finds something.

- [ ] **Step 1: Run the full gate.** From the worktree root, run `tools/verify.sh` with no flags, in the background, using the task-verifier agent. Expected: exit 0.
- [ ] **Step 2: Fix and re-run if needed.** On failure, fix the first failing block in the task that owns the file, commit `fix(frontend): …`, and re-run Step 1. Do not report done from a flagged or partial run.
