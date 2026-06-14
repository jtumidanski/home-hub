# Task 053 — Meal Cook Tracking & Frequency Sort — Context

Companion to `plan.md`. Captures the key files, decisions, dependencies, and gotchas an engineer needs before executing.

## What this task does

Adds an ascending/descending **cook-frequency sort** to the recipe list endpoint and exposes it in two UI surfaces (meal-planner recipe selector + Recipes page). The cook count = number of `plan_items` rows referencing a recipe (scheduling == cooking). It stays **computed on the fly** — no persisted counter, no migration, no backfill.

## Key files (verified against source on 2026-06-14)

### Backend — `services/recipe-service/internal/recipe/`
- `provider.go` — `getAll` (list query, currently `ORDER BY created_at DESC`, returns `([]Entity, int64, error)`); `getRecipeUsageFromPlanItems` (current aggregation, **unscoped** — no `plan_weeks` join); `ListFilters`; `recipeUsageResult`/`recipeUsageRow` structs.
- `processor.go` — `Processor.List` (calls `getAll`, returns `([]Model, int64, error)`); `GetRecipeUsage` (1-arg today); `BuildListEnrichments`.
- `resource.go` — `listHandler` parses query params, calls `proc.List`, conditionally enriches usage when `include_usage=true`, marshals JSON:API + meta.
- `rest.go` — `RestModel.UsageCount int64 json:"usageCount,omitempty"` and `LastUsedDate *string` already exist; `ListEnrichment` carries both.
- `entity.go` — `recipes` table has **both** `TenantId` and `HouseholdId`.

### Backend — sibling packages (read-only for this task)
- `planitem/entity.go` — `plan_items`: `Id`, `PlanWeekId`, `Day (date)`, `RecipeId`, … **no** tenant/household columns. Index `idx_plan_item_recipe` on `recipe_id`.
- `plan/entity.go` — `plan_weeks`: has `TenantId` + `HouseholdId`, composite index `idx_plan_week_tenant_household`.

### Frontend — `frontend/src/`
- `types/models/recipe.ts` — `RecipeListAttributes.usageCount?` and `lastUsedDate?` already typed.
- `services/api/recipe.ts` — `RecipeListParams` + `listRecipes` (builds query string; never sends `include_usage`).
- `lib/hooks/api/use-recipes.ts` — `UseRecipesParams`; `useRecipes` spreads `params` into the query key.
- `components/features/meals/recipe-selector.tsx` — already renders `used {usageCount}x` / `last used {date}` (dormant until usage is returned); has a classification `<Select>`; `pageSize: 50`, single page.
- `pages/RecipesPage.tsx` — badge filter row; `useRecipes` with default page/pageSize; no sort control.
- `components/features/recipes/recipe-card.tsx` — metadata row (total time + tags); no usage display.

## Architecture decisions (from design.md)

1. **LEFT JOIN on a grouped derived table** (`SELECT recipe_id, COUNT(*), MAX(day) … GROUP BY recipe_id`), `COALESCE(count,0)`, 1:1 join → count + last-used in one pass; recipe row count (and `total`) unaffected.
2. **Join only on the sort path.** No frequency sort → `getAll` keeps `ORDER BY created_at DESC` and adds no join; existing callers byte-for-byte unchanged.
3. **Unified count definition.** Both the sort-path join and the legacy `include_usage` path (`getRecipeUsageFromPlanItems`) join `plan_weeks` and scope by `tenant_id` AND `household_id`.
4. **Sort token:** `usageCount` (asc) / `-usageCount` (desc), matching the existing JSON:API attribute and the FE type. Unknown values → default order (lenient).
5. **Household scoping (FR-17):** scope counts by `plan_weeks.tenant_id` AND `plan_weeks.household_id`. The recipe *list* is scoped by `tenant_id` only (the tenant callback ignores household), so within a shared tenant a recipe could show `usageCount: 0` to a non-owning household — this is the literal FR-17 behavior and strictly improves on today's unscoped aggregation. Dropping the `household_id` predicate (one line in two queries) would make it tenant-wide if product later wants that.
6. **Frontend:** shared `RecipeSortSelect` (labels "Default / Most cooked / Least cooked"); sort state is component-local and resets to Default each mount; card shows `cooked Nx` when `usageCount > 0` (naturally only when sort is active, since the endpoint returns the count only on the sort path).

