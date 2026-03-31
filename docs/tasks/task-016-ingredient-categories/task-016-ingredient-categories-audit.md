# Plan Audit — task-016-ingredient-categories

**Plan Path:** docs/tasks/task-016-ingredient-categories/tasks.md
**Audit Date:** 2026-03-31
**Branch:** task-016
**Base Branch:** main

## Executive Summary

All 36 implementation tasks across Phases 1–4 are implemented, builds pass, and all backend tests pass. The 7 Phase 5 integration/Docker tasks cannot be verified in this audit environment and are marked SKIPPED. The tasks.md file was never updated to mark tasks as checked. The new `ingredient/category` domain has 4 guideline violations in new code: provider functions pass explicit `tenantID` instead of relying on GORM callbacks, the processor's `Update`/`Delete` methods accept `tenantID` (anti-pattern), helper provider functions return raw function types instead of `database.EntityProvider[T]`, and `seedDefaults` performs direct DB access in the processor. Frontend has 1 minor gap: category forms lack Zod validation schemas.

## Task Completion

| # | Task | Status | Evidence / Notes |
|---|------|--------|-----------------|
| 1.1 | Create `category/model.go` — immutable Model | DONE | `category/model.go:9-30`: private fields, public accessors, `WithIngredientCount` returns copy |
| 1.2 | Create `category/builder.go` — fluent builder | DONE | `category/builder.go:16-55`: validates name required, max 100, sortOrder >= 0 |
| 1.3 | Create `category/entity.go` — GORM Entity, Migration, Make | DONE | `category/entity.go:10-45`: unique index on (tenant_id, name), `Migration()`, `Make()` with error |
| 1.4 | Create `category/provider.go` — query functions | DONE | `category/provider.go:9-60`: GetByID, GetByTenantID, GetByName, CountIngredientsByCategory, GetMaxSortOrder, CountByTenantID |
| 1.5 | Create `category/processor.go` — List (auto-seed), Create, Update, Delete | DONE | `category/processor.go:45-172`: auto-seed with race-condition protection via transaction re-check |
| 1.6 | Create `category/rest.go` — RestModel with JSON:API mapping | DONE | `category/rest.go:9-69`: RestModel, CreateRequest, UpdateRequest, Transform/TransformSlice with error returns |
| 1.7 | Create `category/resource.go` — GET/POST/PATCH/DELETE routes | DONE | `category/resource.go:16-153`: all 4 routes under `/ingredient-categories` |
| 1.8 | Modify `ingredient/entity.go` — add CategoryId FK | DONE | `ingredient/entity.go:16`: `CategoryId *uuid.UUID` with index, FK ON DELETE SET NULL |
| 1.9 | Modify `ingredient/model.go` — add categoryID, categoryName | DONE | `ingredient/model.go:42-43`: private fields with accessors + `WithCategoryName` |
| 1.10 | Modify `ingredient/builder.go` — add WithCategoryID | DONE | `ingredient/builder.go:37`: `SetCategoryID(*uuid.UUID)` method |
| 1.11 | Register category.Migration before ingredient.Migration | DONE | `cmd/main.go:32`: category migration registered before ingredient for FK dependency |
| 1.12 | Build recipe-service and verify | DONE | `go build ./...` passes cleanly |
| 2.1 | Modify `ingredient/provider.go` — join category for categoryName | DONE | `ingredient/provider.go:111-155`: `searchWithUsage` LEFT JOINs `ingredient_categories` |
| 2.2 | Modify `ingredient/processor.go` — accept categoryID, validate | DONE | `ingredient/processor.go:35,97-134`: Create and Update accept categoryID with validation |
| 2.3 | Modify `ingredient/rest.go` — add category fields | DONE | `ingredient/rest.go:14-15,31-32,52,67`: CategoryId and CategoryName on list and detail models |
| 2.4 | Modify `ingredient/resource.go` — parse categoryId | DONE | `ingredient/resource.go`: parse in create and update handlers |
| 2.5 | Add BulkCategorize to `ingredient/processor.go` | DONE | `ingredient/processor.go:224-242`: validates category + tenant, single transaction update |
| 2.6 | Add POST /ingredients/bulk-categorize route | DONE | `ingredient/resource.go:28`: route registered |
| 2.7 | Modify `export/processor.go` — add category fields | DONE | `export/processor.go:48-49`: CategoryName and CategorySortOrder on ConsolidatedIngredient |
| 2.8 | Modify `export/resource.go` + `rest.go` — add category fields | DONE | `export/resource.go:21-22`: CategoryName and CategorySortOrder as nullable pointers |
| 2.9 | Modify `export/markdown.go` — group by category headers | DONE | `export/markdown.go:96-157`: groups by category, sort by sort_order, uncategorized at end |
| 2.10 | Build and test recipe-service | DONE | `go build ./...` and `go test ./... -count=1` all pass |
| 3.1 | Add IngredientCategory type to `types/models/ingredient.ts` | DONE | `ingredient.ts:60-81`: IngredientCategory, CreateAttributes, UpdateAttributes |
| 3.2 | Add categoryId, categoryName to canonical types | DONE | `ingredient.ts:10-11,23-24`: on both list and detail attributes |
| 3.3 | Add category_name, category_sort_order to PlanIngredientAttributes | DONE | `meal-plan.ts:83-84`: both fields added |
| 3.4 | Add category CRUD + bulkCategorize to ingredient service | DONE | `services/api/ingredient.ts:93-121`: all 5 methods |
| 3.5 | Create `use-ingredient-categories.ts` — query and mutation hooks | DONE | Full file: useIngredientCategories + 4 mutation hooks |
| 3.6 | Update `use-ingredients.ts` type references | DONE | Updated with categoryId param, categoryKeys invalidation |
| 4.1 | Category management UI | DONE | `category-manager.tsx`: list, create, rename, delete with confirmation |
| 4.2 | Category selector on IngredientDetailPage | DONE | `IngredientDetailPage.tsx:183-202`: Select dropdown with "Uncategorized" option |
| 4.3 | Category badge on IngredientsPage cards | DONE | `IngredientsPage.tsx:249-253`: Badge with category name |
| 4.4 | Category filter on IngredientsPage | DONE | `IngredientsPage.tsx:159-174`: Select with all/uncategorized/per-category |
| 4.5 | Uncategorized count on IngredientsPage | DONE | `IngredientsPage.tsx:105-109`: Badge showing uncategorized count |
| 4.6 | Bulk category assignment UI | DONE | `bulk-categorize.tsx`: multi-select, filter, search, apply |
| 4.7 | Modify ingredient-preview — group by category | DONE | `ingredient-preview.tsx:22-70`: groups by category, sorts by sort_order, uncategorized at end |
| 4.8 | Frontend build verification | DONE | `tsc --noEmit` and `vite build` both pass |
| 5.1 | Docker build recipe-service | SKIPPED | Cannot verify Docker builds in this audit environment |
| 5.2 | Docker build frontend | SKIPPED | Cannot verify Docker builds in this audit environment |
| 5.3 | E2E: create categories, assign, verify grouping | SKIPPED | No automated E2E test infrastructure |
| 5.4 | E2E: markdown export with category headers | SKIPPED | No automated E2E test infrastructure |
| 5.5 | E2E: delete category, verify uncategorized | SKIPPED | No automated E2E test infrastructure |
| 5.6 | E2E: bulk categorize | SKIPPED | No automated E2E test infrastructure |
| 5.7 | Verify backward compatibility | SKIPPED | No automated E2E test infrastructure |

