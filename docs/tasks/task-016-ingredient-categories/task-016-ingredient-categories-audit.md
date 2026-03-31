# Plan Audit — task-016-ingredient-categories

**Plan Path:** docs/tasks/task-016-ingredient-categories/tasks.md
**Audit Date:** 2026-03-31
**Branch:** task-016
**Base Branch:** main

## Executive Summary

All 40 tasks across Phases 1–5 are implemented and builds/tests pass for both backend and frontend. Docker builds succeed. However, the new `ingredient/category` domain has 3 guideline violations in new code: the `GetByTenantID` provider does not use `database.SliceQuery`, two provider functions (`CountIngredientsByCategory`, `GetMaxSortOrder`) bypass the curried provider pattern entirely, and `Transform`/`TransformSlice` in `category/rest.go` are missing the error return type that every other service uses. The `BulkCategorize` processor also lacks category existence validation before updating.

## Task Completion

| # | Task | Status | Evidence / Notes |
|---|------|--------|-----------------|
| 1.1 | Create `category/model.go` — immutable Model | DONE | `category/model.go`: private fields, public accessors |
| 1.2 | Create `category/builder.go` — fluent builder | DONE | `category/builder.go`: validates name required, max 100 chars, sortOrder >= 0 |
| 1.3 | Create `category/entity.go` — GORM Entity, Migration, Make | DONE | `category/entity.go`: unique index on (tenant_id, name), Migration(), Make() |
| 1.4 | Create `category/provider.go` — GetByID, GetByTenantID, GetByName, CountIngredientsByCategory | DONE | `category/provider.go`: all 4 + GetMaxSortOrder |
| 1.5 | Create `category/processor.go` — List (auto-seed), Create, Update, Delete | DONE | `category/processor.go`: auto-seed with race-condition protection, all CRUD ops |
| 1.6 | Create `category/rest.go` — RestModel with JSON:API mapping | DONE | `category/rest.go`: RestModel, CreateRequest, UpdateRequest, Transform, TransformSlice |
| 1.7 | Create `category/resource.go` — GET/POST/PATCH/DELETE routes | DONE | `category/resource.go`: all 4 routes under `/ingredient-categories` |
| 1.8 | Modify `ingredient/entity.go` — add CategoryId FK | DONE | `ingredient/entity.go:16`: `CategoryId *uuid.UUID` with index, FK with ON DELETE SET NULL |
| 1.9 | Modify `ingredient/model.go` — add categoryID, categoryName | DONE | `ingredient/model.go:42-43`: private fields with accessors |
| 1.10 | Modify `ingredient/builder.go` — add WithCategoryID | DONE | `ingredient/builder.go:37`: `SetCategoryID(*uuid.UUID)` method |
| 1.11 | Register category.Migration before ingredient.Migration | DONE | `cmd/main.go:32`: `category.Migration` registered before `ingredient.Migration` |
| 1.12 | Build recipe-service and verify | DONE | `go build ./...` passes cleanly |
| 2.1 | Modify `ingredient/provider.go` — join category for categoryName | DONE | `ingredient/provider.go:111-155`: `searchWithUsage` joins `ingredient_categories` |
| 2.2 | Modify `ingredient/processor.go` — accept categoryID, validate | DONE | `ingredient/processor.go:35,97-100`: Create and Update accept categoryID/categoryOpt |
| 2.3 | Modify `ingredient/rest.go` — add category fields to REST models | DONE | `ingredient/rest.go:14-15,31-32,52,67`: CategoryId and CategoryName on both models |
| 2.4 | Modify `ingredient/resource.go` — parse categoryId | DONE | `ingredient/resource.go`: parse in create and update handlers |
| 2.5 | Add BulkCategorize to `ingredient/processor.go` | DONE | `ingredient/processor.go:223-232`: single transaction update |
| 2.6 | Add POST /ingredients/bulk-categorize route | DONE | `ingredient/resource.go:28`: route registered, handler at line 333 |
| 2.7 | Modify `export/processor.go` — add category fields to ConsolidatedIngredient | DONE | `export/processor.go:48-49`: CategoryName and CategorySortOrder fields |
| 2.8 | Modify `export/resource.go` + `rest.go` — add category fields | DONE | `export/resource.go:21-22`: CategoryName and CategorySortOrder on RestIngredientModel |
| 2.9 | Modify `export/markdown.go` — group by category headers | DONE | `export/markdown.go:96-157`: groups by category with ### headers, uncategorized at end |
| 2.10 | Build and test recipe-service | DONE | `go build ./...` and `go test ./... -count=1` both pass |
| 3.1 | Add IngredientCategory type to `types/models/ingredient.ts` | DONE | `ingredient.ts:60-81`: IngredientCategory type with create/update variants |
| 3.2 | Add categoryId, categoryName to canonical ingredient types | DONE | `ingredient.ts:10-11,23-24`: on both list and detail attributes |
| 3.3 | Add category_name, category_sort_order to PlanIngredientAttributes | DONE | `meal-plan.ts:83-84`: both fields added |
| 3.4 | Add category CRUD + bulkCategorize to ingredient service | DONE | `services/api/ingredient.ts:93-121`: all 5 methods |
| 3.5 | Create `use-ingredient-categories.ts` — query and mutation hooks | DONE | Full file with query + 4 mutation hooks |
| 3.6 | Update `use-ingredients.ts` type references | DONE | Updated with categoryId param, categoryKeys invalidation |
| 4.1 | Category management UI | DONE | `category-manager.tsx`: list, create, rename, delete |
| 4.2 | Category selector on IngredientDetailPage | DONE | `IngredientDetailPage.tsx:183-202`: Select dropdown with "Uncategorized" option |
| 4.3 | Category badge on IngredientsPage cards | DONE | `IngredientsPage.tsx:249-253`: Badge with category name |
| 4.4 | Category filter on IngredientsPage | DONE | `IngredientsPage.tsx:159-174`: Select with all/uncategorized/per-category |
| 4.5 | Uncategorized count on IngredientsPage | DONE | `IngredientsPage.tsx:105-109`: Badge showing uncategorized count |
| 4.6 | Bulk category assignment UI | DONE | `bulk-categorize.tsx`: multi-select, filter, apply |
| 4.7 | Modify ingredient-preview — group by category | DONE | `ingredient-preview.tsx:22-70`: groups by category, sorts by sort_order |
| 4.8 | Frontend build verification | DONE | `npm run build` passes (`tsc -b && vite build`) |
| 5.1 | Docker build recipe-service | DONE | Docker multi-stage build passes |
| 5.2 | Docker build frontend | DONE | Docker build passes |
| 5.3 | E2E: create categories, assign, verify grouping | DONE | Manual E2E verification completed |
| 5.4 | E2E: markdown export with category headers | DONE | Manual E2E verification completed |
| 5.5 | E2E: delete category, verify uncategorized | DONE | Manual E2E verification completed |
| 5.6 | E2E: bulk categorize | DONE | Manual E2E verification completed |
| 5.7 | Verify backward compatibility | DONE | Manual E2E verification completed |

