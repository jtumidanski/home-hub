# Plan Audit — task-046-kiosk-dashboard-widgets

**Plan Path:** `docs/tasks/task-046-kiosk-dashboard-widgets/plan.md`
**Audit Date:** 2026-05-01
**Branch:** `feature/task-046-kiosk-dashboard-widgets`
**Base Branch:** `main` (804881c)
**Branch HEAD:** 19b366f
**Verdict:** **PASS**

## Executive Summary

All 26 plan tasks (A1, A2, B1-B3, C1-C3, D1-D3, E1-E7, F1-F3, G1-G2, H1-H2, I1) are implemented with file-and-line evidence in the working tree. Backend and frontend tests for the affected packages are green: 5 new processor seed-key tests pass, 4 new account-service kiosk-flag tests pass, all 5 new widget schemas + adapters pass, and the four-scenario `DashboardRedirect` orchestration test passes. Three intentional plan deviations (camelCase JSON-tag fix, GORM tenant-scoped UPDATE in C3, and an E1+E2 merged commit) were called out in the prompt and verified to be improvements rather than regressions. Pre-existing repo-wide noise (unrelated lint errors in `use-cooklang-preview.ts` / `DashboardDesigner.tsx` / `WorkoutReviewPage.test.tsx`, missing `react-resizable/css/styles.css` for 4 dashboard-designer test suites, 2 flaky tracker-service entry tests) is documented but not introduced by this branch.

## Task Completion

