# Plan Audit — task-019-shopping-list

**Plan Path:** docs/tasks/task-019-shopping-list/tasks.md
**Audit Date:** 2026-03-31
**Branch:** task-019
**Base Branch:** main

## Executive Summary

Implementation is substantially complete with 22 of 26 tasks done, covering both new backend services (category-service, shopping-service), recipe-service migration, frontend shopping features, and CI/CD integration. Three tasks are partial (category domain not deleted from recipe-service, limited unit tests, no service documentation), and one task is skipped (service documentation). Both backend services and the frontend build and pass all tests. Developer guidelines compliance is strong with minor violations around missing Zod schemas and orphaned files.

## Task Completion

| # | Task | Status | Evidence / Notes |
|---|------|--------|-----------------|
| 1.1 | Scaffold category-service (cmd, Dockerfile, go.mod, config, go.work) | DONE | `services/category-service/cmd/main.go`, `Dockerfile`, `go.mod`, `go.work:7` |
| 1.2 | Implement category domain (entity, model, builder, processor, provider, resource, rest) | DONE | All 8 domain files in `services/category-service/internal/category/` |
| 1.3 | Add category-service to docker-compose and nginx | DONE | `docker-compose.yml:163-178`, `nginx.conf:160-168` |
| 1.4 | Verify data migration — auto-seed 15 defaults, tenant scoping works | DONE | `processor.go:76-99` seeds 15 default categories; unique index on (tenant_id, name) in `entity.go:12` |
| 1.5 | Unit tests for category domain | PARTIAL | `builder_test.go` has 6 tests. No tests for processor, provider, or REST handlers |
| 2.1 | Add HTTP client for category-service in recipe-service | DONE | `services/recipe-service/internal/categoryclient/client.go` |
| 2.2 | Remove category domain from recipe-service | PARTIAL | Routes and migration disconnected in `cmd/main.go`. Category import removed from `ingredient/processor.go`. But all 11 files still tracked in `services/recipe-service/internal/ingredient/category/` |
| 2.3 | Update recipe-service tests | DONE | All recipe-service tests pass (`go test ./... -count=1`). Export processor updated to use catClient |
| 3.1 | Scaffold shopping-service (cmd, Dockerfile, go.mod, config, go.work) | DONE | `services/shopping-service/cmd/main.go`, `Dockerfile`, `go.mod`, `go.work:11` |
| 3.2 | Implement shopping list domain (entity, model, builder, processor, provider, resource, rest) | DONE | All domain files in `services/shopping-service/internal/list/` including `rest.go` with 13 handlers |
| 3.3 | Implement shopping item domain (entity, model, builder, processor, provider, resource, rest) | DONE | All domain files in `services/shopping-service/internal/item/` |
| 3.4 | Implement meal plan import (HTTP clients, transform, append) | DONE | `recipeclient/client.go`, `categoryclient/client.go`, import handler in `list/rest.go:439-531` |
| 3.5 | Add shopping-service to docker-compose and nginx | DONE | `docker-compose.yml:179-196`, `nginx.conf:170-178` |
| 3.6 | Unit tests for shopping domains | PARTIAL | Builder tests for both item and list. No processor, provider, or REST handler tests |
| 4.1 | Update category API calls to `/categories`, update types and query keys | DONE | `ingredient.ts` updated, `query-keys.ts` updated with categoryKeys, `category-manager.tsx` and `bulk-categorize.tsx` updated |
| 5.1 | Types and API service layer (ShoppingService, models) | DONE | `types/models/shopping.ts`, `services/api/shopping.ts` with 13 methods |
| 5.2 | React Query hooks (lists, items, import) | DONE | `lib/hooks/api/use-shopping.ts` with query and mutation hooks |
| 5.3 | Shopping list management page (list, create, rename, delete) | DONE | `pages/ShoppingListsPage.tsx` with tab interface, create dialog, rename, delete |
| 5.4 | Shopping list detail / edit mode (items, grouped by category) | DONE | `pages/ShoppingListDetailPage.tsx` with category grouping via `groupItemsByCategory()` |
| 5.5 | Shopping mode (check/uncheck, in-cart section, uncheck all) | DONE | Toggle between edit/shopping modes, checkbox toggle, uncheck all, checked items styled with opacity/line-through |
| 5.6 | Archive and history (finish shopping, history page, unarchive, delete) | DONE | "Finish Shopping" archives list, "History" tab shows archived lists, unarchive action available |
| 5.7 | Meal plan import integration (button on plan detail, list picker dialog) | DONE | Import dialog in `MealsPage.tsx:309-459`, import button in `ShoppingListDetailPage.tsx:190-192` |
| 5.8 | Navigation and routing | DONE | Routes in `App.tsx:62-63`, nav entry in `nav-config.ts:41` with ShoppingCart icon |
| 6.1 | Add category-service and shopping-service to CI pipeline | DONE | `main.yml:38-43`, `pr.yml:156-186, 276-283` |
| 6.2 | Service documentation (domain.md, rest.md, storage.md for both services) | SKIPPED | No documentation files found in either service directory |