**Completion Rate:** 40/40 tasks (100%)
**Skipped without approval:** 0
**Partial implementations:** 0

## Skipped / Deferred Tasks

None — all tasks completed.

## Developer Guidelines Compliance

### Passes

**Backend:**
- Immutable Model with private fields and public accessors (`category/model.go`)
- Builder pattern with validation in `Build()` (`category/builder.go`)
- Entity with GORM tags separated from Model (`category/entity.go`)
- `Make(Entity) (Model, error)` converter on entity (`category/entity.go`)
- Write operations extracted to `administrator.go` (`category/administrator.go`)
- `logrus.FieldLogger` interface used in processor constructor (`category/processor.go:36`)
- `server.RegisterHandler` / `server.RegisterInputHandler[T]` used correctly (`category/resource.go:18-20`)
- Tenant scoping via `tenantctx.MustFromContext` in every handler (`category/resource.go:32,56,84,122`)
- JSON:API interface methods on RestModel (`GetName`, `GetID`, `SetID`) (`category/rest.go:18-20`)
- Migration registration order correct — category before ingredient for FK dependency (`cmd/main.go:32-33`)
- Proper error mapping to HTTP status codes (404, 409, 422, 500) (`category/resource.go`)
- Unit tests for builder, entity, and rest layers (`category/*_test.go`)
- Ingredient entity FK with `ON DELETE SET NULL` and index (`ingredient/entity.go:16,39-41`)
- Export markdown groups by category with sort order, uncategorized at end (`export/markdown.go:96-157`)

**Frontend:**
- JSON:API model structure (`{ id, type, attributes }`) for all types
- Service layer extends `BaseService` with tenant as first parameter
- Query key factory with hierarchical keys and `as const` (`query-keys.ts`)
- Tenant context `enabled` guards on all queries
- Mutations invalidate correct query keys including cross-domain (ingredient → category)
- Skeleton loading states used throughout
- Toast notifications for success/error feedback
- Named exports used (no default exports)
- Error handling via `createErrorFromUnknown()` pattern
- No `any` types found
- `useMemo` optimization for ingredient preview grouping logic

### Violations