**Completion Rate:** 36/43 tasks DONE (84%), 7 SKIPPED (Phase 5 integration tasks)
**Skipped without approval:** 7 (all Phase 5 — cannot be verified in audit environment)
**Partial implementations:** 0

## Skipped / Deferred Tasks

All 7 skipped tasks are in Phase 5 (Integration & Verification). These require a running Docker environment and database to verify. The previous audit claimed these were "manually verified" but no evidence artifacts exist. The implementation itself is complete — these are verification tasks, not implementation tasks.

## Developer Guidelines Compliance

### Passes

**Backend:**
- Immutable Model with private fields and public accessors (`category/model.go:9-30`)
- Entity separated from Model with GORM tags only on Entity (`category/entity.go:10-17`)
- `Make(Entity) (Model, error)` converter with proper error return (`category/entity.go:36-45`)
- `ToEntity()` method on Entity (`category/entity.go:25-34`)
- Builder pattern with fluent setters and `Build()` validation (`category/builder.go:16-55`)
- Write operations extracted to `administrator.go` (`category/administrator.go:10-25`)
- `logrus.FieldLogger` interface used in processor constructor (`category/processor.go:36`)
- `d.Logger()` used in all handlers, not `logrus.StandardLogger()` (`category/resource.go`)
- `server.RegisterHandler` for GET/DELETE, `server.RegisterInputHandler[T]` for POST/PATCH (`category/resource.go:18-25`)
- Tenant scoping via `tenantctx.MustFromContext` in every handler (`category/resource.go:32,61,94,137`)
- JSON:API interface methods on RestModel (`GetName`, `GetID`, `SetID`) (`category/rest.go:18-20`)
- Transform and TransformSlice return `(T, error)` (`category/rest.go:48-69`)
- Migration registration order correct for FK dependency (`cmd/main.go:32-33`)
- Proper error mapping: 404, 409, 422, 500 (`category/resource.go`)
- Unit tests for builder, entity, and rest layers (`category/*_test.go`)
- Ingredient entity FK with `ON DELETE SET NULL` and index (`ingredient/entity.go:16,39-41`)
- BulkCategorize validates category existence and tenant ownership before update (`ingredient/processor.go:225-232`)
- Export markdown groups by category with sort order, uncategorized at end (`export/markdown.go:96-157`)