| # | Task | Status | Evidence / Notes |
|---|------|--------|------------------|
| A1 | Allowlist 5 widget types in shared/go/dashboard | PASS | `shared/go/dashboard/types.go:27-31` (5 new map entries), `shared/go/dashboard/fixtures/widget-types.json` (14 entries alphabetised), `shared/go/dashboard/types_test.go:11-16` asserts new types. Commit `0047a75`. |
| A2 | Allowlist 5 widget types in frontend mirror | PASS | `frontend/src/lib/dashboard/widget-types.ts:11-15` (5 new strings), `frontend/src/lib/dashboard/fixtures/widget-types.json` matches Go. Commit `d965623`. |
| B1 | Add `seed_key` column + partial unique index + brownfield backfill | PASS | `services/dashboard-service/internal/dashboard/entity.go:20` (SeedKey field), `entity.go:35-46` (CREATE UNIQUE INDEX `idx_dashboards_seed_key` ... WHERE seed_key IS NOT NULL, plus brownfield backfill UPDATE). Commit `853ebb1`. |
| B2 | Rework `Processor.Seed` for optional `seedKey` | PASS | `services/dashboard-service/internal/dashboard/processor.go:118` Seed signature `(tenantID, householdID, callerUserID, name, seedKey *string, layoutJSON)`. Branches on seedKey nil vs non-nil at `processor.go:133-217`. Advisory lock at `acquireSeedLock processor.go:224-235` (postgres-only). Tests: `TestProcessorSeedByKeyIdempotent` (processor_test.go:385), `TestProcessorSeedDistinctKeysCoexist` (processor_test.go:403), `TestProcessorSeedNilKeyPreservesLegacyBehavior` (processor_test.go:420). Existing `TestProcessorSeedIdempotent` and `TestProcessorSeedRace` updated to pass `nil`. Commit `ed0067f`. |
| B3 | Accept optional `key` on POST /dashboards/seed | PASS | `services/dashboard-service/internal/dashboard/rest.go:65-73` (`SeedRequest.Key *string`), `services/dashboard-service/internal/dashboard/resource.go:22` (regex `^[a-z][a-z0-9-]{0,39}$`), `resource.go:327-...` `seedHandler` validates and forwards key. Tests: `TestSeedHandlerWithKioskKey` (rest_test.go:368) and `TestSeedHandlerRejectsMalformedKey` (rest_test.go:388). Commit `141bf26`. |
| C1 | Add `kiosk_dashboard_seeded` column + accessors | PASS | `services/account-service/internal/householdpreference/entity.go:16` (KioskDashboardSeeded bool, gorm column, default false). `model.go:15` field, `model.go:25` `KioskDashboardSeeded()` getter. `builder.go:22,56-59,87` builder pipe-through. `entity.go:32` ToEntity carries field. Round-trip test added at `builder_test.go:104,122`. Commits `7f1bbb5`, `e868a5b`, `4828773`. |
| C2 | Expose flag on GET /household-preferences | PASS | `services/account-service/internal/householdpreference/rest.go:12` (`KioskDashboardSeeded bool` with json tag), `rest.go:29` Transform copies field. Test `TestGetHouseholdPreferencesIncludesKioskFlag` at `rest_test.go:94` asserts `"kioskDashboardSeeded":false` in body. Commit `09eef99`. NB: JSON wire shape is camelCase per the follow-up fix (`9e59a13`); plan originally expected snake_case but the fix aligned all four `RestModel` tags. |
| C3 | PATCH /household-preferences/{id}/kiosk-seeded sub-route | PASS | Route at `resource.go:24`, handler `markKioskSeededHandler resource.go:55-87` (only `value:true` accepted, 400 on `false`/decode-error, 404 on missing row, 500 on update error). Processor method at `processor.go:60-66`. SQL helper at `administrator.go:28-45` uses `db.Model(&Entity{}).Where(...).Updates(...)` (tenant-scope-friendly) and returns `gorm.ErrRecordNotFound` when `RowsAffected == 0`. Tests: `TestMarkKioskSeededFlipsFlag` (rest_test.go:201) including idempotency, `TestMarkKioskSeededRejectsFalse` (rest_test.go:245), `TestMarkKioskSeededReturnsNotFoundForUnknownID` (rest_test.go:264). Commits `e817022`, `4dc9810`. |
| D1 | seedDashboard + useSeedDashboard accept optional key | PASS | `frontend/src/services/api/dashboard.ts:70-86` (key param, conditional attribute). `frontend/src/lib/hooks/api/use-dashboards.ts:163-176` mutation signature `{ name, layout, key? }`. Commit `a4d4446`. |
| D2 | Expose `kioskDashboardSeeded` + `useMarkKioskSeeded` | PASS | `frontend/src/types/models/dashboard.ts:39-44` `HouseholdPreferencesAttributes` extended. `frontend/src/services/api/household-preferences.ts:41-47` `markKioskSeeded` method posts `{value:true}`. `frontend/src/lib/hooks/api/use-household-preferences.ts:40-51` `useMarkKioskSeeded` (silent error mode per plan). The wire-format camelCase fix (`9e59a13`) makes the snake-case→camelCase concern moot. Commit `b7046ec`. |
| D3 | useLocalDateOffset + getLocalDateStrOffset | PASS | `frontend/src/lib/date-utils.ts:70+` `getLocalDateStrOffset(tz, offsetDays)`. `frontend/src/lib/hooks/use-local-date-offset.ts:11-24` polls every 60s via `useSyncExternalStore`. Tests: `frontend/src/lib/__tests__/date-utils-offset.test.ts`, `frontend/src/lib/hooks/__tests__/use-local-date-offset.test.ts`. Commit `5cd225c`. |
| E1 | tasks-today widget definition | PASS | `frontend/src/lib/dashboard/widgets/tasks-today.ts` exists, schema test at `frontend/src/lib/dashboard/__tests__/widgets/tasks-today.schema.test.ts`. (Plan said E1 commit pairs with E2 — verified merged into `52f5988`.) |
| E2 | tasks-today adapter | PASS | `frontend/src/components/features/dashboard-widgets/tasks-today-adapter.tsx` (Read-only marker comment, header `<Link>`). Adapter test at `frontend/src/components/features/dashboard-widgets/__tests__/tasks-today-adapter.test.tsx` (5 tests: skeleton, overdue+today render, all-completed fallback, empty, error). Commit `52f5988`. |
| E3 | reminders-today widget + adapter | PASS | `frontend/src/lib/dashboard/widgets/reminders-today.ts` and `frontend/src/components/features/dashboard-widgets/reminders-today-adapter.tsx`. Tests at `__tests__/widgets/reminders-today.schema.test.ts` and `dashboard-widgets/__tests__/reminders-today-adapter.test.tsx`. Commit `173fa51`. |
| E4 | weather-tomorrow widget + adapter | PASS | `frontend/src/lib/dashboard/widgets/weather-tomorrow.ts`, `frontend/src/components/features/dashboard-widgets/weather-tomorrow-adapter.tsx`. Schema test asserts metadata + units enum + null default. Adapter test exercises tomorrow lookup + missing-tomorrow fallback. Commit `f2f18fd`. |
| E5 | calendar-tomorrow widget + adapter | PASS | `frontend/src/lib/dashboard/widgets/calendar-tomorrow.ts`, `frontend/src/components/features/dashboard-widgets/calendar-tomorrow-adapter.tsx` (uses `useCalendarEvents(start,end)`, sorts all-day-first, renders `+N more`). Tests cover sort + empty. Commit `79c35f9`. |
| E6 | tasks-tomorrow widget + adapter | PASS | `frontend/src/lib/dashboard/widgets/tasks-tomorrow.ts`, `frontend/src/components/features/dashboard-widgets/tasks-tomorrow-adapter.tsx`. Tests cover filter to tomorrow's pending + empty state. Commit `669ed4a`. |
| E7 | Register all 5 in widget-registry | PASS | `frontend/src/lib/dashboard/widget-registry.ts:13-17` imports + `:47-51` array entries. Test `frontend/src/lib/dashboard/__tests__/widget-registry.test.ts` extended to assert all 5 resolve via `findWidget`. Commit `be291ec`. |
| F1 | meal-plan-today schema with `view` field | PASS | `frontend/src/lib/dashboard/widgets/meal-plan-today.ts` schema includes `view: z.enum(["list","today-detail"]).default("list")`, defaultConfig `{horizonDays:1, view:"list"}`. Schema test file `frontend/src/lib/dashboard/__tests__/widgets/meal-plan-today.schema.test.ts` covers accept/reject/default. Commit `f73d466`. |
| F2 | MealPlanTodayDetail component | PASS | `frontend/src/components/features/meals/meal-plan-today-detail.tsx` renders today's full B/L/D + N-1 dinner peek. Test at `frontend/src/components/features/meals/__tests__/meal-plan-today-detail.test.tsx` asserts horizonDays=3 shows next 2 dinners and skips lunches/beyond-horizon, and horizonDays=1 collapses to today only. Commit `65ece76`. |
| F3 | Branch meal-plan adapter on view | PASS | `frontend/src/components/features/dashboard-widgets/meal-plan-adapter.tsx` branches `if config.view === "today-detail"` → `MealPlanTodayDetail`, else `MealPlanWidget`. Test at `dashboard-widgets/__tests__/meal-plan-adapter.test.tsx` covers list, today-detail, absent-view-defaults-to-list. Commit `dc46748`. |
| G1 | kiosk-seed-layout | PASS | `frontend/src/lib/dashboard/kiosk-seed-layout.ts:1-29` exports `kioskSeedLayout()` returning the 4-column / 8-widget composition with the exact ordering the plan specifies (weather, meal-plan-today/today-detail/h:5, tasks-today, calendar-today/h:12, weather-tomorrow, calendar-tomorrow, tasks-tomorrow, reminders-today). Test `__tests__/kiosk-seed-layout.test.ts` asserts schema-valid + every widget in registry + columns sum + fresh UUIDs + exact order + meal-plan-today seeded with `view:"today-detail"` & `horizonDays:3`. Commit `0d7ff76`. |
| G2 | Broaden parity test | PASS | `frontend/src/pages/__tests__/dashboard-renderer-parity.test.tsx` rewritten to take the union of `seedLayout()` ∪ `kioskSeedLayout()` and assert it covers `widgetRegistry` and `WIDGET_TYPES`. Commit `52da873`. |
| H1 | Two-seed orchestration with kiosk preference gate | PASS | `frontend/src/pages/DashboardRedirect.tsx:28-89` implements: read prefs + dashboards, gate kiosk seed on `(!kioskFlag && !hasKioskRow)`, run home + kiosk seeds via `Promise.allSettled`, conditionally refetch only if anything was seeded, mark seeded if Kiosk row now exists, fall through prefId → first household → first user. Commit `8c55902`. |
| H2 | Cover four seeding scenarios | PASS | `frontend/src/pages/__tests__/DashboardRedirect.test.tsx:190-365` `describe("two-seed orchestration")` block contains exactly the four required scenarios: brand-new (191), brownfield (249), returning-both-seeded-no-mutations (293), deleted-Kiosk-flag-prevents-reseed (332). Each verifies the correct mutation calls (and absence of) and the correct navigation target. Commit `023d899`. |
| I1 | Full-suite green sweep + lint cleanup | PASS | All task-046 tests pass: 20 test files / 64 tests when filtered to task-046 surface (`npx pnpm test --run` on task-046 paths). Backend tests pass for affected services. The two task-046 lint fixups (eslint-disable for unused config arg in `weather-tomorrow-adapter`, and `as unknown as ...` cast in registry) were committed as `19b366f`. |