**Completion Rate:** 22/26 tasks (85%)
**Skipped without approval:** 1 (task 6.2)
**Partial implementations:** 3 (tasks 1.5, 2.2, 3.6)

## Skipped / Deferred Tasks

### Task 2.2 — Remove category domain from recipe-service (PARTIAL)
**Missing:** The 11 files in `services/recipe-service/internal/ingredient/category/` are still tracked by git. Routes and migration were disconnected in `cmd/main.go`, and the import was removed from `ingredient/processor.go`, so the category domain is effectively orphaned — not wired up but not deleted.
**Impact:** Low — the orphaned code is dead and doesn't affect functionality, but adds maintenance noise and could cause confusion. The old category tests still run and pass, wasting CI time.

### Tasks 1.5, 3.6 — Unit tests (PARTIAL)
**Missing:** Both new services have builder validation tests only. No tests exist for processors, providers, REST handlers, or administrator functions.
**Impact:** Medium — builder tests verify construction invariants, but the business logic in processors and the HTTP layer in REST handlers are untested. Import logic in shopping-service is particularly complex and would benefit from testing.

### Task 6.2 — Service documentation (SKIPPED)
**Missing:** No `domain.md`, `rest.md`, or `storage.md` files in either `category-service/` or `shopping-service/`.
**Impact:** Low — the code is well-structured and follows established patterns, making it self-documenting. However, API contracts for external consumers would benefit from documentation.

## Developer Guidelines Compliance

### Passes

| Guideline | Evidence |
|-----------|----------|
| Immutable models with unexported fields and accessors | category `model.go`, item `model.go`, list `model.go` — all fields unexported, getter-only methods |
| Entity separation from model (GORM tags only on entities) | All three domains have separate `entity.go` with GORM tags and `Make()` conversion functions |
| Builder pattern with invariant enforcement | All three domains have `builder.go` with `Build() (Model, error)` and validation logic |
| Processor pattern for business logic | Struct-based processors matching project convention across all services |
| Provider pattern for data access | `provider.go` in all domains using `database.Query`/`database.SliceQuery` |
| REST resource/handler separation | `resource.go` for transport models, `rest.go` for handlers in all domains |
| Multi-tenancy context propagation | `tenantctx.MustFromContext()` in handlers, tenant passed through processor layer, DB uses `WithContext()` |
| JSON:API type structure (frontend) | `shopping.ts` uses `{ id, type, attributes }` structure |
| React Query with query key factory | `query-keys.ts` has hierarchical `shoppingKeys` with `as const` pattern |
| Multi-tenancy in frontend hooks | All hooks use `useTenant()` and include tenant/household in query keys |
| No `any` types in new frontend code | Verified — no `any` usage in new shopping files |
| Component patterns (pages as composition) | Shopping pages follow existing patterns with data fetching and layout |

### Violations

1. **Rule:** Zod validation schemas should be defined in `lib/schemas/`
   - **File:** `pages/ShoppingListsPage.tsx`, `pages/ShoppingListDetailPage.tsx`
   - **Issue:** Shopping list name creation and item quick-add use inline validation (empty string checks) rather than Zod schemas. The category manager correctly uses `categoryNameSchema` from `lib/schemas/`.
   - **Severity:** low
   - **Fix:** Create `lib/schemas/shopping.schema.ts` with name validation schema. This is a minor violation since the shopping forms are simple inline inputs rather than full form dialogs.

2. **Rule:** Category domain files should be removed from recipe-service
   - **File:** `services/recipe-service/internal/ingredient/category/*.go` (11 files)
   - **Issue:** Files are orphaned (not wired up) but still tracked by git. Routes and migration are disconnected, but the package still compiles and its tests still run.
   - **Severity:** medium
   - **Fix:** Delete the `services/recipe-service/internal/ingredient/category/` directory and remove it from git tracking.

## Build & Test Results

| Service | Build | Tests | Notes |
|---------|-------|-------|-------|
| category-service | PASS | PASS | 1 test package (builder), 2 packages with no test files |
| shopping-service | PASS | PASS | 2 test packages (item builder, list builder), 3 packages with no test files |
| recipe-service | PASS | PASS | All 9 test packages pass including orphaned category tests |
| frontend | PASS | PASS | 43 test files, 399 tests all pass. Build produces 1MB bundle (chunk warning) |

## Overall Assessment

- **Plan Adherence:** MOSTLY_COMPLETE
- **Guidelines Compliance:** MINOR_VIOLATIONS
- **Recommendation:** NEEDS_FIXES

## Action Items

1. **Delete orphaned category domain** — Remove `services/recipe-service/internal/ingredient/category/` directory (11 files) from git. These files are disconnected from all entry points and serve no purpose.
2. **Add processor/handler tests** — At minimum, add unit tests for the shopping-service import handler logic (`list/rest.go:439-531`) which contains complex ingredient consolidation and category enrichment. Category-service processor tests (especially default seeding logic) would also add value.
3. **Add Zod schema for shopping list name** — Create `frontend/src/lib/schemas/shopping.schema.ts` with a name validation schema, consistent with the existing `categoryNameSchema` pattern.
4. **Create service documentation** — Add `domain.md`, `rest.md`, and `storage.md` for both category-service and shopping-service per task 6.2.