## Signature changes (each has exactly one caller — verified)

| Symbol | Before | After | Caller updated |
| --- | --- | --- | --- |
| `getAll` | `func(ListFilters) func(*gorm.DB) ([]Entity, int64, error)` | `… ([]Entity, map[uuid.UUID]recipeUsageResult, int64, error)` | `Processor.List` (processor.go:161) |
| `Processor.List` | `(ListFilters) ([]Model, int64, error)` | `(ListFilters) ([]Model, map[uuid.UUID]recipeUsageResult, int64, error)` | `listHandler` (resource.go:99) |
| `getRecipeUsageFromPlanItems` | `(db, recipeIDs)` | `(db, recipeIDs, tenantID, householdID)` | `GetRecipeUsage` |
| `Processor.GetRecipeUsage` | `(recipeIDs)` | `(recipeIDs, tenantID, householdID)` | `listHandler` (resource.go:114) |

> `planitem.Processor.GetRecipeUsage` is a **different** method in a different package — do not touch it.

## Gotchas / non-obvious constraints

- **Import cycle:** `plan/processor.go` imports `recipe`. The recipe internal test package therefore **cannot** import `plan` or `planitem`. Tests declare minimal local `plan_weeks`/`plan_items` structs (`plan_test_fixtures_test.go`) instead.
- **Test DB is SQLite in-memory.** `setupTestDB` only migrates recipe-package entities; frequency-sort tests must also migrate the local plan structs (`migratePlanTables`). The `LEFT JOIN (subquery)` and `MAX(date)` must work under SQLite — the test run is the first place a dialect issue would surface.
- **Tenant callback** (`shared/go/database`) injects `WHERE tenant_id = ?` only on GORM model queries (`db.Model(&Entity{})`). The aggregation uses raw `db.Table("plan_items AS pi")`, so the callback does **not** scope it — that's why the join must scope `plan_weeks` explicitly, and why `ListFilters` carries `TenantID`/`HouseholdID`.
- **Column ambiguity:** in the joined query only `recipes` has `tenant_id`/`title`/`servings`/`deleted_at` and only the derived table `u` exposes `usage_count`/`last_used_day`/`recipe_id`, so bare column refs are unambiguous. Qualify `recipes.<col>` defensively only if a DB complains (SQLite test will catch it).
- **Radix `<Select>` in tests:** works in this repo (see `new-dashboard-modal.test.tsx`): `user.click(getByRole("combobox"))` then `user.click(findByRole("option", { name }))`. The selector has two combos, so `RecipeSortSelect` sets `aria-label="Sort recipes"` and tests target `getByRole("combobox", { name: /sort/i })`.
- **No empty-string `SelectItem`:** `RecipeSortSelect` uses a `"default"` sentinel for the no-sort option to avoid Radix's empty-value restriction.
- **TZ=UTC:** CI runs UTC, dev machine is EDT. Run `TZ=UTC npx vitest run` before pushing (memory: verify-frontend-tz-tests-under-utc).
- **Run tests, not just builds**, before claiming completion (memory: run-tests-before-commit).

## Dependencies / scope boundaries

- Only `recipe-service` (Go) and `frontend` (TS) change. No other services. No plan-item write-path changes. No shared-library changes expected → only `recipe-service` needs a Docker rebuild.
- Indexes already cover the join (`idx_plan_item_recipe`, `idx_plan_week_tenant_household`); no new index added unless profiling demands it (out of scope).

## Build & test commands

- Backend: `cd services/recipe-service && go build ./... && go test ./internal/recipe/`
- Frontend: `cd frontend && npx tsc --noEmit && TZ=UTC npx vitest run && npm run build`
- Docker: build `recipe-service` (see `scripts/local-up.sh` for canonical context).
