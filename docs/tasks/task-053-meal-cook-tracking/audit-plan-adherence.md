# Plan Audit — task-053-meal-cook-tracking

**Plan Path:** docs/tasks/task-053-meal-cook-tracking/plan.md
**Audit Date:** 2026-06-14
**Branch:** task-053-meal-cook-tracking
**Base Branch:** main (f8aff3a)
**Implementation Range:** f8aff3a..a0b53fc

## Executive Summary

All 12 plan tasks were faithfully implemented. Every backend and frontend source change matches the plan's prescribed code essentially verbatim, all prescribed test files exist, and every spec requirement (FR-1..FR-18) in the coverage map is genuinely satisfied. Backend builds and tests pass, frontend type-check / full Vitest suite (TZ=UTC, 673 tests) / production build all pass, and the recipe-service Docker image builds successfully. No shared-library files changed. Verdict: READY_TO_MERGE.

The only nit: the plan's checkboxes are left unchecked (`- [ ]`) — purely cosmetic, since every step's artifact is committed and verified.

## Task Completion

| # | Task | Status | Evidence / Notes |
|---|------|--------|------------------|
| 1 | UsageSort type + parseUsageSort | DONE | provider.go:35-55 (type/consts/parser exactly as planned); usage_sort_test.go (5 cases incl. unknown→None). Commit 7900033. |
| 2 | Plan test fixtures + scope GetRecipeUsage by household | DONE | plan_test_fixtures_test.go (local testPlanWeek/testPlanItem + seed helpers, avoids import cycle); provider.go:202-222 (getRecipeUsageFromPlanItems joins plan_weeks, scopes tenant+household); processor.go:441-443 (3-arg GetRecipeUsage); resource.go:82,117 (tenant lookup + scoped call); usage_scope_test.go (asserts 2 of 2/3/4 counted). Commit b7deed1. |
| 3 | Frequency-sorted, paginated list query | DONE | provider.go:57-68 (ListFilters + TenantID/HouseholdID/UsageSort); provider.go:194-200 (recipeWithUsage scan struct); provider.go:70-175 (getAll: nil-usage default path preserved; sort path LEFT JOINs grouped derived table, COALESCE, order before LIMIT/OFFSET, title/id tie-breaker); processor.go:153-175 (List returns usageMap); resource.go:92,103,110-128 (parse sort, merge usageMap regardless of include_usage). frequency_sort_test.go: order, zero-count inclusion, tie-break across pages, search composition, default-order-unchanged. Commit d6105c8. |
| 4 | Verify recipe-service Docker build | DONE | `git diff --name-only main...HEAD -- shared/` empty (no shared changes). `docker build -f services/recipe-service/Dockerfile -t recipe-service:task-053 .` succeeded (image e5f45014). |
| 5 | Add RecipeSort type | DONE | recipe.ts:59 `export type RecipeSort = "usageCount" \| "-usageCount";`. Commit df2ff38. |
| 6 | Thread sort through recipe service | DONE | services/api/recipe.ts:10 (import), :24 (RecipeListParams.sort), :52 (`if (params?.sort) query.set("sort", params.sort)`; no include_usage sent). __tests__/recipe.test.ts (3 cases: desc/asc/omitted). Commit 96b79df. |
| 7 | Add sort to useRecipes hook params | DONE | use-recipes.ts:6 (RecipeSort import), :35 (UseRecipesParams.sort); params spread into queryKey at :41 so sort yields distinct cache entry. Commit 4b93540. |
| 8 | Shared RecipeSortSelect control | DONE | recipe-sort-select.tsx (DEFAULT_VALUE sentinel, aria-label "Sort recipes", Default/Most cooked/Least cooked items). __tests__/recipe-sort-select.test.tsx (3 cases incl. Default→undefined). Commit b7755be. |
| 9 | Wire sort into recipe selector | DONE | recipe-selector.tsx:10,12 (imports folded), :22 (sort state), :29 (passed to useRecipes), :69 (control rendered after classification Select). __tests__/recipe-selector.test.tsx (default undefined + most-cooked). Commit 80bca40. |
| 10 | Wire sort into Recipes page | DONE | RecipesPage.tsx:9,17 (imports), :25 (state), :36 (passed to useRecipes), :79-100 (search+sort flex row with RecipeSortSelect). __tests__/RecipesPage.test.tsx (least-cooked→usageCount). Commit 772c69f. |
| 11 | Show cook count on recipe card | DONE | recipe-card.tsx:93-97 (`{(attributes.usageCount ?? 0) > 0 && ... cooked {usageCount}x}` placed after total-time span). __tests__/recipe-card.test.tsx:103-115 (present>0 / =0 / absent). Commit a0b53fc. |
| 12 | Full frontend verification (TZ=UTC) | DONE | `npx tsc --noEmit` clean; `TZ=UTC npx vitest run` 673/673 pass; `npm run build` succeeds. |