#### 1. `GetByTenantID` does not use `database.SliceQuery`
- **Rule:** Provider pattern — list queries must use `database.SliceQuery[Entity]` (see every other service: `forecast/provider.go`, `planitem/provider.go`, `task/provider.go`, etc.)
- **File:** `services/recipe-service/internal/ingredient/category/provider.go:15-21`
- **Issue:** Returns `func(db *gorm.DB) ([]Entity, error)` instead of using `database.SliceQuery[Entity]`. Every other list provider in the codebase uses the curried `SliceQuery` pattern.
- **Severity:** medium
- **Fix:** Refactor to: `func GetByTenantID(tenantID uuid.UUID) database.EntityListProvider[Entity] { return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB { return db.Where("tenant_id = ?", tenantID).Order("sort_order ASC") }) }`

#### 2. `CountIngredientsByCategory` and `GetMaxSortOrder` bypass curried provider pattern
- **Rule:** Provider pattern — all database access must use curried functions for lazy evaluation and composition
- **File:** `services/recipe-service/internal/ingredient/category/provider.go:29-35` and `provider.go:37-50`
- **Issue:** Both functions take `*gorm.DB` as first parameter and execute eagerly, bypassing the functional provider pattern. Every other provider function in the codebase returns a closure over `*gorm.DB`.
- **Severity:** medium
- **Fix:** Refactor both to curried signatures returning provider functions (e.g., `func CountIngredientsByCategory(categoryID uuid.UUID) func(db *gorm.DB) (int64, error)`)

#### 3. `Transform` / `TransformSlice` missing error return type
- **Rule:** Transform functions must return `(RestModel, error)` — every other service in the codebase follows this pattern
- **File:** `services/recipe-service/internal/ingredient/category/rest.go:48-65`
- **Issue:** `Transform(m Model) RestModel` and `TransformSlice(models []Model) []RestModel` omit the error return. All 13 other Transform implementations across the codebase return `(RestModel, error)`.
- **Severity:** medium
- **Fix:** Change signatures to `Transform(m Model) (RestModel, error)` and `TransformSlice(models []Model) ([]RestModel, error)`, update callers in `resource.go` to check errors.

#### 4. `BulkCategorize` lacks category existence validation
- **Rule:** Processor validation — validate external references belong to tenant before persisting
- **File:** `services/recipe-service/internal/ingredient/processor.go:223-232`
- **Issue:** `BulkCategorize` updates ingredient `category_id` without verifying the `categoryID` exists in `ingredient_categories` and belongs to the same tenant. If an invalid or wrong-tenant category ID is provided, the FK constraint would catch it at the DB level, but the plan specifies "Validates category belongs to tenant" as an acceptance criterion.
- **Severity:** low (FK constraint provides safety net)
- **Fix:** Add category lookup via `category.GetByID` and tenant check before the update transaction.

#### 5. `seedDefaults` has direct database access in processor
- **Rule:** Processor purity — processors should not contain direct DB calls; delegate to providers/administrators
- **File:** `services/recipe-service/internal/ingredient/category/processor.go:74-97`
- **Issue:** `seedDefaults` directly queries `tx.Model(&Entity{}).Where(...).Count(...)` inside a transaction. Processors should delegate all DB access to providers. The race-condition check inside the transaction makes this harder to refactor, but the direct Entity reference in processor code violates layer separation.
- **Severity:** low (contained within a single transaction, correct behavior)
- **Fix:** Extract the count check to a provider function that accepts a `*gorm.DB` (transaction-aware), or accept this as a pragmatic exception documented with a comment.

## Build & Test Results

| Service | Build | Tests | Notes |
|---------|-------|-------|-------|
| recipe-service | PASS | PASS | `go build ./...` clean, `go test ./... -count=1` all pass including new category tests |
| recipe-service (Docker) | PASS | — | Multi-stage Docker build succeeds |
| frontend | PASS | N/A | `tsc -b && vite build` clean (chunk size warning, non-blocking) |
| frontend (Docker) | PASS | — | Docker build succeeds |

## Overall Assessment

- **Plan Adherence:** FULL — 40/40 tasks completed
- **Guidelines Compliance:** MINOR_VIOLATIONS — 5 violations found in new code (3 medium, 2 low severity)
- **Recommendation:** NEEDS_FIXES

## Action Items

1. **Refactor `GetByTenantID`** in `category/provider.go:15-21` to use `database.SliceQuery[Entity]` pattern consistent with every other list provider in the codebase.
2. **Refactor `CountIngredientsByCategory` and `GetMaxSortOrder`** in `category/provider.go:29-50` to use curried provider signatures.
3. **Add error return** to `Transform` and `TransformSlice` in `category/rest.go:48-65` and update callers in `resource.go` to check the returned error.
4. **Add category existence validation** to `BulkCategorize` in `ingredient/processor.go:223-232` — verify category exists and belongs to tenant before updating.
5. **Consider extracting** the direct DB count query in `seedDefaults` (`category/processor.go:77-78`) to a provider function, or add a comment documenting why the exception is acceptable.
