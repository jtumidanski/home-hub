# Frontend Audit — task-050 Dashboard Management Page

- **Audit Scope:** Branch `task-050-dashboard-management-page` (frontend changes only, per `git diff main...HEAD`)
- **Guidelines Source:** frontend-dev-guidelines skill
- **Date:** 2026-06-05
- **Build:** PASS
- **Tests:** 72 passed, 0 failed (18 scoped test files)
- **Overall:** NEEDS-WORK (build/tests green; only MINOR/non-blocking observations — no hard-guideline FAIL)

## Build & Test Results

- `npm run build` (tsc -b && vite build): PASS — 3429 modules transformed, built in 735ms. Only the pre-existing chunk-size advisory warning (unrelated to this change set).
- `npx vitest run` over the in-scope test files: `Test Files 18 passed (18)`, `Tests 72 passed (72)`.

## File Inventory

- `frontend/src/lib/dashboard/ordering.ts` — **Other** (pure util: sort + reorder)
- `frontend/src/lib/dashboard/__tests__/ordering.test.ts` — Test (unit)
- `frontend/src/components/features/dashboards/dashboard-row.tsx` — **Component** (feature, presentational row)
- `frontend/src/components/features/dashboards/__tests__/dashboard-row.test.tsx` — Test
- `frontend/src/components/features/dashboards/dashboard-kebab-menu.tsx` — **Component** (trigger className only changed)
- `frontend/src/pages/DashboardManagementPage.tsx` — **Page**
- `frontend/src/pages/__tests__/DashboardManagementPage.test.tsx` — Test
- `frontend/src/components/features/navigation/dashboards-nav-group.tsx` — **Component** (rewritten: management affordances removed)
- `frontend/src/components/features/navigation/__tests__/dashboards-nav-group.test.tsx` — Test
- `frontend/src/components/features/navigation/nav-config.ts` — **Other** (nav config: added "Manage Dashboards")
- `frontend/src/App.tsx` — **Other** (route wiring)
- `frontend/src/pages/DashboardsIndexRedirect.tsx` — Deleted (no dangling references remain)

## Anti-Pattern Checklist

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| FE-01 | No `any` type | PASS | grep `: any`/`as any` across all in-scope files → zero matches |
| FE-02 | No manual class concat | PASS | All conditional classes use `cn()` (dashboard-row.tsx:61, nav-group.tsx:45,62,67, kebab:93). No `className={"..." + }` or template concat found |
| FE-03 | No direct API client in components/pages | PASS | grep `lib/api/client` in row/page/kebab/nav-group → zero. Page consumes `useDashboards`/`useReorderDashboards`/`useHouseholdPreferences` hooks (DashboardManagementPage.tsx:17-18,28-30); kebab consumes hooks (kebab:24-28); no direct fetches |
| FE-04 | No inline Zod in components | PASS (in scope) | Only inline `z.object` is `renameSchema` at dashboard-kebab-menu.tsx:32, and it composes the exported `dashboardNameSchema` primitive from new-dashboard-modal.tsx:31. This line is **pre-existing** — the branch diff for this file touches only the trigger className (line 94). Out of scope; not newly introduced |
| FE-05 | No spinners for content loading | PASS | grep `animate-spin` in dashboards/nav/page → zero. Page loading uses `<Skeleton>` (DashboardManagementPage.tsx:92-99). Delete/save buttons use text ("Deleting...", "Saving...") not spinners |
| FE-06 | No hardcoded colors | PASS | grep for `bg-/text-/border-` + raw palette in row/kebab/page/nav-group → zero. Semantic tokens used: `bg-card`, `text-muted-foreground`, `bg-accent`, `text-accent-foreground`, `bg-sidebar-accent`, `text-destructive` |
| FE-07 | No state mutation | PASS | `ordering.ts:7` `.sort()` operates on `[...list]` (copy); `:29-32` `.splice()` operate on `next = [...sorted]` (copy). Never mutates props/state. `computeReorderEntries` returns a fresh `.map()` payload. `moved` guarded by `if (!moved) return null` (`:31`) satisfying `noUncheckedIndexedAccess` |
| FE-08 | No default exports for components | PASS | grep `export default` in all component/page files → zero. Named: `DashboardRow`, `DashboardKebabMenu`, `DashboardManagementPage`, `DashboardsNavGroup` |
| FE-09 | Tenant guard in hooks | PASS (consumed hooks) | No new hooks authored. Consumed `useDashboards` (use-dashboards.ts:26-35) and `useHouseholdPreferences` (use-household-preferences.ts:15-23) both use `useTenant()` and `enabled: !!tenant?.id && !!household?.id`. Page uses these correctly |
| FE-10 | Tenant ID in query keys | PASS | `dashboardKeys` (use-dashboards.ts:15-22) and `householdPreferencesKeys` (use-household-preferences.ts:10-13) include `tenant?.id ?? "no-tenant"` and household id. Not modified by this branch but verified for correct consumption |
| FE-11 | Error handling via `createErrorFromUnknown` | PASS | Page surfaces query errors via `isError → <ErrorCard>` (DashboardManagementPage.tsx:102-108). Mutation error handling lives in the hooks (use-dashboards.ts:61-68,81-88,103-105,118-120 etc.) using `createErrorFromUnknown`/`getErrorMessage` + toast. No raw `.catch` in components |