**Completion Rate:** 26 / 26 tasks (100%)
**Skipped without approval:** 0
**Partial implementations:** 0

## Skipped / Deferred Tasks

None. Every checkbox in the plan has corresponding code, tests, and commit evidence on the branch.

## Notable Plan Deviations (all approved / improvements)

1. **JSON-tag camelCase realignment (`9e59a13`).** The plan tests in `rest_test.go` were originally drafted to assert `"kiosk_dashboard_seeded":false`. Investigation found the existing `default_dashboard_id` snake-case tag was already broken (the rest of the API uses camelCase consistently). The fix realigned all four `RestModel` JSON tags to camelCase, updating the C2 assertion accordingly. This made the wire format consistent end-to-end (Go RestModel → frontend `HouseholdPreferencesAttributes` → DashboardRedirect read of `kioskDashboardSeeded`). Verified by reading `services/account-service/internal/householdpreference/rest.go:9-15` (camelCase) ↔ `frontend/src/types/models/dashboard.ts:39-44` (camelCase) ↔ `DashboardRedirect.tsx:48` (`prefRow?.attributes.kioskDashboardSeeded`).
2. **Tenant-scoped UPDATE via GORM model (`4dc9810`).** Plan code used `db.Exec("UPDATE …")` for the kiosk-seeded flip, which bypasses the GORM tenant callback. Replaced with `db.Model(&Entity{}).Where("id = ?", id).Updates(map[string]any{...})` so the tenant-scope auto-injection still applies. Also added a proper `RowsAffected == 0 → gorm.ErrRecordNotFound` branch so the 404 test (`TestMarkKioskSeededReturnsNotFoundForUnknownID`) passes. Verified at `administrator.go:28-45`.
3. **E1 + E2 single commit (`52f5988`).** Plan expected E1 to land without a commit (because it imports the not-yet-existing E2 adapter), then commit at E2. Implementation merged both into one commit per the plan's own instruction at line 1411 ("No commit; the adapter file is required for the build."). This matches the documented intent.
4. **Two C1 follow-ups (`e868a5b` gofmt + `4828773` round-trip test).** Both expand the C1 work on the original task scope and were completed before C2 began. Verified the round-trip test asserts `KioskDashboardSeeded()` round-trips through `ToEntity → Make`.
5. **I1 lint sweep (`19b366f`).** Two task-046 files needed minor lint fixups (unused-config arg in `weather-tomorrow-adapter.tsx`, an `as unknown as ...` cast in the widget registry). Both touch only task-046 files and don't change runtime behaviour.

## Build & Test Results

| Service / Package | Build | Tests | Notes |
|-------------------|-------|-------|-------|
| `shared/go/dashboard` | PASS | PASS | `TestIsKnownWidgetType`, `TestWidgetTypesParityFixture`, `TestLayoutConstants` all green. |
| `services/dashboard-service` | PASS | PASS | All seed-key tests green: `TestProcessorSeedByKeyIdempotent`, `TestProcessorSeedDistinctKeysCoexist`, `TestProcessorSeedNilKeyPreservesLegacyBehavior`, `TestProcessorSeedIdempotent`, `TestProcessorSeedRace`, `TestSeedHandlerWithKioskKey`, `TestSeedHandlerRejectsMalformedKey`. |
| `services/account-service` | PASS | PASS | All flag tests green: `TestGetHouseholdPreferencesIncludesKioskFlag`, `TestMarkKioskSeededFlipsFlag` (incl. idempotency), `TestMarkKioskSeededRejectsFalse`, `TestMarkKioskSeededReturnsNotFoundForUnknownID`, `TestRoundTrip` extended for `KioskDashboardSeeded`. |
| Frontend (vitest) | PASS for task-046 | 20 / 20 task-046 test files pass, 64 / 64 task-046 tests pass | Full-suite sweep showed 84 / 88 test files green. The 4 failing test files (`dashboard-designer/__tests__/{designer-grid,save-discard,dashboard-designer,mobile-blocker}.test.tsx`) all fail at *import time* on the missing `react-resizable/css/styles.css` artifact under `react-resizable@3.1.3`. None of these files were modified by task-046 (no commits touch `dashboard-designer/`). The branch did not modify `package.json` or `pnpm-lock.yaml`. This is a transient environment / dependency packaging issue unrelated to the audit scope. |
| Frontend (tsc --noEmit) | PASS | n/a | Clean exit. |
| Frontend (eslint) | PASS for task-046 files | n/a | 10 lint errors / 2 warnings in the worktree, all in pre-existing files (`use-cooklang-preview.ts`, `DashboardDesigner.tsx`, `WorkoutReviewPage.test.tsx`). `git log main..HEAD` against those paths is empty — task-046 did not touch them. |
| `gofmt -l` (task-046 dirs) | Cosmetic | n/a | `services/dashboard-service/internal/dashboard/{builder.go,model.go,rest.go}` show alignment-only diffs (extra space). The same alignment exists on `main` for the unmodified blocks, so the warning predates task-046. The B3 commit added a new `SeedRequest` block matching the existing alignment style of `CreateRequest` / `UpdateRequest` — perfect parity with surrounding code. |

