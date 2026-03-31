# Plan Audit — task-016-ingredient-categories

**Plan Path:** docs/tasks/task-016-ingredient-categories/tasks.md
**Audit Date:** 2026-03-31
**Branch:** task-016
**Base Branch:** main

## Executive Summary

All 36 implementation tasks across Phases 1–4 are implemented with evidence in the git diff. The frontend Docker build fails due to a TypeScript strict-mode error in `category-manager.tsx`, blocking deployment. Backend builds, tests, and Docker build all pass cleanly. The new `ingredient/category` domain is fully compliant with backend developer guidelines. Two low-severity items remain in the category domain (aggregate provider functions use raw return types; `seedDefaults` constructs entities directly for transaction safety). Frontend code is fully compliant with frontend guidelines. Phase 5 E2E tasks cannot be verified (no automated infrastructure).

## Task Completion

| # | Task | Status | Evidence / Notes |
|---|------|--------|-----------------|
| 1.1 | Create `category/model.go` — immutable Model | DONE | `category/model.go:9-30` — 7 private fields, 7 accessors, `WithIngredientCount` returns copy |
| 1.2 | Create `category/builder.go` — fluent builder | DONE | `category/builder.go:26-55` — `NewBuilder()`, 7 fluent setters, `Build()` validates name required/length, sortOrder >= 0 |
| 1.3 | Create `category/entity.go` — GORM Entity, Migration, Make | DONE | `category/entity.go:10-45` — Entity with unique index `(tenant_id, name)`, `Migration()`, `Make()` with error, `ToEntity()` |
| 1.4 | Create `category/provider.go` — query functions | DONE | `category/provider.go:9-63` — `GetByID`, `GetAll`, `GetByName` use `database.Query[T]`/`database.SliceQuery[T]`; `CountIngredientsByCategory` for ingredient count |
| 1.5 | Create `category/processor.go` — List (auto-seed), Create, Update, Delete | DONE | `category/processor.go:45-174` — auto-seed with transaction re-check for race protection; delegates writes to `administrator.go` |
| 1.6 | Create `category/rest.go` — RestModel with JSON:API mapping | DONE | `category/rest.go:9-69` — RestModel with `GetName`/`GetID`/`SetID`, `Transform`/`TransformSlice` with error returns, flat request models |
| 1.7 | Create `category/resource.go` — GET/POST/PATCH/DELETE routes | DONE | `category/resource.go:16-151` — `RegisterHandler` for GET/DELETE, `RegisterInputHandler[T]` for POST/PATCH, `d.Logger()` used throughout, Transform errors checked |
| 1.8 | Modify `ingredient/entity.go` — add CategoryId FK | DONE | `ingredient/entity.go:16` — `CategoryId *uuid.UUID` with GORM index; lines 38-41 add FK `ON DELETE SET NULL` |
| 1.9 | Modify `ingredient/model.go` — add categoryID, categoryName | DONE | `ingredient/model.go:42-43` — private fields; accessors at lines 56-57; `WithCategoryName` at lines 69-72 |
| 1.10 | Modify `ingredient/builder.go` — add WithCategoryID | DONE | `ingredient/builder.go:37` — `SetCategoryID(*uuid.UUID)` method (uses `Set` prefix consistent with this domain's builder pattern) |
| 1.11 | Register category.Migration before ingredient.Migration | DONE | `cmd/main.go:32` — category migration registered before ingredient for FK dependency |
| 1.12 | Build recipe-service and verify | DONE | `go build ./...` passes cleanly |
| 2.1 | Modify `ingredient/provider.go` — join category for categoryName | DONE | `ingredient/provider.go:133-134` — LEFT JOIN `ingredient_categories`, COALESCE for null category names |
| 2.2 | Modify `ingredient/processor.go` — accept categoryID, validate | DONE | `ingredient/processor.go:36` — Create accepts `categoryID *uuid.UUID`; line 103 Update accepts `*UpdateCategoryOpt` with Set/Value pattern |
| 2.3 | Modify `ingredient/rest.go` — add category fields | DONE | `ingredient/rest.go:14-15,31-32` — `CategoryId`/`CategoryName` on RestModel and RestDetailModel; `52,67` on request models |
| 2.4 | Modify `ingredient/resource.go` — parse categoryId | DONE | `ingredient/resource.go:93-101` — UUID parsing with null/empty handling in create; lines 157-169 in update with clear-category support |
| 2.5 | Add BulkCategorize to `ingredient/processor.go` | DONE | `ingredient/processor.go:224-242` — validates category exists and belongs to tenant, single-transaction update |
| 2.6 | Add POST /ingredients/bulk-categorize route | DONE | `ingredient/resource.go:28` — route registered with `RegisterInputHandler[BulkCategorizeRequest]` |
| 2.7 | Modify `export/processor.go` — add category fields | DONE | `export/processor.go:48-49` — `CategoryName` and `CategorySortOrder` on `ConsolidatedIngredient`; populated via category lookup map at lines 93-106, 161-166 |
| 2.8 | Modify `export/resource.go` — add category fields to REST | DONE | `export/resource.go:21-22` — `CategoryName *string` and `CategorySortOrder *int` (nullable pointers); `TransformIngredient` at lines 42-46 handles null |
| 2.9 | Modify `export/markdown.go` — group by category headers | DONE | `export/markdown.go:96-157` — groups by category name, sorts by sort_order (line 128), uncategorized at end (line 143), empty categories omitted |
| 2.10 | Build and test recipe-service | DONE | `go build ./...` and `go test ./... -count=1` — all 9 test packages pass |
| 3.1 | Add IngredientCategory type to `types/models/ingredient.ts` | DONE | `ingredient.ts:60-81` — `IngredientCategory`, `IngredientCategoryAttributes`, Create/Update attribute types |
| 3.2 | Add categoryId, categoryName to canonical types | DONE | `ingredient.ts:10-11` (list), `22-23` (detail), `45` (create), `52` (update) |
| 3.3 | Add category_name, category_sort_order to PlanIngredientAttributes | DONE | `meal-plan.ts:83-84` |
| 3.4 | Add category CRUD + bulkCategorize to ingredient service | DONE | `services/api/ingredient.ts:93-121` — `listCategories`, `createCategory`, `updateCategory`, `deleteCategory`, `bulkCategorize` |
| 3.5 | Create `use-ingredient-categories.ts` — hooks | DONE | Full file — `useIngredientCategories` query + 4 mutation hooks; proper `enabled` guard, `onSettled` invalidation, cross-domain invalidation |
| 3.6 | Update `use-ingredients.ts` type references | DONE | `use-ingredients.ts:15-20` — `UseIngredientsParams` includes `categoryId`; mutations invalidate `categoryKeys.lists()` for cross-domain |
| 4.1 | Category management UI | DONE | `category-manager.tsx` — list, create, rename, delete with Zod validation (`categoryNameSchema`), confirmation on delete |
| 4.2 | Category selector on IngredientDetailPage | DONE | `IngredientDetailPage.tsx:183-202` — Select dropdown with "Uncategorized" option, persists on change |
| 4.3 | Category badge on IngredientsPage cards | DONE | `IngredientsPage.tsx:249-252` — Badge with category name or "uncategorized" styling |
| 4.4 | Category filter on IngredientsPage | DONE | `IngredientsPage.tsx:159-174` — Select with All/Uncategorized/per-category options; filter parameter sent to API |
| 4.5 | Uncategorized count on IngredientsPage | DONE | `IngredientsPage.tsx:64-71,105-109` — Badge showing uncategorized count with smart calculation |
| 4.6 | Bulk category assignment UI | DONE | `bulk-categorize.tsx:1-190` — multi-select with toggleAll, category/uncategorized/search filters, target category selector, 200-ingredient page size |
| 4.7 | Modify ingredient-preview — group by category | DONE | `ingredient-preview.tsx:22-70` — `useMemo` groups by `category_name`, sorts by `category_sort_order`, uncategorized last, alphabetical within groups |
| 4.8 | Frontend build verification | PARTIAL | `tsc --noEmit` passes locally; Docker `tsc -b` fails — see violation #1 |
| 5.1 | Docker build recipe-service | DONE | Docker build succeeds |
| 5.2 | Docker build frontend | FAIL | `tsc -b` fails: TS2532 at `category-manager.tsx:31,46` |
| 5.3 | E2E: create categories, assign, verify grouping | SKIPPED | No automated E2E infrastructure |
| 5.4 | E2E: markdown export with category headers | SKIPPED | No automated E2E infrastructure |
| 5.5 | E2E: delete category, verify uncategorized | SKIPPED | No automated E2E infrastructure |
| 5.6 | E2E: bulk categorize | SKIPPED | No automated E2E infrastructure |
| 5.7 | Verify backward compatibility | SKIPPED | No automated E2E infrastructure |

**Completion Rate:** 36/43 tasks DONE (84%), 1 PARTIAL (4.8), 1 FAIL (5.2), 5 SKIPPED (5.3–5.7)
**Skipped without approval:** 5 (Phase 5 E2E — no infrastructure)
**Partial implementations:** 1 (task 4.8 — Docker build fails)

## Skipped / Deferred Tasks

Tasks 5.3–5.7 are E2E verification tasks requiring a running Docker + database environment. No automated E2E infrastructure exists in the project. The underlying feature implementation is complete — these are verification-only tasks.

## Developer Guidelines Compliance

### Passes

**Backend — Category Domain (new code):**
- Immutable Model: all 7 fields private with public accessors (`category/model.go:10-25`)
- Entity with GORM tags, `Migration()`, `Make(Entity)(Model, error)`, `ToEntity()` on Model (`category/entity.go:10-45`)
- Builder with `NewBuilder()`, 7 fluent setters, `Build()` validates 3 invariants (`category/builder.go:26-55`)
- Processor uses `logrus.FieldLogger` interface (`category/processor.go:36`)
- Processor delegates all writes to `administrator.go` functions (`category/processor.go:123,156,173` → `category/administrator.go:10-25`)
- Administrator: `createCategory` does not take tenantID explicitly (entity field set by caller); `updateCategory`/`deleteCategory` do not take tenantID (`category/administrator.go:10-25`)
- Providers use `database.Query[T]` and `database.SliceQuery[T]` for lazy evaluation (`category/provider.go:9-29`)
- No provider function takes tenantID — filtering automatic via GORM callbacks (`category/provider.go`)
- Handlers use `d.Logger()` throughout, never `logrus.StandardLogger()` (`category/resource.go`)
- `RegisterHandler` for GET/DELETE, `RegisterInputHandler[T]` for POST/PATCH (`category/resource.go:18-25`)
- All Transform errors checked and logged (`category/resource.go:42-46,79-83,120-124`)
- Handlers delegate to Processor, never call providers directly (`category/resource.go`)
- Tenant scoping via `tenantctx.MustFromContext` (`category/resource.go:32,61`)
- REST models: flat structure, `json:"-"` on ID, `GetName`/`GetID`/`SetID` interface methods (`category/rest.go`)
- Both `Transform` and `TransformSlice` exist with error returns (`category/rest.go:48-69`)
- Unit tests: builder (187 lines), entity roundtrip (133 lines), REST transform (108 lines) (`category/*_test.go`)
- Error mapping: 404, 409, 422, 500 (`category/resource.go:66-76,103-114,139-143`)

**Backend — Ingredient Domain (modified code):**
- `categoryID`/`categoryName` are private with accessors (`ingredient/model.go:42-43,56-57`)
- FK with `ON DELETE SET NULL` and index (`ingredient/entity.go:16,38-41`)
- `Make()` handles CategoryId (`ingredient/entity.go:64`)
- Category migration registered before ingredient migration (`cmd/main.go:32`)
- BulkCategorize validates category existence and tenant ownership (`ingredient/processor.go:225-232`)
- Processor uses `logrus.FieldLogger` (`ingredient/processor.go:23`)

**Backend — Export Domain (modified code):**
- `ConsolidatedIngredient` has `CategoryName` and `CategorySortOrder` (`export/processor.go:48-49`)
- Category data populated via lookup map during consolidation (`export/processor.go:93-106,161-166`)
- REST model has nullable category fields (`export/resource.go:21-22`)
- Markdown groups by category, sorts by sort_order, uncategorized at end, empty categories omitted (`export/markdown.go:96-157`)

**Frontend (all new/modified code):**
- JSON:API model structure throughout (`ingredient.ts:29-39,68-72`)
- Service extends `BaseService`, tenant parameter on all methods (`services/api/ingredient.ts`)
- Query key factories with `as const`, hierarchical keys, tenant in keys (`query-keys.ts`)
- `enabled` guards with `!!tenant?.id && !!household?.id` on all query hooks (`use-ingredient-categories.ts:21`)
- Mutations invalidate correct keys with cross-domain awareness (ingredient delete → category lists, etc.)
- Zod schema in `lib/schemas/ingredient-category.schema.ts`, used by `category-manager.tsx`
- Named exports on all components (no default exports)
- `cn()` for conditional classes (`ingredient-preview.tsx:114`)
- Skeleton loading states throughout (`IngredientsPage.tsx:196-203`, `category-manager.tsx:71-78`, etc.)
- `createErrorFromUnknown()` for error handling in all async operations
- Toast feedback via sonner for all user actions
- No `any` types in new code

### Violations

#### 1. Frontend Docker build fails — TS2532 in `category-manager.tsx`
- **Rule:** Task 5.2 — Docker build frontend must succeed; TypeScript strict mode enforced
- **File:** `frontend/src/components/features/ingredients/category-manager.tsx:31,46`
- **Issue:** `result.error.issues[0].message` triggers TS2532 ("Object is possibly 'undefined'"). Zod's `safeParse` returns a discriminated union, but `tsc -b` with `strict: true` (in `tsconfig.app.json`) does not narrow `result.error` after the `!result.success` guard. Local `tsc --noEmit` uses the root `tsconfig.json` which has `"files": []` and no compiler options, so it type-checks nothing — it's effectively a no-op.
- **Severity:** high — blocks Docker deployment
- **Fix:** Use optional chaining: `result.error?.issues[0]?.message ?? "Validation failed"`, or assign the narrowed error before the guard: `const { error } = result; if (!result.success) { toast.error(error.issues[0].message); }`.

#### 2. Aggregate provider functions use raw return types (category domain)
- **Rule:** Provider pattern: "Return `model.Provider[T]` for lazy evaluation"
- **File:** `services/recipe-service/internal/ingredient/category/provider.go:33-63`
- **Issue:** `CountIngredientsByCategory` returns `func(db *gorm.DB) (int64, error)` instead of `database.EntityProvider[T]`. `getMaxSortOrder` and `countAll` are private helpers that accept `*gorm.DB` directly. These are aggregate queries (COUNT, MAX) that don't map to entities.
- **Severity:** low — acceptable pragmatic exception for aggregate queries
- **Fix:** Document as convention exception. The core entity providers (`GetByID`, `GetAll`, `GetByName`) correctly use `database.Query[T]`/`database.SliceQuery[T]`.

#### 3. `seedDefaults` constructs entities directly in processor (category domain)
- **Rule:** Anti-pattern: "processor.go → entity.go directly for database queries"
- **File:** `services/recipe-service/internal/ingredient/category/processor.go:77-99`
- **Issue:** Constructs `Entity{}` structs directly for static seed data within a transaction. This is documented with a comment (line 74-76) explaining the transaction-based race protection requires operating within a single tx. The actual write delegates to `createCategory()` from `administrator.go`.
- **Severity:** low — documented exception with clear rationale
- **Fix:** None needed. Comment explains the design trade-off.

#### 4. `BulkCategorize` performs direct DB update in processor (ingredient domain)
- **Rule:** "All writes must go through administrator functions"
- **File:** `services/recipe-service/internal/ingredient/processor.go:234-241`
- **Issue:** Uses `tx.Table("canonical_ingredients").Updates(...)` directly in the processor. The ingredient domain has no `administrator.go` file — this is a pre-existing pattern where all ingredient writes happen directly in the processor (e.g., `Create` at line 58 does `p.db.WithContext(p.ctx).Create(&e)`).
- **Severity:** low — consistent with pre-existing ingredient domain patterns
- **Fix:** If the ingredient domain is refactored to add `administrator.go` in a future task, this should be extracted there.

## Build & Test Results

| Service | Build | Tests | Notes |
|---------|-------|-------|-------|
| recipe-service (`go build`) | PASS | PASS | All 9 test packages pass |
| recipe-service (Docker) | PASS | — | Image built successfully |
| frontend (`tsc --noEmit`) | PASS | N/A | Note: root tsconfig has `"files": []` — this is effectively a no-op |
| frontend (Docker `tsc -b`) | FAIL | — | TS2532 at `category-manager.tsx:31,46` |

## Overall Assessment

- **Plan Adherence:** MOSTLY_COMPLETE — 36/43 tasks done, 1 partial, 1 fail, 5 skipped (E2E)
- **Guidelines Compliance:** MINOR_VIOLATIONS — 1 high (Docker build), 3 low (pragmatic exceptions / pre-existing patterns)
- **Recommendation:** NEEDS_FIXES

## Action Items

1. **Fix frontend Docker build** — Add null safety to `category-manager.tsx:31,46` where `result.error.issues[0].message` triggers TS2532. This blocks deployment. (High — required before merge)
2. **Update `tasks.md`** — All 43 task checkboxes remain unchecked despite implementation being complete. (Low — housekeeping)