**Frontend:**
- JSON:API model structure for all types (`ingredient.ts:60-81`)
- Service layer extends `BaseService` with tenant as first parameter (`ingredient.ts:93-111`)
- Query key factory with hierarchical keys and `as const` (`query-keys.ts:17-22`)
- Tenant context `enabled` guards on all queries (`use-ingredient-categories.ts:20`)
- Mutations invalidate correct query keys including cross-domain (`use-ingredient-categories.ts`)
- Skeleton loading states used throughout (`IngredientsPage.tsx`, `category-manager.tsx`)
- `useMemo` optimization for ingredient preview grouping logic (`ingredient-preview.tsx`)
- Named exports used (no default exports)
- No `any` types found in new code

### Violations

#### 1. Provider functions pass explicit `tenantID` instead of relying on GORM callbacks
- **Rule:** Anti-pattern: "Manual `Where("tenant_id = ?", ...)` in queries — Use `db.WithContext(ctx)` — GORM callback injects tenant filter"
- **File:** `services/recipe-service/internal/ingredient/category/provider.go:15-18,21-24,37-51,54-59`
- **Issue:** `GetByTenantID`, `GetByName`, `GetMaxSortOrder`, and `CountByTenantID` all pass `tenantID` as a parameter and manually filter with `Where("tenant_id = ?", tenantID)`. Per guidelines, the GORM callback should automatically inject tenant filtering when `db.WithContext(ctx)` is used.
- **Severity:** medium
- **Fix:** Remove `tenantID` parameters from provider functions. Rely on GORM callback for tenant filtering via context. Only `CountIngredientsByCategory` (which queries a different table) may need explicit filtering.

#### 2. Processor `Update` and `Delete` accept `tenantID` parameter
- **Rule:** Anti-pattern: "Passing TenantId to providers/update/delete — Automatic via GORM callbacks — only pass to create functions"
- **File:** `services/recipe-service/internal/ingredient/category/processor.go:126,165`
- **Issue:** `Update(id, tenantID, ...)` and `Delete(id, tenantID)` take tenantID and manually check `e.TenantId != tenantID`. Per guidelines, tenant scoping on reads/updates/deletes should be automatic via GORM callbacks. Only `Create` should receive tenantID.
- **Severity:** medium
- **Fix:** Remove `tenantID` from `Update` and `Delete` signatures. The `GetByID` provider call already uses `p.db.WithContext(p.ctx)` which applies tenant filtering automatically. Remove the manual `e.TenantId != tenantID` check.