## Architecture Checklist

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| FE-12 | JSON:API model shape | PASS | `Dashboard` = `{ id, type, attributes }` (types/models/dashboard.ts:15-19). `ordering.ts` reads `d.attributes.sortOrder/createdAt`; row reads `dashboard.attributes.name`. Page reads `data?.data ?? []` matching `ApiListResponse<T> = { data: T[] }` (base.ts:29-32) |
| FE-13 | Service extends BaseService | N/A | No service authored/changed in this branch. (Consumed `householdPreferencesService` extends BaseService — household-preferences.ts:9) |
| FE-14 | Query key factory uses `as const` | PASS | Verified in consumed factories (use-dashboards.ts:15-22 all `as const`); none authored here |
| FE-15 | Forms use react-hook-form + zodResolver | PASS | RenameDialog (kebab:202-211) uses `useForm({ resolver: zodResolver(renameSchema) })`. New-dashboard modal reused unchanged. Page itself has no form |
| FE-16 | Schema paired with `z.infer` type | PASS | `renameSchema` paired with `type RenameFormData = z.infer<typeof renameSchema>` (kebab:32-33); `dashboardNameSchema` exported from modal (pre-existing) |

## Styling Checklist

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| FE-19 | Interactive elements show cursor affordance | PASS (newly-authored elements) | dashboard-row grip is `<button>` with `cursor-grab` (row:36) — correct drag-handle affordance. Edit is a native `<a>` Link (`router-dom`) with `href` (row:58-67) — browser supplies pointer; Tailwind preflight does not reset anchor cursor. Kebab `MenuPrimitive.Trigger` renders `<button>`; its className change in this branch (kebab:93-95) did not add `cursor-pointer`, but this matches the surrounding sidebar trigger convention and the trigger's interactivity is unchanged — see Non-Blocking note. nav-group `CollapsibleTrigger` (nav-group:61) lacks `cursor-pointer`, but the line is **byte-identical to main** (pre-existing, out of scope) |

## Testing Checklist

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| FE-17 | Tests exist for changed components | PASS | ordering.ts→ordering.test.ts (sort + reorder, 5 cases); dashboard-row.tsx→dashboard-row.test.tsx (link/edit/kebab/grip labels + default badge); DashboardManagementPage.tsx→DashboardManagementPage.test.tsx (loading/error/sections/empty/modal); dashboards-nav-group.tsx→dashboards-nav-group.test.tsx (sorted links, no management affordances, empty). All use MemoryRouter and assert behavior via `getByRole`/`getByText`, not mock internals |
| FE-18 | Mocks updated when services changed | N/A | No service interface changed in this branch |

## Summary

### Blocking (must fix)
- None. Build passes, all 72 scoped tests pass, and no hard-guideline FAIL was found within the change set.

### Non-Blocking (should fix) — MINOR
- **[FE-17 / test fidelity] Preference mock shape mismatch (Known Observation 1, confirmed).** Both `DashboardManagementPage.test.tsx:56-57` and `dashboard-row.test.tsx:14-16` mock `useHouseholdPreferences` returning `data: { data: { ...object } }`, but the real path is `ApiListResponse` → `data: { data: [ ...array ] }` (base.ts:29-32; service comment household-preferences.ts:14-21). The page consumes it as `prefsData?.data?.[0]?.attributes.defaultDashboardId` (DashboardManagementPage.tsx:45-46), i.e. array indexing. Against the object mock, `data[0]` is `undefined` → `defaultDashboardId` resolves to `null`. This is **inert**: no test asserts the default-badge path through the page (the badge is exercised by passing `defaultDashboardId` directly to `DashboardRow` in dashboard-row.test.tsx:67-68). Recommend aligning the mock to `data: { data: [ {prefs} ] }` so a future default-badge integration assertion is exercisable. Severity: MINOR.
- **[FE-19 / accessibility-styling] Grip and edit-link use `outline-none` without a `focus-visible` ring (Known Observation 2, confirmed).** dashboard-row.tsx:36 (grip button) and :63 (edit link) include `outline-none` with no `focus-visible:ring-*`/`focus-visible:border-*` replacement, so keyboard focus has no visible indicator on these two controls. The sibling kebab trigger shares this exact pattern (kebab:94), and the project's base `<Button>` demonstrates the intended remedy (`focus-visible:ring-3 focus-visible:ring-ring/50`, button.tsx:9). Functionally keyboard reorder still works (KeyboardSensor wired in DashboardManagementPage.tsx:50), but the focus affordance is missing. Severity: MINOR (consistency with an established pre-existing pattern; not a regression).
- **[FE-19 informational] Kebab `MenuPrimitive.Trigger` (`<button>`) carries no `cursor-pointer`** after this branch's className edit (kebab:93-95). Tailwind preflight resets native `<button>` to `cursor: default` (hence the base Button CVA opts into `cursor-pointer`). The grip correctly uses `cursor-grab`; the kebab trigger does not declare any cursor. The pre-edit className also lacked `cursor-pointer`, so this is not a regression, but adding `cursor-pointer` would bring it in line with FE-19. Severity: MINOR / optional.

### Notes
- `ordering.ts` `.sort`/`.splice` flagged by the mechanical FE-07 grep are **correct** immutable usage (operate on spread copies, return new arrays) — not violations.
- The inline `z.object` in `dashboard-kebab-menu.tsx:32` is pre-existing and outside this branch's diff for that file (only the trigger className changed); not counted against FE-04.
- `CollapsibleTrigger` missing `cursor-pointer` in `dashboards-nav-group.tsx:61` is byte-identical to `main`; pre-existing, out of scope.
