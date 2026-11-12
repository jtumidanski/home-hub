
## Frontend guidelines

- **Verdict:** APPROVED_WITH_FINDINGS
- **Scope:** 23 .ts/.tsx files under frontend/ changed vs main (pages, use-query-param-state hook, calendar-url, date-utils, app-routes, redirect-preserving-query, tests)
- **Build (`npx tsc -b`):** PASS. **Lint:** 0 errors, 1 pre-existing warning (new-dashboard-modal.tsx:80, out of scope). **Tests:** 175 files, 1120 passed, 0 failed.

| ID | Status | Evidence |
|----|--------|----------|
| FE-01 no any | PASS | grep `: any\|as any\|<any>` over all scoped files: 0 matches |
| FE-02 class concat | WARN | frontend/src/pages/TrackerPage.tsx:63 template-literal `` `mx-auto ${widthClass}` `` (widthClass is a static constant, TrackerPage.tsx:46); should be `cn()` |
| FE-03 no API client in UI | PASS | 0 matches for `lib/api/client` |
| FE-04 no inline Zod | PASS | 0 matches |
| FE-05 spinners | PASS | 0 `animate-spin` |
| FE-06 hardcoded colors | PASS | 0 matches |
| FE-07 mutation | PASS | only `days.push` on a local array in MealsPage.tsx:42, not state; no setState mutation |
| FE-08 default exports | PASS | 0 `export default` |
| FE-09/10 tenant hooks, query keys | N/A | no files in lib/hooks/api, services, or query-key factories changed |
| FE-11 error handling | PASS | IngredientsPage.tsx:94,103 use `createErrorFromUnknown` + toast; CalendarPage.tsx:252 catch (unchanged by this task) uses `getErrorMessage` + toast |
| FE-12/13/16 models, services, schemas | N/A | none changed |
| FE-14 `as const` | N/A | no query-key factory changed |
| FE-15 forms | N/A | no form changes |
| FE-17 tests | PASS | tests for each changed unit: use-query-param-state.test.tsx, calendar-url.test.ts, date-utils.test.ts, redirect-preserving-query.test.tsx, app-routes.test.tsx, nav-config.test.ts, and one test per changed page |
| FE-18 mocks | N/A | no services changed |
| FE-19 cursor-pointer | PASS | new onClick sites in IngredientsPage.tsx:128,137, ShoppingListsPage.tsx:111,123, TrackerPage.tsx view buttons are `Button` components, not bare divs; IngredientsPage.tsx:212 has cursor-pointer |

### Blocking
- None.

### Non-blocking
- FE-02: frontend/src/pages/TrackerPage.tsx:63 use `cn("mx-auto", widthClass)` (or inline the literal `max-w-3xl`).
- Observation: frontend/src/lib/hooks/use-query-param-state.ts:41 has an `eslint-disable react-hooks/exhaustive-deps` with a justifying comment (codec identity changes per render); acceptable but worth a second look.

## Plan adherence

**Verdict: APPROVED_WITH_FINDINGS** (no blocking items)

Plan: docs/tasks/task-058-url-addressable-views/plan.md. Range: merge-base(main)..HEAD. Plan checkboxes are all unticked in plan.md (progress.md ledger is the tracking record); 8 implementation tasks + Task 9 gate.

| # | Task | Status | Evidence |
|---|------|--------|----------|
| 1 | Local date-string helpers | DONE | frontend/src/lib/date-utils.ts, tests in lib/__tests__/date-utils.test.ts; MealsPage private formatDateStr removed, uses formatLocalDateStr |
| 2 | useQueryParamState + withParam | DONE | frontend/src/lib/hooks/use-query-param-state.ts (read-in-render, replace writes, omit-current, canonicalize effect keyed on raw) + test file |
| 3 | AppRoutes extraction, RedirectPreservingQuery, probe, nav guard | DONE | src/app-routes.tsx, components/common/redirect-preserving-query.tsx, src/test/location-probe.tsx, nav-config.test.ts, app-routes.test.tsx; App.tsx slimmed |
| 4 | Grocery tab from route | DONE | ShoppingListsPage.tsx (tab prop, aria-pressed, navigate); app-routes.tsx archived before :id |
| 5 | Ingredients view from route | DONE | IngredientsPage.tsx; routes list/categories/bulk, :id kept, no ingredients/*; app-routes tests incl. ?category=abc and not-a-view |
| 6 | Habits view + month | DONE | TrackerPage.tsx monthCodec/goTo; habits routes + habits/* + tracker retargeted to /app/habits/today (app-routes.tsx); baseline test updated per Ruling 1 (app-routes.test.tsx:70-72) |
| 7 | Meals week | DONE | MealsPage.tsx weekCodec (Monday-normalizing via isoMondayOf); MealsPage.test.tsx covers default, canonical, Wednesday, 02-31/current via it.each, selector |
| 8 | Calendar view + date + OAuth | DONE | calendar-url.ts, CalendarPage.tsx (clearOAuthParams preserves date, line ~50/84/96), routes week/month/calendar/*; CalendarPage.test.tsx covers month normalize, week 1 toggle, day-click + Back, connected with/without date, auth_failed |
| 9 | Repo gate | DEFERRED to controller (Ruling 2) | branch-end flagless verify running separately; not run here |

Completion: 8/8 implementation tasks DONE; Task 9 covered by controller. Skipped without approval: 0. Partial: 0.

Global constraints spot-checked: no `key` props on new routes; changes confined to frontend/src; no nav-config/ShoppingListDetailPage/reauthorize-banner/connection-status edits in the diff; view toggles remain buttons with aria-pressed; param names month/date/week verbatim.

### Build and test
- `npx tsc -b`: PASS (exit 0).
- Task-relevant vitest (src/pages, src/__tests__, src/lib, components/common, calendar, navigation): 114 files / 825 tests PASS.
- Full `npx vitest run`: two runs under load (verify.sh running concurrently) showed 11 then 1 timeout-type failures in unrelated files (last: a recurrence-form "Ends radio" test, not touched by this branch). Judged load flakiness; the branch-end flagless verify is the authority.

### Deferred `minor` ledger lines: any plan gap?
None is a plan gap; all are test-strength or tidy-up notes. The closest to plan text: Task 8 "no page-level mobile CalendarPage test" - the plan's page tests specify desktop only and mobile anchor math is unit-tested in calendar-url.test.ts, so plan satisfied. Task 7 "exactly one rewrite" is asserted by final URL + REPLACE type, not navigation count; plan wording ("exactly one rewrite") is only indirectly pinned. Task 5 `?category=` round trip: plan requires list->categories->list keeping the query and bare redirect keeping it; both tested. Task 2/3/7 no-RED notes are process, not deliverable, gaps.

### Non-blocking observations
1. Branch diff also contains unrelated workflow tooling (tools/*.sh, .claude/commands/*, docs/*.md from commits 54cdcc4, 5621bbc, c366725 "claude-drive port"). Outside plan scope ("frontend only"); confirm intended to ride in this PR.
2. plan.md checkboxes were never ticked; progress.md is the record.
3. progress.md and go.work.sum are modified/uncommitted in the working tree (go.work.sum is gate side effect; revert before PR).