#### 3. Helper provider functions return raw function types instead of `database.EntityProvider[T]`
- **Rule:** Provider pattern: "Return `model.Provider[T]` for lazy evaluation. Compose via `model.Map`, `model.SliceMap`"
- **File:** `services/recipe-service/internal/ingredient/category/provider.go:27-59`
- **Issue:** `CountIngredientsByCategory`, `GetMaxSortOrder`, and `CountByTenantID` return raw `func(db *gorm.DB) (T, error)` instead of using `database.EntityProvider[T]` or `database.Query[T]`. This prevents composition with `model.Map` and `model.ParallelMap`.
- **Severity:** low
- **Fix:** Refactor to use `database.Query[T]` where possible. For aggregate queries (COUNT, MAX) that don't map to an entity, the raw function pattern may be acceptable — document as a pragmatic exception.

#### 4. `seedDefaults` performs direct DB access in processor
- **Rule:** Anti-pattern: "`processor.go` → `entity.go` directly for database queries (should use provider)"
- **File:** `services/recipe-service/internal/ingredient/category/processor.go:74-97`
- **Issue:** `seedDefaults` directly constructs `Entity` structs and calls `CountByTenantID` (a raw function, not a provider). The transaction-based race-condition protection makes this harder to refactor cleanly.
- **Severity:** low
- **Fix:** Extract the count check to a proper provider function. The entity construction within the transaction is acceptable since it uses `createCategory` from `administrator.go`.

#### 5. Frontend category forms lack Zod validation schemas
- **Rule:** Frontend guidelines: forms should use react-hook-form with Zod validation schemas in `lib/schemas/`
- **File:** `frontend/src/components/features/ingredients/category-manager.tsx`, `bulk-categorize.tsx`
- **Issue:** Category create/rename uses simple string trimming (`newCategoryName.trim()`) instead of Zod schema validation. No `ingredientCategory.schema.ts` file exists.
- **Severity:** low
- **Fix:** Create `lib/schemas/ingredient-category.schema.ts` with Zod validation and integrate with category form inputs.

## Build & Test Results

| Service | Build | Tests | Notes |
|---------|-------|-------|-------|
| recipe-service | PASS | PASS | `go build ./...` clean, `go test ./... -count=1` all 11 packages pass including category tests |
| frontend (TypeScript) | PASS | N/A | `tsc --noEmit` passes cleanly |
| frontend (Vite) | PASS | N/A | `vite build` succeeds (chunk size warning, non-blocking) |
| Docker (both) | NOT_VERIFIED | — | Docker not available in audit environment |

## Overall Assessment

- **Plan Adherence:** MOSTLY_COMPLETE — 36/43 tasks implemented, 7 Phase 5 verification tasks skipped (not implementable without Docker/running environment)
- **Guidelines Compliance:** MINOR_VIOLATIONS — 5 violations found in new code (2 medium, 3 low severity)
- **Recommendation:** NEEDS_FIXES

## Action Items

1. **Remove explicit `tenantID` from provider functions** in `category/provider.go:15-59` — rely on GORM callback filtering via `db.WithContext(ctx)` instead of manual `Where("tenant_id = ?", tenantID)`. (Medium)
2. **Remove `tenantID` from `Update` and `Delete` signatures** in `category/processor.go:126,165` — tenant scoping should be automatic. Update callers in `category/resource.go:102,140`. (Medium)
3. **Refactor helper provider functions** in `category/provider.go:27-59` to use `database.EntityProvider[T]` where feasible, or document the aggregate-query exception. (Low)
4. **Extract `seedDefaults` count query** to a proper provider function, or add a comment documenting the transaction-based exception. (Low)
5. **Add Zod validation schemas** for category forms — create `lib/schemas/ingredient-category.schema.ts`. (Low)
6. **Verify Docker builds** and perform E2E testing for Phase 5 tasks. (Deferred)
7. **Update `tasks.md`** — all task checkboxes are still unchecked despite implementation being complete.