**Completion Rate:** 12/12 tasks (100%)
**Skipped without approval:** 0
**Partial implementations:** 0

## Skipped / Deferred Tasks

None.

The plan explicitly documented (and the implementation honored) one intentional non-addition: no `httptest`-based handler test for FR-9's HTTP layer, because the recipe package has no HTTP test harness and the merge is thin reviewed glue over the tested `Processor.List`. This is a documented design decision in the plan, not a silent skip.

## Spec Coverage (FR-1..FR-18) — verified, not just claimed

- FR-1/FR-3 (count = COUNT(plan_items), computed at query time, no counter/migration): no entity or migration files in the diff; aggregation is a runtime `COUNT(*)` over `plan_items` (provider.go:138, :208).
- FR-2 (zero-count sortable/included): `COALESCE(u.usage_count, 0)` + LEFT JOIN (provider.go:148-153); TestListFrequencySortOrder asserts cherry=0 present and ordered.
- FR-4 (both directions): parseUsageSort asc/desc; dir flip at provider.go:143-146.
- FR-5 (order full set before pagination): Order applied before Offset/Limit (provider.go:155-157).
- FR-6 (composes with filters): search/tag/classification/plannerReady/normalization filters all applied to `query` before the join; TestListFrequencySortComposesWithSearch.
- FR-7 (deterministic tie-breaker stable across pages): `…, recipes.title ASC, recipes.id ASC` (provider.go:148); TestListFrequencySortTieBreakerAcrossPages across 2 pages.
- FR-8 (default order unchanged): UsageSortNone branch keeps `Order("created_at DESC")`, no join, nil usageMap (provider.go:123-132); TestListDefaultOrderUnchanged.
- FR-9 (usageCount present on sort without include_usage): handler merges whenever `usageMap != nil` (resource.go:121-128); getAll returns usageMap on sort path.
- FR-10/FR-14/FR-18 (selector + page controls, direction-explicit labels): RecipeSortSelect ("Most cooked"/"Least cooked") wired into both surfaces.
- FR-11 (selector "used Nx"): recipe-selector.tsx:100-102 lights up from the same usageMap source.
- FR-12 (selector re-queries full set): sort in queryKey (use-recipes.ts:41).
- FR-13/FR-16 (defaults unchanged, composes with page filters): sort defaults to undefined; added alongside existing filters.
- FR-15 (recipes page surfaces count): recipe-card.tsx:93-97.
- FR-17 (counts scoped to household): both the sort-path join (provider.go:140) and legacy path (provider.go:210) scope `pw.tenant_id AND pw.household_id`; TestGetRecipeUsageScopedToHousehold proves only the requesting household's 2 items count.

## Build & Test Results

| Service | Build | Tests | Notes |
|---------|-------|-------|-------|
| recipe-service (Go) | PASS | PASS | `go build ./...` clean; `go test ./internal/recipe/ -count=1` ok. |
| recipe-service (Docker) | PASS | n/a | image recipe-service:task-053 built. |
| frontend (tsc) | PASS | n/a | `npx tsc --noEmit` no errors. |
| frontend (vitest) | n/a | PASS | `TZ=UTC npx vitest run`: 102 files, 673 tests passed. |
| frontend (build) | PASS | n/a | `npm run build` succeeded (chunk-size warning only, pre-existing). |

Working-tree noise unrelated to source: `frontend/package-lock.json` and `go.work.sum` show as modified after running tooling; these are not source changes and not part of the committed implementation diff.

## Overall Assessment

- **Plan Adherence:** FULL
- **Recommendation:** READY_TO_MERGE

## Action Items

1. (Optional, cosmetic) Mark the plan.md checkboxes as `- [x]` to reflect completion; no functional impact.