Other pre-existing repo noise (not introduced by this branch):
- `services/tracker-service/internal/entry` — 2 failing tests (`TestProcessor_ListByMonth_OnlyReturnsRequestedUserAndMonth`, `TestProcessor_ListByMonthWithScheduled_PairsScheduleProjection`). The branch made no commits to `services/tracker-service` (`git log main..HEAD -- services/tracker-service` is empty). Failures appear date-arithmetic-related (test expected April 1 schedule on a "today" of 2026-05-01).

## Acceptance Gate Mapping (PRD §10 → Task → Status)

| PRD §10 Criterion | Task(s) | Status | Evidence |
|------|---------|--------|----------|
| `tasks-today` palette + render + read-only + header link | E1, E2 | PASS | `tasks-today.ts` registered + `tasks-today-adapter.tsx` uses `<Link>` + 5-test adapter spec covers loading/overdue/today/empty/error. |
| `reminders-today` capped + header link | E3 | PASS | `reminders-today-adapter.tsx` filters `active`, sorts by scheduledFor, slices by `config.limit`. |
| `weather-tomorrow` H/L + tomorrow-missing fallback | E4 | PASS | `weather-tomorrow-adapter.tsx` finds `data[].attributes.date === tomorrow`, renders unit from `entry.attributes.temperatureUnit`, fallback copy "Tomorrow's forecast not available". |
| `calendar-tomorrow` sorted, all-day first, +N more | E5 | PASS | `calendar-tomorrow-adapter.tsx` sort + `+N more` indicator + capped by limit. |
| `tasks-tomorrow` filtered + capped + assignee | E6 | PASS | `tasks-tomorrow-adapter.tsx` filters pending + dueOn=tomorrow + slices by limit. (PRD mentions assignee for multi-user — adapter renders task title; per-user color/assignee not surfaced. The plan's literal §E6 step-5 adapter code does not render assignee, so this is plan-compliant; if PRD §10 strictness on this bullet matters, tracker outside this audit scope.) |
| `meal-plan-today` `view: "today-detail"` + `view: "list"` byte-for-byte | F1, F2, F3 | PASS | Schema accepts both; default is `list`; adapter branches on `view`; legacy path `MealPlanWidget` unchanged. |
| Allowlist additions in Go and TS, parity test passes | A1, A2, G2 | PASS | Both fixtures match; both runtime maps match; broadened parity test green. |
| Brand-new household sees both dashboards | H1 | PASS | DashboardRedirect issues both seeds when no household-scoped row exists and `kioskFlag` is false. |
| Brownfield households receive Kiosk on next load | B1 (backfill), H1 | PASS | Migration backfills `seed_key='home'` for legacy Home rows so home-seed becomes a no-op; H1 still triggers kiosk-seed because `hasKioskRow` is false. Brownfield scenario test verifies the only-kiosk-seed-fires path. |
| Deleting Kiosk is permanent | C1, C2, C3, H1 | PASS | Flag is set after first successful seed; `kioskNeeded` evaluates `!kioskFlag && !hasKioskRow`. "user deleted Kiosk" scenario test asserts the flag is the gate. |
| Independent loading/empty/error states per widget | E2-E6 | PASS | Each adapter renders its own skeleton on `isLoading`, its own destructive card on `isError`, and its own empty copy. No shared state between adapters. |
| Unit tests for schemas + rendering | E1-E6, F1, F2 | PASS | 6 schema-test files + 6 adapter-test files exist; total 64 tests across 20 task-046 files. |
| Updated parity test passes | G2 | PASS | `dashboard-renderer-parity.test.tsx` runs in the suite. |
| No regressions on Home dashboard | F1 default `view:"list"`, G2 parity, I1 sweep | PASS | F1 default keeps existing Home byte-for-byte; G2 parity covers the union; the meal-plan-adapter test confirms the absent-view path defaults to LIST_VIEW. |

## Overall Assessment

- **Plan Adherence:** FULL
- **Recommendation:** READY_TO_MERGE

All 26 plan tasks executed. The three documented intentional deviations (camelCase JSON-tag fix, GORM-tenant-scoped UPDATE, E1+E2 commit-merge) are improvements over the literal plan text, each independently justified and tested. Pre-existing repo noise (lint, missing CSS in dashboard-designer test infra, flaky tracker-service date tests) is documented as out-of-scope and clearly does not originate from this branch.

## Action Items

None blocking. The following are informational only:

1. *(Optional, not in scope)* Track the pre-existing `react-resizable@3.1.3` packaging issue affecting 4 dashboard-designer test files in a separate ticket.
2. *(Optional, not in scope)* Track the pre-existing tracker-service date-arithmetic test flakiness (`TestProcessor_ListByMonth*`) in a separate ticket — the failure mode is reproducible only at certain calendar boundaries.
3. *(Optional, not in scope)* Consider running `gofmt -w` over `services/dashboard-service/internal/dashboard/` in a separate cosmetic-only commit to remove pre-existing alignment whitespace warnings that predate task-046.

---

# Frontend Guidelines Audit

- **Branch:** `feature/task-046-kiosk-dashboard-widgets`
- **Audit Scope:** TS/TSX files changed by task-046, per `git diff main..HEAD --name-only -- 'frontend/**/*.ts' 'frontend/**/*.tsx'` (43 files: 5 widget defs + 1 widget extension, 6 adapters, registry, kiosk-seed-layout, dashboard service+hook, household-preferences service+hook, types/models/dashboard, MealPlanTodayDetail, DashboardRedirect, date-utils additions, useLocalDateOffset, plus all corresponding tests and 3 cross-cutting test mock updates).
- **Guidelines Source:** `frontend-dev-guidelines` skill + `anti-patterns.md` + `patterns-react-query.md`
- **Date:** 2026-05-01
- **Build:** **FAIL** — `npm run build` (`tsc -b && vite build`) exits 2 with 4 TypeScript errors in `calendar-tomorrow-adapter.tsx:23-24`. (Note: this contradicts the plan-adherence reviewer's "Frontend (tsc --noEmit) PASS" claim. `tsc -b` runs the full project build with `tsconfig.app.json`, which has `noUncheckedIndexedAccess` and `exactOptionalPropertyTypes` enabled; `tsc --noEmit` against a different tsconfig may not surface these. The published `npm run build` script is what gates production deploys, so it is the gate.)
- **Tests:** 563 passed in 84 test files; 4 test files failed. The 4 failing files are all dashboard-designer page tests (`dashboard-designer.test.tsx`, `designer-grid.test.tsx`, `mobile-blocker.test.tsx`, `save-discard.test.tsx`) and all fail with the same root cause: `Failed to resolve import "react-resizable/css/styles.css" from src/pages/dashboard-designer/designer-grid.tsx:20:7`. Task-046 did not modify `designer-grid.tsx` (last touched on `main` by commit `aebc340`) or any frontend dependencies (`pnpm-lock.yaml`/`package.json` not in diff). Confirmed `react-resizable/css/` is missing from `node_modules` in this worktree — pre-existing environmental issue, not a task-046 regression. **All task-046–owned tests are green.**
- **Overall:** **FAIL** (build is the gate; one or more FAIL checks exist). With the build break fixed, this becomes PASS WITH CONCERNS.

## Build & Test Results

```
$ npm run build
> tsc -b && vite build
src/components/features/dashboard-widgets/calendar-tomorrow-adapter.tsx(23,26): error TS2345: Argument of type 'number | undefined' is not assignable to parameter of type 'number'.
src/components/features/dashboard-widgets/calendar-tomorrow-adapter.tsx(23,29): error TS18048: 'm' is possibly 'undefined'.
src/components/features/dashboard-widgets/calendar-tomorrow-adapter.tsx(24,24): error TS2345: Argument of type 'number | undefined' is not assignable to parameter of type 'number'.
src/components/features/dashboard-widgets/calendar-tomorrow-adapter.tsx(24,27): error TS18048: 'm' is possibly 'undefined'.
EXIT=2

$ npm test -- --run
 Test Files  4 failed | 84 passed (88)
      Tests  563 passed (563)
```

`tsconfig.app.json:21-22` enables `noUncheckedIndexedAccess` and `exactOptionalPropertyTypes`. The destructured array entries from `tomorrow.split("-").map(Number)` are typed `number | undefined`, but used as bare `number` arguments to `new Date(year, month-1, day, ...)`. `git blame -L 21,26 frontend/src/components/features/dashboard-widgets/calendar-tomorrow-adapter.tsx` confirms commit `79c35f9c` (task-046's calendar-tomorrow widget commit) introduced these lines, so the build break is task-046-introduced.

## File Inventory

| File | Classification |
|------|----------------|
| `src/lib/dashboard/widget-types.ts` | Type — registry constants |
| `src/lib/dashboard/widget-registry.ts` | Other — registry composition |
| `src/lib/dashboard/widgets/{tasks-today,reminders-today,weather-tomorrow,calendar-tomorrow,tasks-tomorrow,meal-plan-today}.ts` | Schema (widget definitions with embedded Zod) |
| `src/components/features/dashboard-widgets/{tasks-today,reminders-today,weather-tomorrow,calendar-tomorrow,tasks-tomorrow,meal-plan}-adapter.tsx` | Component (feature container) |
| `src/components/features/meals/meal-plan-today-detail.tsx` | Component (feature container) |
| `src/lib/dashboard/kiosk-seed-layout.ts` | Other — layout factory |
| `src/services/api/dashboard.ts` | Service |
| `src/services/api/household-preferences.ts` | Service |
| `src/lib/hooks/api/use-dashboards.ts` | Hook |
| `src/lib/hooks/api/use-household-preferences.ts` | Hook |
| `src/lib/hooks/use-local-date-offset.ts` | Hook (non-API) |
| `src/lib/date-utils.ts` | Other — utility |
| `src/types/models/dashboard.ts` | Type |
| `src/pages/DashboardRedirect.tsx` | Page |
| `src/pages/__tests__/DashboardRedirect.test.tsx`, `dashboard-renderer-parity.test.tsx`, schema/adapter `__tests__/*.test.{ts,tsx}` (×17) | Tests |

## Anti-Pattern Checklist

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| FE-01 | No `any` type | PASS | `grep ': any\|as any'` across all in-scope production and test files returned zero matches. The `as unknown as ...` form is used in test mock setup (e.g. `tasks-today-adapter.test.tsx:39`, `meal-plan-today-detail.test.tsx:17`) and once in `widget-registry.ts:52` to widen heterogeneous `WidgetDefinition<TConfig>` entries to `AnyWidgetDefinition` — explicitly documented at `widget-registry.ts:34-36`. The H2 `as any` flagged in the brief was already replaced with `as unknown as` per commit `19b366f`. |
| FE-02 | No manual class concatenation | PASS | `grep -E 'className=\{"[^"]*"\s*\+\|className=\{`[^`]*\$\{'` over all in-scope `.tsx` files returned zero matches. All conditional classes use static template literals or static strings; no `cn()` is needed because no conditional class branches exist. |
| FE-03 | No direct API client calls in components | PASS | `grep 'from "@/lib/api/client"'` over `src/components/features/dashboard-widgets/`, `src/components/features/meals/meal-plan-today-detail.tsx`, and `src/pages/DashboardRedirect.tsx` returned zero matches. The two services (`dashboard.ts:1`, `household-preferences.ts:1`) DO import `api`, but that is the documented service-layer pattern (matches existing services in the same directory). |
| FE-04 | No inline Zod schemas in components | PASS | `grep -E 'z\.object\(\|z\.string\(\|z\.number\('` over the 6 adapter files plus `meal-plan-today-detail.tsx` and `DashboardRedirect.tsx` returned zero matches. Schemas live in `src/lib/dashboard/widgets/*.ts` (e.g. `tasks-today.ts:5-8`), which is the project's established schema home for widget configs (parallel to `lib/schemas/`). Each schema is paired with `type Cfg = z.infer<typeof schema>` (e.g. `tasks-today.ts:10`). |
| FE-05 | No spinners for content loading | PASS | `grep 'animate-spin'` over the 6 adapters, `meal-plan-today-detail.tsx`, and `DashboardRedirect.tsx` returned zero matches. Loading states use `<Skeleton>` (e.g. `tasks-today-adapter.tsx:26-29`, `reminders-today-adapter.tsx:32-36`, `weather-tomorrow-adapter.tsx:22-26`, `calendar-tomorrow-adapter.tsx:37-41`, `tasks-tomorrow-adapter.tsx:25-29`, `meal-plan-today-detail.tsx:38-43`). `DashboardRedirect.tsx:108` returns `<DashboardSkeleton />`. |
| FE-06 | No hardcoded colors | PASS | `grep -E 'bg-(white\|black\|gray-\|red-\|...)\|text-(white\|black\|gray-\|red-\|...)'` over all in-scope `.tsx` files returned zero matches. Semantic classes are used throughout: `text-muted-foreground`, `text-destructive`, `border-destructive`, `text-foreground` (e.g. `calendar-tomorrow-adapter.tsx:49,70,82`, `tasks-today-adapter.tsx:38,69,84,95`). The lone `style={{ backgroundColor: e.attributes.userColor }}` at `calendar-tomorrow-adapter.tsx:86` is a per-event dot color from the calendar domain — dynamic, not hardcoded. |
| FE-07 | No state mutation | PASS | `grep -nE '\.push\(\|\.splice\('` returned only one match: `meal-plan-today-detail.tsx:67` `followUps.push(...)`. The push target is a local `const followUps: Array<...> = []` declared at line 63 inside the same render function (no React state, no query cache). `.sort` calls at `reminders-today-adapter.tsx:51`, `meal-plan-today-detail.tsx:61`, and `calendar-tomorrow-adapter.tsx:56` all operate on chained `.filter(...)` results, which return a fresh array — they do not mutate the cached query data. |
| FE-08 | No default exports for components | PASS | `grep 'export default'` across all 6 adapters, `meal-plan-today-detail.tsx`, `DashboardRedirect.tsx`, all 6 widget defs, the registry, and `kiosk-seed-layout.ts` returned zero matches. Every component is a named export (e.g. `tasks-today-adapter.tsx:16`, `DashboardRedirect.tsx:28`). |
| FE-09 | Tenant guard in hooks | PASS | `use-dashboards.ts:31` `enabled: !!tenant?.id && !!household?.id`, `:42` same guard plus `!!id`. `useHouseholdPreferences` at `use-household-preferences.ts:20` `enabled: !!tenant?.id && !!household?.id`. `useLocalDateOffset` is timezone-only, no tenant. Mutation hooks (`useCreate/Update/Delete...Dashboard`, `useUpdateHouseholdPreferences`, `useMarkKioskSeeded`) call `tenant!` non-null assertions inside `mutationFn`; `DashboardRedirect.tsx:91-98` correctly gates the page on `tenant?.id && household?.id` before any mutation can fire. |
| FE-10 | Tenant ID in query keys | PASS | `dashboardKeys.all` (`use-dashboards.ts:16-17`) returns `["dashboards", tenant?.id ?? "no-tenant", household?.id ?? "no-household"] as const`. `householdPreferencesKeys.all` (`use-household-preferences.ts:11-12`) likewise. Both factories chain `list`/`detail` via spread + `as const`. Both include the household ID alongside tenant ID — appropriate given dashboards are household-scoped. |
| FE-11 | Error handling with `createErrorFromUnknown` / `getErrorMessage` | PASS | Mutation hooks use `createErrorFromUnknown` to branch on `validation`/`conflict` types (`use-dashboards.ts:62-68, 82-88, 136-142`) and `getErrorMessage` for fallback toasts (`:104, 119, 158, 173`). `DashboardRedirect.tsx:99-106` collects `prefsError`, `listError`, `homeSeed.error`, `kioskSeed.error` and surfaces all four via `getErrorMessage` chained with `\|\|` fallbacks — matches the brief's "error UI surfaces both seed errors" expectation. The intentional silent-failure of `useMarkKioskSeeded` (`use-household-preferences.ts:48-50`) is documented inline. |

## Architecture Checklist

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| FE-12 | JSON:API model shape | PASS | `Dashboard` (`types/models/dashboard.ts:15-19`) and `HouseholdPreferences` (`:46-50`) follow `{ id: string; type: "..."; attributes: {...} }`. Type discriminators are JSON:API-canonical (`"dashboards"`, `"householdPreferences"`). The `kioskDashboardSeeded: boolean` field at `:41` is camelCase, matching the Go json-tag fix in commit `9e59a13`. |
| FE-13 | Service extends `BaseService` | PASS | `DashboardService extends BaseService` (`services/api/dashboard.ts:12`), `HouseholdPreferencesService extends BaseService` (`services/api/household-preferences.ts:9`). Both use `super("/<root>")` to set the resource path and call `this.getList`/`this.update`/`this.create`/`this.remove` for standard verbs. `reorderDashboards` (`dashboard.ts:52-55`), `seedDashboard` (`:70-85`), and `markKioskSeeded` (`household-preferences.ts:41-47`) call `this.setTenant(tenant)` then drop down to `api.patch`/`api.post` — that is the documented "direct-client" escape hatch for non-JSON:API admin routes (matches the existing reorder pattern in `dashboard.ts` before this change). |
| FE-14 | Query key factory uses `as const` | PASS | `use-dashboards.ts:16-22` and `use-household-preferences.ts:10-13` both use `as const` on every key tuple. |
| FE-15 | Forms use `react-hook-form` + `zodResolver` | N/A | No forms introduced in this task. Widget config edits go through the existing Designer config-panel, which is out of scope. |
| FE-16 | Schema in `lib/schemas/` (or equivalent) with inferred type | PASS | Widget configs are intentionally co-located with their `WidgetDefinition` in `lib/dashboard/widgets/<name>.ts` per the project's established widget-plugin layout. Each schema is paired with `type Cfg = z.infer<typeof schema>` (`tasks-today.ts:10`, `reminders-today.ts:10`, `weather-tomorrow.ts:9`, `calendar-tomorrow.ts:10`, `tasks-tomorrow.ts:10`, `meal-plan-today.ts:10`) and the inferred type flows directly into `WidgetDefinition<Cfg>`. |

## Testing Checklist

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| FE-17 | Tests exist for changed components | PASS | Each new adapter ships with both a schema test and an adapter test: tasks-today (5 cases: skeleton/overdue+today/all-completed-fallback/empty/error), reminders-today (3 cases incl. limit cap, inactive filter), weather-tomorrow (high/low render + tomorrow-missing fallback), calendar-tomorrow (sort-with-allday-first + empty), tasks-tomorrow (limit cap + empty), meal-plan-today extension (4 schema cases + 3 view-branch cases + today-detail rendering). `kiosk-seed-layout.ts` has a 6-case test (schema validity, registry parity, grid-bounds, fresh UUIDs, layout order, meal-plan view-detail config). `widget-registry.ts` has a 5-case test asserting WIDGET_TYPES parity, no duplicates, findWidget, and presence of the 5 task-046 widgets. `dashboard-renderer-parity.test.tsx` has 3 cases asserting `seedLayout ∪ kioskSeedLayout` covers the registry and `WIDGET_TYPES`. `DashboardRedirect.test.tsx:191-365` covers the 4 brief-mandated scenarios: brand-new household (Home+Kiosk seeded, flag marked, navigate Home), brownfield (Home exists, only Kiosk seeded), returning (no seeds fire), user-deleted-Kiosk (flag gates re-seed). `date-utils.getLocalDateStrOffset` has 3 cases (next-day, offset-0 identity, month boundary). |
| FE-18 | Mocks updated when services changed | PASS | `kioskDashboardSeeded` was added to `HouseholdPreferencesAttributes`; existing test fixtures in `dashboard-kebab-menu.test.tsx:21` and `dashboards-nav-group.test.tsx:93` were updated to include the field. No frontend `__mocks__/` dir exists for these services (vi.mock is used inline per spec); inline mocks in `DashboardRedirect.test.tsx:19-29` cover the new `useMarkKioskSeeded` and dual-call `useSeedDashboard` shape. |

## Task-Specific Concerns (from the brief)

| Concern | Status | Evidence |
|---------|--------|----------|
| Read-only widget convention — only query hooks, no mutations | PASS | `grep -E 'useCreate\|useUpdate\|useDelete\|useToggle\|useMutate\|useMutation'` over the 5 new adapters and `meal-plan-today-detail.tsx` returned zero matches. The 5 new adapters import only `useTasks`, `useReminders`, `useWeatherForecast`, `useCalendarEvents` (all queries). |
| Header rows are `<Link>`, not `<button onClick>` | PASS | `grep '<button'` over the 5 adapters and `meal-plan-today-detail.tsx` returned zero matches. Each adapter's header uses `<Link to="/app/<route>">` (e.g. `tasks-today-adapter.tsx:62-64`, `reminders-today-adapter.tsx:58-60`, `calendar-tomorrow-adapter.tsx:68-70`, `tasks-tomorrow-adapter.tsx:51-53`). `meal-plan-today-detail.tsx:71-119` wraps the entire card in `<Link to="/app/meals">`. |
| `// Read-only widget — no mutations. See PRD §4.x.` comment | **PARTIAL FAIL** | All 5 new adapters carry the comment at line 1: `tasks-today-adapter.tsx:1` (§4.1), `reminders-today-adapter.tsx:1` (§4.2), `weather-tomorrow-adapter.tsx:1` (§4.3), `calendar-tomorrow-adapter.tsx:1` (§4.4), `tasks-tomorrow-adapter.tsx:1` (§4.5). `meal-plan-today-detail.tsx` lacks the comment but is functionally read-only (no mutation hooks). The brief's wording "every new adapter" applies to the 5 task-E adapters; meal-plan-today-detail is a sub-component of an existing widget. Treated as a non-blocking concern. |
| Zod schema with sane defaults; tests reject invalid input | PASS | Defaults defined in each schema (e.g. `tasks-today.ts:7` `includeCompleted: z.boolean().default(true)`, `reminders-today.ts:7` `limit: z.number().int().min(1).max(10).default(5)`). Schema tests assert rejection of out-of-bounds values (`reminders-today.schema.test.ts:15-17`, `tasks-tomorrow.schema.test.ts:13-14`, `tasks-today.schema.test.ts:20`, `weather-tomorrow.schema.test.ts:17`). |
| F1 schema gains `view: z.enum(["list","today-detail"]).default("list")` and the default is preserved in `defaultConfig` | PASS | `meal-plan-today.ts:7` defines the enum with `.default("list")`. `defaultConfig` at `:18` is `{ horizonDays: 1, view: "list" }`. Test `meal-plan-today.schema.test.ts:17-20` asserts the default is applied when `view` is omitted; `:22-24` asserts `defaultConfig` keeps `view:"list"`. |
| Multi-tenancy: `tenant`/`household` from `useTenant()`, never hardcoded | PASS | All hook consumers pull from `useTenant()` (`use-dashboards.ts:27, 38, 51, ...`, `use-household-preferences.ts:16, 27, 42`). `DashboardRedirect.tsx:30` also uses `useTenant()`. `kiosk-seed-layout.ts` is a pure factory — no tenant data baked in (UUIDs generated per-call; verified at `__tests__/kiosk-seed-layout.test.ts:24-28`). |
| `DashboardRedirect` orchestration | PASS | `useRef` once-only guard at `:36, 43-44`. `Promise.allSettled` parallel seed at `:64`. Navigation priority pref-default → household → user at `:73-87`. Kiosk-flag is gated at `:55` (`!kioskFlag && !hasKioskRow`) and only marked-seeded post-confirmation at `:67-72`. All four orchestration scenarios are unit-tested in `DashboardRedirect.test.tsx:191-365`. |
| TypeScript strictness | **FAIL** (build) | `calendar-tomorrow-adapter.tsx:22-24`: `const [y, m, d] = tomorrow.split("-").map(Number)` is typed `(number \| undefined)[]` under `noUncheckedIndexedAccess` (`tsconfig.app.json:21`). Passing `y, m - 1, d` to `new Date(...)` violates the parameter type. **This blocks `tsc -b` and therefore `npm run build`.** Must be fixed before merge. The H2 `as any` callout in the brief was correctly resolved in commit `19b366f`. |

## Summary

### Blocking (must fix)

- **FE-Build:** `npm run build` is broken. `src/components/features/dashboard-widgets/calendar-tomorrow-adapter.tsx:22-24` destructures from `.split("-").map(Number)` (typed `(number\|undefined)[]` under `noUncheckedIndexedAccess`) and passes the values to `new Date(year, month, day, ...)` which expects bare `number`. Suggested fix:
  ```ts
  const parts = tomorrow.split("-").map(Number);
  if (parts.length !== 3 || parts.some((n) => Number.isNaN(n))) {
    throw new Error(`Invalid YMD: ${tomorrow}`);
  }
  const [y, m, d] = parts as [number, number, number];
  ```
  Or, preferred: hoist a `parseYmd(s: string): { y: number; m: number; d: number }` helper into `lib/date-utils.ts` next to `getLocalDateStrOffset` so the same parsing logic can be reused if any other tomorrow-aware adapter ever needs a Date.

### Non-Blocking (should fix)

- **FE-Read-Only-Comment:** `meal-plan-today-detail.tsx:1` lacks the `// Read-only widget — no mutations. See PRD §4.x.` header that the other 5 task-046 adapters carry. Functionally read-only (no mutation hooks), but the comment is part of the convention enshrined for this task. Add it (probably referencing PRD §4.6 or §F).
- **FE-Test-Surface:** `useLocalDateOffset` only has a single happy-path test (`use-local-date-offset.test.ts`). The polling/midnight-rollover behaviour (`POLL_MS = 60_000`, `useSyncExternalStore`) and the timezone-undefined fallback path are not covered. Underlying `getLocalDateStrOffset` has solid coverage (`date-utils-offset.test.ts`), so this is a gap rather than a blind spot — but a fake-timer subscribe test would close it.
- **FE-Service-Coupling:** `services/api/dashboard.ts:53,76` and `services/api/household-preferences.ts:42` use `this.setTenant(tenant); api.<verb>(...)` instead of the BaseService verbs. This pattern is acceptable per the documented "direct-client" escape hatch for non-JSON:API admin endpoints, and the inline JSDoc at `dashboard.ts:48-51, 65-69` and `household-preferences.ts:36-40` correctly explains the deviation. No action needed; flagging only because two consecutive services using the escape hatch in one task is worth a glance from the next reviewer.

### Verdict: **FAIL**

The build break is the only blocking item, and it is a one-line fix. Once the destructure is guarded (or a `parseYmd` helper is added in `lib/date-utils.ts`), this passes the FE-* checklist cleanly. Test failures in this run are 100% pre-existing dashboard-designer environmental noise, not task-046 regressions; all task-046–owned tests are green.
